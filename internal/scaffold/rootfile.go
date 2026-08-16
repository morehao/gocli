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

// GenerateRootWorkspace 生成根 go.work，指向 backend 下的各子模块（而非 ./backend，因为
// backend/ 本身是含 go.work 的嵌套 workspace、无 go.mod，Go 不允许 go.work use 一个 workspace 目录）。
// 即时可用的模块是 ./backend/apps/demo 与 ./backend/pkg；后续 create app 会向本根 go.work 追加 ./backend/apps/<X>。
func GenerateRootWorkspace(absRoot string) error {
	content := "go 1.26.1\n\nuse (\n\t./backend/apps/demo\n\t./backend/pkg\n)\n"
	return os.WriteFile(filepath.Join(absRoot, "go.work"), []byte(content), 0o644)
}

// RemoveNestedWorkspaceFiles 删除 backend/ 内嵌套的 go.work 与 go.work.sum，
// 使根 go.work 成为唯一 workspace（避免同一模块被两个 workspace 覆盖）。
// 模板资产保留 backend/go.work.tmpl（忠实 go-ark-template），生成时由本函数移除还原产物。
func RemoveNestedWorkspaceFiles(backendDir string) error {
	for _, f := range []string{"go.work", "go.work.sum"} {
		if err := RemoveDirIfExists(filepath.Join(backendDir, f)); err != nil {
			return err
		}
	}
	return nil
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
