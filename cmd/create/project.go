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

// createProject 基于内置 monorepo 模板创建新项目。
func createProject(dir, modulePath string, gitInit, tidy bool) error {
	if err := scaffold.ValidateModulePath(modulePath); err != nil {
		return err
	}
	if dir == "" {
		dir = filepath.Base(modulePath)
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

	tplFS, err := fs.Sub(template.MonorepoFS, "monorepo")
	if err != nil {
		return fmt.Errorf("load built-in monorepo template fail: %w", err)
	}

	// 复制模板并重写 Go import 路径（github.com/example/... -> <modulePath>/...）
	prefixMappings := map[string]string{templateBaseModule: modulePath}
	if err := scaffold.CopyTreeFS(tplFS, ".", absDir, func(relPath string, content []byte) ([]byte, error) {
		if strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, ".go.tmpl") {
			return scaffold.RewriteGoContent(relPath, content, "", "", prefixMappings)
		}
		return content, nil
	}); err != nil {
		return fmt.Errorf("copy template fail: %w", err)
	}

	// 恢复模板化的 go.mod/go.sum/go.work 文件名
	if err := scaffold.RestoreTemplateFiles(absDir); err != nil {
		return fmt.Errorf("restore template module files fail: %w", err)
	}

	// 重写所有 go.mod 的 module 路径（github.com/example/pkg -> <modulePath>/pkg 等）
	if err := scaffold.RewriteGoModsInTree(absDir, templateBaseModule, modulePath); err != nil {
		return fmt.Errorf("rewrite go.mod module paths fail: %w", err)
	}

	// 防御性移除 .git
	if err := scaffold.RemoveDirIfExists(filepath.Join(absDir, ".git")); err != nil {
		return fmt.Errorf("remove .git fail: %w", err)
	}

	if gitInit {
		if err := scaffold.RunGitInit(absDir); err != nil {
			fmt.Printf("Warning: git init failed: %v\n", err)
		}
	}
	if tidy {
		if err := scaffold.RunGoWorkSync(absDir); err != nil {
			fmt.Printf("Warning: go work sync failed (run it manually in %s): %v\n", absDir, err)
		}
	}

	fmt.Printf("Successfully created monorepo project at %s\n", absDir)
	fmt.Printf("  cd %s\n", absDir)
	fmt.Printf("  gocli create app -n <app-name>   # add a new app\n")
	fmt.Printf("  gocli generate module -a demoapp # generate code for the sample app\n")
	return nil
}
