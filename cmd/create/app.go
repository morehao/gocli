package create

import (
	"bytes"
	"fmt"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/morehao/gocli/internal/scaffold"
	"github.com/morehao/gocli/template"
)

// templateAppModule 模板中示例 app 的占位模块路径。
const templateAppModule = templateBaseModule + "/demoapp"

// createApp 在既有 monorepo 中基于内置模板新增一个 app。
func createApp(appName, moduleOverride string, tidy bool) error {
	if err := scaffold.ValidateAppName(appName); err != nil {
		return err
	}

	rootDir, goWorkPath, err := findMonorepoRoot()
	if err != nil {
		return err
	}
	appsDir := filepath.Join(rootDir, "apps")
	if _, err := os.Stat(appsDir); err != nil {
		return fmt.Errorf("%s is not a monorepo: apps directory does not exist", rootDir)
	}
	newAppDir := filepath.Join(appsDir, appName)
	if _, err := os.Stat(newAppDir); err == nil {
		return fmt.Errorf("app already exists: %s", newAppDir)
	}

	// 推断新 app 的模块路径
	var newAppModule string
	if moduleOverride != "" {
		newAppModule = moduleOverride
	} else {
		base, err := scaffold.InferBaseModule(rootDir)
		if err != nil {
			return fmt.Errorf("infer base module fail (use --module to override): %w", err)
		}
		newAppModule = base + "/" + appName
	}
	if err := scaffold.ValidateModulePath(newAppModule); err != nil {
		return err
	}

	// 推断 pkg 模块路径（模板 import 的 github.com/example/pkg 需指向本仓库的公共库）
	pkgModule, err := resolvePkgModulePath(rootDir, newAppModule)
	if err != nil {
		return err
	}

	tplAppFS, err := fs.Sub(template.MonorepoFS, filepath.ToSlash(filepath.Join("monorepo", "apps", "demoapp")))
	if err != nil {
		return fmt.Errorf("load built-in app template fail: %w", err)
	}

	prefixMappings := map[string]string{
		templateAppModule:           newAppModule,
		templateBaseModule + "/pkg": pkgModule,
	}
	if err := scaffold.CopyTreeFS(tplAppFS, ".", newAppDir, func(relPath string, content []byte) ([]byte, error) {
		if strings.HasSuffix(relPath, ".go") || strings.HasSuffix(relPath, ".go.tmpl") {
			content, err = scaffold.RewriteGoContent(relPath, content, "demoapp", appName, prefixMappings)
			if err != nil {
				return nil, err
			}
			// 注释与字符串字面量中的 app 名（如 Swagger 路由、路由组名）做文本替换
			return bytes.ReplaceAll(content, []byte("demoapp"), []byte(appName)), nil
		}
		return content, nil
	}); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return fmt.Errorf("copy app template fail: %w", err)
	}

	// 恢复模板化的 go.mod/go.sum 文件名
	if err := scaffold.RestoreTemplateFiles(newAppDir); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return fmt.Errorf("restore app template module files fail: %w", err)
	}

	// 重写新 app 的 go.mod 模块路径
	appGoMod := filepath.Join(newAppDir, "go.mod")
	if _, err := os.Stat(appGoMod); err == nil {
		if err := scaffold.RewriteModuleStmt(appGoMod, newAppModule); err != nil {
			scaffold.RemoveDirIfExists(newAppDir)
			return fmt.Errorf("rewrite app go.mod fail: %w", err)
		}
	}

	// 替换 yaml 等配置文本中的 app 名
	if err := scaffold.ReplaceTextInTree(newAppDir, "demoapp", appName, ".yaml", ".yml"); err != nil {
		scaffold.RemoveDirIfExists(newAppDir)
		return fmt.Errorf("replace app name in configs fail: %w", err)
	}

	// 注册进 go.work
	if goWorkPath != "" {
		if err := scaffold.AddGoWorkUse(goWorkPath, filepath.ToSlash(filepath.Join("apps", appName))); err != nil {
			scaffold.RemoveDirIfExists(newAppDir)
			return fmt.Errorf("add app to go.work fail: %w", err)
		}
	}

	if tidy {
		if err := scaffold.RunGoModTidy(newAppDir); err != nil {
			fmt.Printf("Warning: go mod tidy failed (run it manually in %s): %v\n", newAppDir, err)
		}
	}

	fmt.Printf("Successfully created app %s at %s (module: %s)\n", appName, newAppDir, newAppModule)
	fmt.Printf("  gocli generate module -a %s   # generate code for the new app\n", appName)
	return nil
}

// resolvePkgModulePath 返回 monorepo 内公共库 pkg 的模块路径。
// 优先读取 <root>/pkg/go.mod；不存在时按 <base>/pkg 约定推断。
func resolvePkgModulePath(rootDir, newAppModule string) (string, error) {
	pkgGoMod := filepath.Join(rootDir, "pkg", "go.mod")
	if _, err := os.Stat(pkgGoMod); err == nil {
		modulePath, err := scaffold.ReadModulePath(pkgGoMod)
		if err != nil {
			return "", fmt.Errorf("read pkg/go.mod fail: %w", err)
		}
		return modulePath, nil
	}
	base := scaffold.BaseModulePath(newAppModule)
	if base == "" {
		return "", fmt.Errorf("cannot infer pkg module path from %s", newAppModule)
	}
	return base + "/pkg", nil
}

// findMonorepoRoot 从当前目录向上查找 monorepo 根（含 go.work 或 go.mod 的最近目录）。
func findMonorepoRoot() (rootDir, goWorkPath string, err error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", "", fmt.Errorf("get current directory fail: %w", err)
	}
	for {
		goWork := filepath.Join(dir, "go.work")
		if _, err := os.Stat(goWork); err == nil {
			return dir, goWork, nil
		}
		goMod := filepath.Join(dir, "go.mod")
		if _, err := os.Stat(goMod); err == nil {
			return dir, "", nil
		}
		parent := filepath.Dir(dir)
		if parent == dir {
			break
		}
		dir = parent
	}
	return "", "", fmt.Errorf("current directory is not inside a Go project: no go.mod or go.work found")
}
