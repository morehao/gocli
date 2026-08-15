package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// AddGoWorkUse 向 go.work 追加 use 目录（已存在则跳过）。
// 相对路径会规范化为 ./xxx 形式，与 go.work 惯例一致。
func AddGoWorkUse(goWorkPath, dir string) error {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return fmt.Errorf("read go.work fail: %w", err)
	}
	workFile, err := modfile.ParseWork(goWorkPath, content, nil)
	if err != nil {
		return fmt.Errorf("parse go.work fail: %w", err)
	}

	cleanDir := filepath.Clean(dir)
	usePath := filepath.ToSlash(cleanDir)
	if usePath != "." && !strings.HasPrefix(usePath, "./") && !strings.HasPrefix(usePath, "/") {
		usePath = "./" + usePath
	}
	for _, use := range workFile.Use {
		if filepath.Clean(use.Path) == cleanDir {
			return nil
		}
	}

	if err := workFile.AddUse(usePath, ""); err != nil {
		return fmt.Errorf("add use %s fail: %w", dir, err)
	}
	// WorkFile 无 Format 方法，通过包级 Format 序列化其语法树
	return os.WriteFile(goWorkPath, modfile.Format(workFile.Syntax), 0o644)
}

// GoWorkUseDirs 返回 go.work 中 use 的目录（按 go.work 中书写形式，未做绝对化）。
func GoWorkUseDirs(goWorkPath string) ([]string, error) {
	content, err := os.ReadFile(goWorkPath)
	if err != nil {
		return nil, err
	}
	workFile, err := modfile.ParseWork(goWorkPath, content, nil)
	if err != nil {
		return nil, err
	}
	dirs := make([]string, 0, len(workFile.Use))
	for _, use := range workFile.Use {
		dirs = append(dirs, filepath.ToSlash(use.Path))
	}
	return dirs, nil
}

// HasGoWorkUse 判断 go.work 是否已包含指定 use 目录。
func HasGoWorkUse(goWorkPath, dir string) (bool, error) {
	dirs, err := GoWorkUseDirs(goWorkPath)
	if err != nil {
		return false, err
	}
	cleanDir := filepath.Clean(dir)
	for _, d := range dirs {
		if filepath.Clean(d) == cleanDir {
			return true, nil
		}
	}
	return false, nil
}

// IsGoWork 判断路径是否包含 go.work 文件。
func IsGoWork(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.work"))
	return err == nil
}

// IsGoProject 判断路径是否为 Go 项目（含 go.mod 或 go.work）。
func IsGoProject(path string) bool {
	if _, err := os.Stat(filepath.Join(path, "go.mod")); err == nil {
		return true
	}
	return IsGoWork(path)
}

// ReplaceTextInTree 遍历 rootDir，对扩展名匹配的文本文件执行 old -> new 全量替换。
// 用于替换 yaml 等配置中的 app 名称。
func ReplaceTextInTree(rootDir string, old, new string, exts ...string) error {
	extSet := make(map[string]struct{}, len(exts))
	for _, ext := range exts {
		extSet[strings.ToLower(ext)] = struct{}{}
	}
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		ext := strings.ToLower(filepath.Ext(path))
		if _, ok := extSet[ext]; !ok {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if !strings.Contains(string(content), old) {
			return nil
		}
		newContent := strings.ReplaceAll(string(content), old, new)
		return os.WriteFile(path, []byte(newContent), 0o644)
	})
}
