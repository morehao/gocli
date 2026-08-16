// Package scaffold 提供基于内置模板快速创建 Go 项目/应用所需的通用能力：
// 目录复制、模块路径重写、go.work 编辑、命名校验等。
package scaffold

import (
	"fmt"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
)

// defaultIgnoreDirMap 默认忽略的目录（按路径片段匹配）。
var defaultIgnoreDirMap = map[string]struct{}{
	".git":         {},
	".idea":        {},
	".vscode":      {},
	".history":     {},
	"node_modules": {},
	"vendor":       {},
	"log":          {},
	"logs":         {},
	"tmp":          {},
	"temp":         {},
}

// defaultIgnoreFiles 默认忽略的文件（支持 *.ext 模式）。
var defaultIgnoreFiles = []string{
	".DS_Store",
	"*.log",
	"*.tmp",
}

// ShouldIgnore 判断相对路径（相对被复制目录的根）是否应被忽略。
func ShouldIgnore(relativePath string) bool {
	normalizedPath := filepath.ToSlash(relativePath)
	parts := strings.Split(normalizedPath, "/")

	for _, part := range parts {
		if _, ok := defaultIgnoreDirMap[part]; ok {
			return true
		}
	}

	fileName := parts[len(parts)-1]
	for _, pattern := range defaultIgnoreFiles {
		if strings.HasPrefix(pattern, "*.") {
			ext := pattern[1:] // ".log"
			if strings.HasSuffix(fileName, ext) {
				return true
			}
		} else if fileName == pattern {
			return true
		}
	}

	return false
}

// CopyFile 复制单个文件。
func CopyFile(src, dst string) error {
	sourceFileStat, err := os.Stat(src)
	if err != nil {
		return err
	}
	if !sourceFileStat.Mode().IsRegular() {
		return fmt.Errorf("%s is not a regular file", src)
	}

	source, err := os.Open(src)
	if err != nil {
		return err
	}
	defer source.Close()

	if err := os.MkdirAll(filepath.Dir(dst), 0o755); err != nil {
		return err
	}

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// CopyTreeFS 将 fs 中以 prefix 为根的子树复制到 dstDir（相对路径按 ShouldIgnore 过滤）。
// onFile 在写入前可改写文件内容；onFile 返回 nil 时跳过该文件。
func CopyTreeFS(src fs.FS, prefix, dstDir string, onFile func(relPath string, content []byte) ([]byte, error)) error {
	return fs.WalkDir(src, prefix, func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		relPath, err := filepath.Rel(prefix, path)
		if err != nil {
			return err
		}
		if relPath == "." {
			return nil
		}
		if ShouldIgnore(filepath.ToSlash(relPath)) {
			if d.IsDir() {
				return fs.SkipDir
			}
			return nil
		}

		target := filepath.Join(dstDir, relPath)
		if d.IsDir() {
			return os.MkdirAll(target, 0o755)
		}

		content, err := fs.ReadFile(src, path)
		if err != nil {
			return err
		}
		if onFile != nil {
			content, err = onFile(relPath, content)
			if err != nil {
				return err
			}
			if content == nil {
				return nil
			}
		}
		if err := os.MkdirAll(filepath.Dir(target), 0o755); err != nil {
			return err
		}
		return os.WriteFile(target, content, 0o644)
	})
}

// RemoveDirIfExists 删除目录（不存在时静默成功）。
func RemoveDirIfExists(dir string) error {
	if _, err := os.Stat(dir); err != nil {
		if os.IsNotExist(err) {
			return nil
		}
		return err
	}
	return os.RemoveAll(dir)
}

// RestoreTemplateFiles 将模板中模板化的文件恢复为标准文件名：
//   - go.mod.tmpl -> go.mod、go.sum.tmpl -> go.sum、go.work.tmpl -> go.work、go.work.sum.tmpl -> go.work.sum
//   - *.go.tmpl    -> *.go
//
// 模板使用 .tmpl 后缀存放这些文件，原因有二：
//  1. go:embed 会拒绝嵌入含 go.mod 的嵌套模块目录
//     （"cannot embed directory ... in different module"）；
//  2. 模板目录内的 .go 文件若保持 .go 后缀，会被父模块的 go build ./... 当作源码编译。
func RestoreTemplateFiles(rootDir string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		dir := filepath.Dir(path)
		base := filepath.Base(path)

		var newBase string
		switch base {
		case "go.mod.tmpl":
			newBase = "go.mod"
		case "go.sum.tmpl":
			newBase = "go.sum"
		case "go.work.tmpl":
			newBase = "go.work"
		case "go.work.sum.tmpl":
			newBase = "go.work.sum"
		default:
			if strings.HasSuffix(base, ".go.tmpl") {
				newBase = strings.TrimSuffix(base, ".tmpl")
			}
		}
		if newBase == "" {
			return nil
		}

		newPath := filepath.Join(dir, newBase)
		if _, err := os.Stat(newPath); err == nil {
			return fmt.Errorf("restore template file fail: %s already exists", newPath)
		}
		return os.Rename(path, newPath)
	})
}

// IsDirEmpty 判断目录是否为空（目录不存在时返回 false, nil）。
func IsDirEmpty(dir string) (bool, error) {
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return len(entries) == 0, nil
}
