package codegen

import (
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strings"

	golibcodegen "github.com/morehao/golib/codegen"
)

// manifestFileName 增量指纹清单文件名（位于 ProjectRoot 根目录）。
const manifestFileName = "codegen.manifest.json"

// manifest 记录每张表最近一次生成时的输入指纹，用于增量跳过未变更的表。
type manifest struct {
	Version int                      `json:"version"`
	Tables  map[string]manifestEntry `json:"tables"`
}

type manifestEntry struct {
	Fingerprint string   `json:"fingerprint"`
	Modes       []string `json:"modes"` // 已生成过的 mode
	UpdatedAt   string   `json:"updatedAt"`
}

func manifestPath(projectRoot string) string {
	return filepath.Join(projectRoot, manifestFileName)
}

func loadManifest(projectRoot string) *manifest {
	m := &manifest{Version: 1, Tables: map[string]manifestEntry{}}
	data, err := os.ReadFile(manifestPath(projectRoot))
	if err != nil {
		return m
	}
	if err := json.Unmarshal(data, m); err != nil {
		return &manifest{Version: 1, Tables: map[string]manifestEntry{}}
	}
	if m.Tables == nil {
		m.Tables = map[string]manifestEntry{}
	}
	return m
}

func saveManifest(projectRoot string, m *manifest) error {
	data, err := json.MarshalIndent(m, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(manifestPath(projectRoot), data, 0o644)
}

// tableFingerprint 依据表结构元数据算出稳定指纹（列名/类型/可空/主键 + 结构名 + 包名）。
func tableFingerprint(res *golibcodegen.ModuleTplAnalysisRes, packageName string) string {
	if res == nil {
		return ""
	}
	var fields []golibcodegen.ModelField
	for _, v := range res.TplAnalysisList {
		if v.OriginLayerName == golibcodegen.LayerNameModel {
			fields = v.ModelFields
			break
		}
	}
	if fields == nil && len(res.TplAnalysisList) > 0 {
		fields = res.TplAnalysisList[0].ModelFields
	}
	lines := make([]string, 0, len(fields)+2)
	lines = append(lines, "struct:"+res.StructName, "pkg:"+packageName)
	for _, f := range fields {
		lines = append(lines, fmt.Sprintf("%s:%s:%t:%t", f.ColumnName, f.FieldType, f.IsNullable, f.IsPrimaryKey))
	}
	sort.Strings(lines)
	sum := sha1.Sum([]byte(strings.Join(lines, "\n")))
	return hex.EncodeToString(sum[:])
}

// skipIfUnchanged 若该表在某 mode 下指纹未变更则返回 true（提示跳过）。
// force / dryRun 不跳过（dry-run 需要展示将写入的内容）。
func (g *Generator) skipIfUnchanged(res *golibcodegen.ModuleTplAnalysisRes, mode Mode) (bool, string) {
	if g.opts.Force || g.opts.DryRun {
		return false, ""
	}
	table := res.TableName
	fp := tableFingerprint(res, g.cfg.Module.PackageName)
	m := loadManifest(g.opts.ProjectRoot)
	entry, ok := m.Tables[table]
	if !ok || entry.Fingerprint != fp {
		return false, ""
	}
	for _, mm := range entry.Modes {
		if mm == string(mode) {
			return true, fmt.Sprintf("表 %s 结构未变更（mode=%s），跳过生成（--force 强制重生成）", table, mode)
		}
	}
	return false, ""
}

// recordManifest 记录本次生成后的指纹与 mode。
func (g *Generator) recordManifest(res *golibcodegen.ModuleTplAnalysisRes, modes ...Mode) {
	if g.opts.DryRun {
		return
	}
	table := res.TableName
	fp := tableFingerprint(res, g.cfg.Module.PackageName)
	m := loadManifest(g.opts.ProjectRoot)
	entry := m.Tables[table]
	entry.Fingerprint = fp
	seen := map[string]bool{}
	for _, x := range entry.Modes {
		seen[x] = true
	}
	for _, mo := range modes {
		if !seen[string(mo)] {
			entry.Modes = append(entry.Modes, string(mo))
			seen[string(mo)] = true
		}
	}
	entry.UpdatedAt = nowStamp()
	m.Tables[table] = entry
	_ = saveManifest(g.opts.ProjectRoot, m)
}

func nowStamp() string {
	return timeNow().Format("2006-01-02T15:04:05Z07:00")
}
