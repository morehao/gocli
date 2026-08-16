package create

import (
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/morehao/gocli/internal/scaffold"
	"github.com/morehao/gocli/template"
)

// templateBaseModule 模板中使用的占位基础模块路径。
const templateBaseModule = "github.com/example"

// CreateOptions 描述一次 create project 的全部输入。
type CreateOptions struct {
	Dir         string // 目标仓库根目录；空则取 ModulePath 末段
	ModulePath  string // 必填模块路径，如 github.com/acme/my-ark
	ProjectName string // 项目名；空则由 defaultProjectName 推导
	GitInit     bool
	Tidy        bool
}

// defaultProjectName 推导项目名：显式值 > Dir 的 basename > ModulePath 末段。
func defaultProjectName(dir, modulePath, projectName string) string {
	if projectName != "" {
		return projectName
	}
	if dir != "" {
		base := filepath.Base(dir)
		if base != "" && base != "." && base != string(filepath.Separator) {
			return base
		}
	}
	if modulePath != "" {
		if idx := strings.LastIndex(modulePath, "/"); idx >= 0 {
			return modulePath[idx+1:]
		}
		return modulePath
	}
	return "app"
}

// createProject 保留旧签名作为薄封装。
func createProject(dir, modulePath string, gitInit, tidy bool) error {
	return createProjectWithOpts(CreateOptions{Dir: dir, ModulePath: modulePath, GitInit: gitInit, Tidy: tidy})
}

func createProjectWithOpts(o CreateOptions) error {
	if err := scaffold.ValidateModulePath(o.ModulePath); err != nil {
		return err
	}
	dir := o.Dir
	if dir == "" {
		dir = filepath.Base(o.ModulePath)
	}
	absDir, err := filepath.Abs(dir)
	if err != nil {
		return fmt.Errorf("resolve destination directory fail: %w", err)
	}
	if _, err := os.Stat(absDir); err == nil {
		empty, err := scaffold.IsDirEmpty(absDir)
		if err != nil {
			return err
		}
		if !empty {
			return fmt.Errorf("destination directory is not empty: %s", absDir)
		}
	}
	if err := os.MkdirAll(absDir, 0o755); err != nil {
		return fmt.Errorf("create destination directory fail: %w", err)
	}

	_ = defaultProjectName(dir, o.ModulePath, o.ProjectName) // 项目名用于文档/命名；module 以用户输入为准

	tplFS, err := fs.Sub(template.ArkFS, "ark")
	if err != nil {
		return fmt.Errorf("load ark template fail: %w", err)
	}
	prefixMappings := map[string]string{scaffold.ArkTemplateModule: o.ModulePath}
	if err := scaffold.CopyTreeFS(tplFS, ".", filepath.Join(absDir, "backend"), func(relPath string, content []byte) ([]byte, error) {
		if strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, ".go.tmpl") {
			return scaffold.RewriteGoContent(relPath, content, "", "", prefixMappings)
		}
		// 非 Go 文本（如 README 中的示例 import）若含模块占位前缀，做纯文本替换以免残留占位符。
		if strings.Contains(string(content), scaffold.ArkTemplateModule) {
			return []byte(strings.ReplaceAll(string(content), scaffold.ArkTemplateModule, o.ModulePath)), nil
		}
		return content, nil
	}); err != nil {
		return fmt.Errorf("copy ark template fail: %w", err)
	}

	if err := scaffold.RestoreTemplateFiles(filepath.Join(absDir, "backend")); err != nil {
		return fmt.Errorf("restore backend template files fail: %w", err)
	}
	if err := scaffold.RewriteGoModsInTree(filepath.Join(absDir, "backend"), scaffold.ArkTemplateModule, o.ModulePath); err != nil {
		return fmt.Errorf("rewrite backend go.mod paths fail: %w", err)
	}

	if err := scaffold.GenerateRootModule(absDir, o.ModulePath); err != nil {
		return err
	}
	if err := scaffold.GenerateRootWorkspace(absDir); err != nil {
		return err
	}
	if err := scaffold.ScaffoldPlaceholderDirs(absDir, "frontend", "docs"); err != nil {
		return err
	}

	if err := scaffold.RemoveDirIfExists(filepath.Join(absDir, ".git")); err != nil {
		return fmt.Errorf("remove .git fail: %w", err)
	}
	if o.GitInit {
		if err := scaffold.RunGitInit(absDir); err != nil {
			fmt.Printf("Warning: git init failed: %v\n", err)
		}
	}
	if o.Tidy {
		if err := scaffold.RunGoWorkSync(absDir); err != nil {
			fmt.Printf("Warning: go work sync failed (run it manually in %s): %v\n", absDir, err)
		}
	}
	fmt.Printf("Successfully created monorepo project at %s\n", absDir)
	fmt.Printf("  cd %s\n", absDir)
	return nil
}
