package cutter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

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

// shouldIgnore 检查路径是否应该被忽略
var defaultIgnoreFiles = []string{
	".DS_Store",
	"*.log",
	"*.tmp",
}

// isGoProject 检查指定路径是否为Go项目（是否包含go.mod文件）
func isGoProject(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return !os.IsNotExist(err)
}

func shouldIgnore(relativePath string) bool {
	// 将路径里的 系统特定的路径分隔符 转成 统一的 / 形式
	normalizedPath := filepath.ToSlash(relativePath)
	parts := strings.Split(normalizedPath, "/")

	// 忽略目录
	for _, part := range parts {
		if _, ok := defaultIgnoreDirMap[part]; ok {
			return true
		}
	}

	// 忽略文件
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

// copyFile 复制文件
func copyFile(src, dst string) error {
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

	destination, err := os.Create(dst)
	if err != nil {
		return err
	}
	defer destination.Close()
	_, err = io.Copy(destination, source)
	return err
}

// isGoWork 检查指定路径是否为 Go workspace（是否包含 go.work 文件）
func isGoWork(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.work"))
	return !os.IsNotExist(err)
}

// readModulePath 读取 go.mod 的 module 路径
func readModulePath(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("read go.mod fail: %w", err)
	}
	modFile, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return "", fmt.Errorf("parse go.mod fail: %w", err)
	}
	return modFile.Module.Mod.Path, nil
}

// findProjectRoot 查找项目根目录
// 1. 如果存在 go.work，使用 workspace 模式
// 2. 否则向上遍历找 go.mod
// 返回: rootDir, isWorkspace, modulePath, error
func findProjectRoot(currentDir string) (string, bool, string, error) {
	dir := currentDir
	for {
		goWorkPath := filepath.Join(dir, "go.work")
		if _, err := os.Stat(goWorkPath); err == nil {
			goModPath := filepath.Join(dir, "go.mod")
			if _, err := os.Stat(goModPath); err == nil {
				modulePath, err := readModulePath(goModPath)
				if err != nil {
					return "", false, "", err
				}
				return dir, true, modulePath, nil
			}
			return "", false, "", fmt.Errorf("go.work found but no go.mod in same directory: %s", dir)
		}

		goModPath := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goModPath); err == nil {
			modulePath, err := readModulePath(goModPath)
			if err != nil {
				return "", false, "", err
			}
			return dir, false, modulePath, nil
		}

		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}

	return "", false, "", fmt.Errorf("not a Go project (no go.mod or go.work found)")
}
