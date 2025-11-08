package cutter

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// 默认忽略的目录和文件模式
var defaultIgnores = []string{
	".git",
	"node_modules",
	"vendor",
	"log",
	"logs",
	"tmp",
	"temp",
	".idea",
	".vscode",
	".DS_Store",
	"*.log",
	"*.tmp",
	".env.local",
	".env.*.local",
	"coverage",
	"dist",
	"build",
}

// isGoProject 检查指定路径是否为Go项目（是否包含go.mod文件）
func isGoProject(path string) bool {
	_, err := os.Stat(filepath.Join(path, "go.mod"))
	return !os.IsNotExist(err)
}

// shouldIgnore 检查路径是否应该被忽略
func shouldIgnore(relativePath string) bool {
	// 分割路径为各个部分
	parts := strings.Split(filepath.ToSlash(relativePath), "/")

	for _, part := range parts {
		for _, ignorePattern := range defaultIgnores {
			// 精确匹配
			if part == ignorePattern {
				return true
			}
			// 通配符匹配（简单实现，支持 *.ext 格式）
			if strings.HasPrefix(ignorePattern, "*.") {
				ext := ignorePattern[1:]
				if strings.HasSuffix(part, ext) {
					return true
				}
			}
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
