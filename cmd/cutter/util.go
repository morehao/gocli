package cutter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
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
