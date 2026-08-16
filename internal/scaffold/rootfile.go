package scaffold

import (
	"fmt"
	"os"
	"path/filepath"

	"golang.org/x/mod/modfile"
)

// ArkTemplateModule 是 template/ark 中使用的模块前缀占位符。
const ArkTemplateModule = "github.com/morehao/go-ark-template"

// GenerateRootModule 在 absRoot 生成根 go.mod，module 为 modulePath。
func GenerateRootModule(absRoot, modulePath string) error {
	content, err := modfile.Parse("go.mod", []byte("module "+modulePath+"\n\ngo 1.26.1\n"), nil)
	if err != nil {
		return fmt.Errorf("parse root go.mod fail: %w", err)
	}
	formatted, err := content.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(filepath.Join(absRoot, "go.mod"), formatted, 0o644)
}

// GenerateRootWorkspace 生成根 go.work，use ./backend。
func GenerateRootWorkspace(absRoot string) error {
	content := "go 1.26.1\n\nuse (\n\t./backend\n)\n"
	return os.WriteFile(filepath.Join(absRoot, "go.work"), []byte(content), 0o644)
}

// ScaffoldPlaceholderDirs 创建纯占位目录并写入 .gitkeep。
func ScaffoldPlaceholderDirs(absRoot string, names ...string) error {
	for _, n := range names {
		d := filepath.Join(absRoot, n)
		if err := os.MkdirAll(d, 0o755); err != nil {
			return err
		}
		if err := os.WriteFile(filepath.Join(d, ".gitkeep"), []byte(""), 0o644); err != nil {
			return err
		}
	}
	return nil
}
