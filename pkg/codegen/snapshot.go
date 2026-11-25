package codegen

import (
	"io"
	"os"
	"path/filepath"
)

// backupDir 存放生成前快照（每个文件一份 .bak），供 Undo 还原。
const backupDir = ".codegen/backup"

// snapshotFiles 需要快照的「聚合型、会被追加/插入修改」的文件（相对 ProjectRoot）。
var snapshotFiles = []string{
	"packages/types/src/index.ts",
	"packages/api/src/index.ts",
	"apps/web/src/App.tsx",
	"apps/api/internal/router/router.go",
	"apps/api/model/table.go",
	"pkg/code/code.go",
}

// snapshot 在生成前把聚合型目标文件复制到 .codegen/backup，供 Undo 还原。dryRun 不快照。
var snapshotTaken = map[string]bool{}

func (g *Generator) snapshot() {
	if g.opts.DryRun || g.opts.NoSnapshot {
		return
	}
	// 同一进程内对同一工程只快照一次（批量多表时保留「批量前」的干净快照）
	if snapshotTaken[g.opts.ProjectRoot] {
		return
	}
	snapshotTaken[g.opts.ProjectRoot] = true
	dir := filepath.Join(g.opts.ProjectRoot, backupDir)
	_ = os.MkdirAll(dir, 0o755)
	for _, rel := range snapshotFiles {
		src := filepath.Join(g.opts.ProjectRoot, rel)
		data, err := os.ReadFile(src)
		if err != nil {
			continue // 文件不存在则跳过
		}
		dst := filepath.Join(dir, backupName(rel))
		_ = os.WriteFile(dst, data, 0o644)
	}
}

// Undo 用快照还原被修改的聚合文件（types/api/App.tsx），并清理生成的 manifest 中相关记录不变。
// 返回被还原的文件列表；无快照时返回空。
func (g *Generator) Undo() ([]string, error) {
	dir := filepath.Join(g.opts.ProjectRoot, backupDir)
	var restored []string
	for _, rel := range snapshotFiles {
		dst := filepath.Join(g.opts.ProjectRoot, rel)
		src := filepath.Join(dir, backupName(rel))
		if _, err := os.Stat(src); err != nil {
			continue
		}
		if err := copyFile(src, dst); err != nil {
			return restored, err
		}
		restored = append(restored, dst)
	}
	return restored, nil
}

func backupName(rel string) string {
	// 把路径分隔符替换为 __，避免子目录
	b := make([]byte, 0, len(rel))
	for i := 0; i < len(rel); i++ {
		if rel[i] == '/' {
			b = append(b, '_', '_')
		} else {
			b = append(b, rel[i])
		}
	}
	return string(b) + ".bak"
}

func copyFile(src, dst string) error {
	in, err := os.Open(src)
	if err != nil {
		return err
	}
	defer in.Close()
	out, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer out.Close()
	_, err = io.Copy(out, in)
	return err
}
