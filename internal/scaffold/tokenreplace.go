package scaffold

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
)

// ReplaceTokensInTree 遍历 rootDir 下文本文件，按 repl 做替换（key 长度降序，避免短 token 吞长 token）。
// exts 为空时替换所有文件；忽略目录/二进制文件跳过。用于复制 demo 后把 app 名 token 换成新名。
func ReplaceTokensInTree(rootDir string, repl map[string]string, exts ...string) error {
	extSet := make(map[string]struct{}, len(exts))
	for _, e := range exts {
		extSet[strings.ToLower(e)] = struct{}{}
	}
	keys := make([]string, 0, len(repl))
	for k := range repl {
		keys = append(keys, k)
	}
	sort.Slice(keys, func(i, j int) bool { return len(keys[i]) > len(keys[j]) })

	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			if SkipIgnoredDir(path) {
				return filepath.SkipDir
			}
			return nil
		}
		base := info.Name()
		if base == ".git" || base == "go.sum" || base == "go.work.sum" {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(base))
		if len(extSet) > 0 {
			if _, ok := extSet[ext]; !ok {
				return nil
			}
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		text := string(content)
		changed := false
		for _, k := range keys {
			if !strings.Contains(text, k) {
				continue
			}
			text = strings.ReplaceAll(text, k, repl[k])
			changed = true
		}
		if !changed {
			return nil
		}
		return os.WriteFile(path, []byte(text), 0o644)
	})
}

// SkipIgnoredDir 返回 true 表示应跳过该（相对会被忽略的）目录，复用 defaultIgnoreDirMap。
// path 为绝对路径；对非目录返回 false。
func SkipIgnoredDir(path string) bool {
	fi, err := os.Stat(path)
	if err != nil || !fi.IsDir() {
		return false
	}
	name := filepath.Base(path)
	_, ok := defaultIgnoreDirMap[name]
	return ok
}
