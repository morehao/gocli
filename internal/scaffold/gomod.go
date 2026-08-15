package scaffold

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"golang.org/x/mod/modfile"
)

// ReadModulePath 读取 go.mod 的 module 路径。
func ReadModulePath(goModPath string) (string, error) {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return "", fmt.Errorf("read go.mod fail: %w", err)
	}
	modFile, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return "", fmt.Errorf("parse go.mod fail: %w", err)
	}
	if modFile.Module == nil {
		return "", fmt.Errorf("go.mod has no module declaration: %s", goModPath)
	}
	return modFile.Module.Mod.Path, nil
}

// BaseModulePath 返回模块路径的父路径（去掉最后一个路径段），
// 例如 github.com/example/demoapp -> github.com/example。
func BaseModulePath(modulePath string) string {
	idx := strings.LastIndex(modulePath, "/")
	if idx == -1 {
		return ""
	}
	return modulePath[:idx]
}

// RewriteModuleStmt 重写 go.mod 的 module 语句为 newModulePath。
func RewriteModuleStmt(goModPath, newModulePath string) error {
	content, err := os.ReadFile(goModPath)
	if err != nil {
		return err
	}
	modFile, err := modfile.Parse(goModPath, content, nil)
	if err != nil {
		return err
	}
	if modFile.Module == nil {
		return fmt.Errorf("go.mod has no module declaration: %s", goModPath)
	}
	if modFile.Module.Mod.Path == newModulePath {
		return nil
	}
	if err := modFile.AddModuleStmt(newModulePath); err != nil {
		return err
	}
	formatted, err := modFile.Format()
	if err != nil {
		return err
	}
	return os.WriteFile(goModPath, formatted, 0o644)
}

// RewriteGoModsInTree 遍历 rootDir 下所有 go.mod，将模块路径中 oldPrefix 前缀替换为 newPrefix。
// 例如 oldPrefix=github.com/example, newPrefix=github.com/acme/backend 时：
// github.com/example/pkg -> github.com/acme/backend/pkg
func RewriteGoModsInTree(rootDir, oldPrefix, newPrefix string) error {
	return filepath.Walk(rootDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() || filepath.Base(path) != "go.mod" {
			return nil
		}
		modulePath, err := ReadModulePath(path)
		if err != nil {
			return err
		}
		if !HasModulePathPrefix(modulePath, oldPrefix) {
			return nil
		}
		newModulePath := newPrefix + strings.TrimPrefix(modulePath, oldPrefix)
		return RewriteModuleStmt(path, newModulePath)
	})
}

// InferBaseModule 从 monorepo 根目录推断基础模块路径，优先级：
//  1. 根目录 go.mod 的 module 路径
//  2. pkg/go.mod 的 module 路径的父路径（如 github.com/acme/backend/pkg -> github.com/acme/backend）
//  3. 任意 apps/*/go.mod 的 module 路径的父路径
func InferBaseModule(rootDir string) (string, error) {
	rootGoMod := filepath.Join(rootDir, "go.mod")
	if _, err := os.Stat(rootGoMod); err == nil {
		modulePath, err := ReadModulePath(rootGoMod)
		if err != nil {
			return "", err
		}
		return modulePath, nil
	}

	pkgGoMod := filepath.Join(rootDir, "pkg", "go.mod")
	if _, err := os.Stat(pkgGoMod); err == nil {
		modulePath, err := ReadModulePath(pkgGoMod)
		if err != nil {
			return "", err
		}
		if base := BaseModulePath(modulePath); base != "" {
			return base, nil
		}
	}

	appsDir := filepath.Join(rootDir, "apps")
	entries, err := os.ReadDir(appsDir)
	if err != nil {
		return "", fmt.Errorf("cannot infer base module: no root go.mod, pkg/go.mod or apps dir found: %w", err)
	}
	for _, entry := range entries {
		if !entry.IsDir() {
			continue
		}
		appGoMod := filepath.Join(appsDir, entry.Name(), "go.mod")
		if _, err := os.Stat(appGoMod); err != nil {
			continue
		}
		modulePath, err := ReadModulePath(appGoMod)
		if err != nil {
			return "", err
		}
		if base := BaseModulePath(modulePath); base != "" {
			return base, nil
		}
	}

	return "", fmt.Errorf("cannot infer base module: no go.mod found under %s", rootDir)
}
