package cutter

import (
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"testing"

	"golang.org/x/mod/modfile"
)

func TestDetectProjectContext(t *testing.T) {
	t.Run("root go.mod only", func(t *testing.T) {
		rootDir := t.TempDir()
		nestedDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		ctx, err := detectProjectContext(nestedDir)
		if err != nil {
			t.Fatalf("detectProjectContext returned error: %v", err)
		}

		if ctx.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", ctx.rootDir, rootDir)
		}
		if ctx.goModPath != filepath.Join(rootDir, "go.mod") {
			t.Fatalf("goModPath = %q, want %q", ctx.goModPath, filepath.Join(rootDir, "go.mod"))
		}
		if ctx.goWorkPath != "" {
			t.Fatalf("goWorkPath = %q, want empty", ctx.goWorkPath)
		}
	})

	t.Run("outer go.work keeps workspace context and nearest module", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		moduleDir := filepath.Join(workspaceRoot, "apps", "demo")
		if err := os.MkdirAll(moduleDir, 0o755); err != nil {
			t.Fatalf("mkdir module dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write nested go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workspaceRoot, "go.work"), []byte("go 1.22\n\nuse ./apps/demo\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		ctx, err := detectProjectContext(filepath.Join(moduleDir, "internal"))
		if err != nil {
			t.Fatalf("detectProjectContext returned error: %v", err)
		}

		if ctx.rootDir != workspaceRoot {
			t.Fatalf("rootDir = %q, want %q", ctx.rootDir, workspaceRoot)
		}
		if ctx.goWorkPath != filepath.Join(workspaceRoot, "go.work") {
			t.Fatalf("goWorkPath = %q, want %q", ctx.goWorkPath, filepath.Join(workspaceRoot, "go.work"))
		}
		if ctx.goModPath != filepath.Join(moduleDir, "go.mod") {
			t.Fatalf("goModPath = %q, want %q", ctx.goModPath, filepath.Join(moduleDir, "go.mod"))
		}

		wantUseDirs := []string{moduleDir}
		if !reflect.DeepEqual(ctx.goWorkUseDirs, wantUseDirs) {
			t.Fatalf("goWorkUseDirs = %#v, want %#v", ctx.goWorkUseDirs, wantUseDirs)
		}
	})

	t.Run("root go.work only", func(t *testing.T) {
		rootDir := t.TempDir()
		memberDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(memberDir, 0o755); err != nil {
			t.Fatalf("mkdir member dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memberDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}
		goWorkContent := []byte("go 1.22\n\nuse ./apps/demo\n")
		if err := os.WriteFile(filepath.Join(rootDir, "go.work"), goWorkContent, 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		ctx, err := detectProjectContext(rootDir)
		if err != nil {
			t.Fatalf("detectProjectContext returned error: %v", err)
		}

		if ctx.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", ctx.rootDir, rootDir)
		}
		if ctx.goModPath != "" {
			t.Fatalf("goModPath = %q, want empty", ctx.goModPath)
		}
		if ctx.goWorkPath != filepath.Join(rootDir, "go.work") {
			t.Fatalf("goWorkPath = %q, want %q", ctx.goWorkPath, filepath.Join(rootDir, "go.work"))
		}

		wantUseDirs := []string{memberDir}
		if !reflect.DeepEqual(ctx.goWorkUseDirs, wantUseDirs) {
			t.Fatalf("goWorkUseDirs = %#v, want %#v", ctx.goWorkUseDirs, wantUseDirs)
		}
	})

	t.Run("drops nearest go.mod when workspace use dirs do not include it", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		unrelatedModuleDir := filepath.Join(workspaceRoot, "apps", "old")
		memberDir := filepath.Join(workspaceRoot, "services", "api")
		currentDir := filepath.Join(unrelatedModuleDir, "internal")
		if err := os.MkdirAll(currentDir, 0o755); err != nil {
			t.Fatalf("mkdir current dir: %v", err)
		}
		if err := os.MkdirAll(memberDir, 0o755); err != nil {
			t.Fatalf("mkdir member dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(unrelatedModuleDir, "go.mod"), []byte("module example.com/old\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write unrelated go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memberDir, "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workspaceRoot, "go.work"), []byte("go 1.22\n\nuse ./services/api\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		ctx, err := detectProjectContext(currentDir)
		if err != nil {
			t.Fatalf("detectProjectContext returned error: %v", err)
		}

		if ctx.rootDir != workspaceRoot {
			t.Fatalf("rootDir = %q, want %q", ctx.rootDir, workspaceRoot)
		}
		if ctx.goWorkPath != filepath.Join(workspaceRoot, "go.work") {
			t.Fatalf("goWorkPath = %q, want %q", ctx.goWorkPath, filepath.Join(workspaceRoot, "go.work"))
		}
		if ctx.goModPath != "" {
			t.Fatalf("goModPath = %q, want empty", ctx.goModPath)
		}

		wantUseDirs := []string{memberDir}
		if !reflect.DeepEqual(ctx.goWorkUseDirs, wantUseDirs) {
			t.Fatalf("goWorkUseDirs = %#v, want %#v", ctx.goWorkUseDirs, wantUseDirs)
		}
	})

	t.Run("returns error when no go.mod or go.work exists", func(t *testing.T) {
		rootDir := t.TempDir()
		nestedDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(nestedDir, 0o755); err != nil {
			t.Fatalf("mkdir nested dir: %v", err)
		}

		_, err := detectProjectContext(nestedDir)
		if err == nil {
			t.Fatal("detectProjectContext returned nil error, want non-nil")
		}

		const wantErr = "current directory is not a Go project root: no go.mod or go.work found"
		if err.Error() != wantErr {
			t.Fatalf("error = %q, want %q", err.Error(), wantErr)
		}
	})
}

func TestParseGoWorkUseDirs(t *testing.T) {
	rootDir := t.TempDir()
	goWorkPath := filepath.Join(rootDir, "go.work")
	goWorkContent := []byte("go 1.22\n\nuse (\n\t./apps/api\n\t./libs/../libs/shared\n)\n")
	if err := os.WriteFile(goWorkPath, goWorkContent, 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}

	useDirs, err := parseGoWorkUseDirs(goWorkPath)
	if err != nil {
		t.Fatalf("parseGoWorkUseDirs returned error: %v", err)
	}

	want := []string{
		filepath.Join(rootDir, "apps", "api"),
		filepath.Join(rootDir, "libs", "shared"),
	}
	if !reflect.DeepEqual(useDirs, want) {
		t.Fatalf("useDirs = %#v, want %#v", useDirs, want)
	}
}

func TestFindProjectRoot(t *testing.T) {
	t.Run("workspace member returns workspace root and member module path", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		memberDir := filepath.Join(workspaceRoot, "apps", "demo")
		currentDir := filepath.Join(memberDir, "internal")
		if err := os.MkdirAll(currentDir, 0o755); err != nil {
			t.Fatalf("mkdir current dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memberDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workspaceRoot, "go.work"), []byte("go 1.22\n\nuse ./apps/demo\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		rootDir, isWorkspace, modulePath, err := findProjectRoot(currentDir)
		if err != nil {
			t.Fatalf("findProjectRoot returned error: %v", err)
		}

		if rootDir != workspaceRoot {
			t.Fatalf("rootDir = %q, want %q", rootDir, workspaceRoot)
		}
		if !isWorkspace {
			t.Fatal("isWorkspace = false, want true")
		}
		if modulePath != "example.com/demo" {
			t.Fatalf("modulePath = %q, want %q", modulePath, "example.com/demo")
		}
	})

	t.Run("workspace without valid member module returns error", func(t *testing.T) {
		workspaceRoot := t.TempDir()
		unrelatedModuleDir := filepath.Join(workspaceRoot, "apps", "old")
		memberDir := filepath.Join(workspaceRoot, "services", "api")
		currentDir := filepath.Join(unrelatedModuleDir, "internal")
		if err := os.MkdirAll(currentDir, 0o755); err != nil {
			t.Fatalf("mkdir current dir: %v", err)
		}
		if err := os.MkdirAll(memberDir, 0o755); err != nil {
			t.Fatalf("mkdir member dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(unrelatedModuleDir, "go.mod"), []byte("module example.com/old\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write unrelated go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memberDir, "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(workspaceRoot, "go.work"), []byte("go 1.22\n\nuse ./services/api\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		_, _, _, err := findProjectRoot(currentDir)
		if err == nil {
			t.Fatal("findProjectRoot returned nil error, want non-nil")
		}

		const wantErr = "workspace found but no matching module root found for current directory"
		if err.Error() != wantErr {
			t.Fatalf("error = %q, want %q", err.Error(), wantErr)
		}
	})
}

func TestResolveAppModulePath(t *testing.T) {
	t.Run("prefers apps source go.mod", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(appDir, "go.mod"), []byte("module example.com/demo\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write app go.mod: %v", err)
		}

		got, err := resolveAppModulePath(projectContext{rootDir: rootDir, goModPath: filepath.Join(rootDir, "go.mod")}, "demo")
		if err != nil {
			t.Fatalf("resolveAppModulePath returned error: %v", err)
		}

		if got.rootDir != appDir {
			t.Fatalf("rootDir = %q, want %q", got.rootDir, appDir)
		}
		if got.goModPath != filepath.Join(appDir, "go.mod") {
			t.Fatalf("goModPath = %q, want %q", got.goModPath, filepath.Join(appDir, "go.mod"))
		}
		if got.modulePath != "example.com/demo" {
			t.Fatalf("modulePath = %q, want %q", got.modulePath, "example.com/demo")
		}
		if got.oldImportPrefix != "example.com/demo" {
			t.Fatalf("oldImportPrefix = %q, want %q", got.oldImportPrefix, "example.com/demo")
		}
	})

	t.Run("falls back to root go.mod", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		rootGoModPath := filepath.Join(rootDir, "go.mod")
		if err := os.WriteFile(rootGoModPath, []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}

		got, err := resolveAppModulePath(projectContext{rootDir: rootDir, goModPath: rootGoModPath}, "demo")
		if err != nil {
			t.Fatalf("resolveAppModulePath returned error: %v", err)
		}

		if got.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", got.rootDir, rootDir)
		}
		if got.goModPath != rootGoModPath {
			t.Fatalf("goModPath = %q, want %q", got.goModPath, rootGoModPath)
		}
		if got.modulePath != "example.com/root" {
			t.Fatalf("modulePath = %q, want %q", got.modulePath, "example.com/root")
		}
		if got.oldImportPrefix != "example.com/root/apps/demo" {
			t.Fatalf("oldImportPrefix = %q, want %q", got.oldImportPrefix, "example.com/root/apps/demo")
		}
	})

	t.Run("prefers root go.mod when root go.mod and go.work both exist", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		memberDir := filepath.Join(rootDir, "services", "api")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(memberDir, 0o755); err != nil {
			t.Fatalf("mkdir member dir: %v", err)
		}
		rootGoModPath := filepath.Join(rootDir, "go.mod")
		if err := os.WriteFile(rootGoModPath, []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(memberDir, "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}

		got, err := resolveAppModulePath(projectContext{
			rootDir:       rootDir,
			goWorkPath:    filepath.Join(rootDir, "go.work"),
			goWorkUseDirs: []string{memberDir},
		}, "demo")
		if err != nil {
			t.Fatalf("resolveAppModulePath returned error: %v", err)
		}

		if got.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", got.rootDir, rootDir)
		}
		if got.goModPath != rootGoModPath {
			t.Fatalf("goModPath = %q, want %q", got.goModPath, rootGoModPath)
		}
		if got.modulePath != "example.com/root" {
			t.Fatalf("modulePath = %q, want %q", got.modulePath, "example.com/root")
		}
		if got.oldImportPrefix != "example.com/root/apps/demo" {
			t.Fatalf("oldImportPrefix = %q, want %q", got.oldImportPrefix, "example.com/root/apps/demo")
		}
	})

	t.Run("nearest module does not override root go.mod priority", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		memberDir := filepath.Join(rootDir, "services", "api")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(memberDir, 0o755); err != nil {
			t.Fatalf("mkdir member dir: %v", err)
		}
		rootGoModPath := filepath.Join(rootDir, "go.mod")
		memberGoModPath := filepath.Join(memberDir, "go.mod")
		if err := os.WriteFile(rootGoModPath, []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(memberGoModPath, []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write member go.mod: %v", err)
		}

		got, err := resolveAppModulePath(projectContext{
			rootDir:       rootDir,
			goModPath:     memberGoModPath,
			goWorkPath:    filepath.Join(rootDir, "go.work"),
			goWorkUseDirs: []string{memberDir},
		}, "demo")
		if err != nil {
			t.Fatalf("resolveAppModulePath returned error: %v", err)
		}

		if got.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", got.rootDir, rootDir)
		}
		if got.goModPath != rootGoModPath {
			t.Fatalf("goModPath = %q, want %q", got.goModPath, rootGoModPath)
		}
		if got.modulePath != "example.com/root" {
			t.Fatalf("modulePath = %q, want %q", got.modulePath, "example.com/root")
		}
		if got.oldImportPrefix != "example.com/root/apps/demo" {
			t.Fatalf("oldImportPrefix = %q, want %q", got.oldImportPrefix, "example.com/root/apps/demo")
		}
	})

	t.Run("returns error when current module does not cover app dir", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		moduleDir := filepath.Join(rootDir, "services", "api")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(moduleDir, 0o755); err != nil {
			t.Fatalf("mkdir module dir: %v", err)
		}
		moduleGoModPath := filepath.Join(moduleDir, "go.mod")
		if err := os.WriteFile(moduleGoModPath, []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write module go.mod: %v", err)
		}

		_, err := resolveAppModulePath(projectContext{
			rootDir:   rootDir,
			goModPath: moduleGoModPath,
		}, "demo")
		if err == nil {
			t.Fatal("resolveAppModulePath returned nil error, want non-nil")
		}
		if !strings.Contains(err.Error(), "go.work found, but cannot resolve module for apps/demo") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "go.work found, but cannot resolve module for apps/demo")
		}
	})

	t.Run("infers unique workspace module that covers app dir", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		moduleGoModPath := filepath.Join(rootDir, "go.mod")
		if err := os.WriteFile(moduleGoModPath, []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write module go.mod: %v", err)
		}

		got, err := resolveAppModulePath(projectContext{
			rootDir:       rootDir,
			goWorkPath:    filepath.Join(rootDir, "go.work"),
			goWorkUseDirs: []string{rootDir},
		}, "demo")
		if err != nil {
			t.Fatalf("resolveAppModulePath returned error: %v", err)
		}

		if got.rootDir != rootDir {
			t.Fatalf("rootDir = %q, want %q", got.rootDir, rootDir)
		}
		if got.goModPath != moduleGoModPath {
			t.Fatalf("goModPath = %q, want %q", got.goModPath, moduleGoModPath)
		}
		if got.modulePath != "example.com/root" {
			t.Fatalf("modulePath = %q, want %q", got.modulePath, "example.com/root")
		}
		if got.oldImportPrefix != "example.com/root/apps/demo" {
			t.Fatalf("oldImportPrefix = %q, want %q", got.oldImportPrefix, "example.com/root/apps/demo")
		}
	})

	t.Run("returns clear error when workspace has no unique covering module", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		moduleDirA := filepath.Join(rootDir, "services", "api")
		moduleDirB := filepath.Join(rootDir, "tools", "cli")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(moduleDirA, 0o755); err != nil {
			t.Fatalf("mkdir module dir A: %v", err)
		}
		if err := os.MkdirAll(moduleDirB, 0o755); err != nil {
			t.Fatalf("mkdir module dir B: %v", err)
		}
		if err := os.WriteFile(filepath.Join(moduleDirA, "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write module A go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(moduleDirB, "go.mod"), []byte("module example.com/cli\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write module B go.mod: %v", err)
		}

		_, err := resolveAppModulePath(projectContext{
			rootDir:       rootDir,
			goWorkPath:    filepath.Join(rootDir, "go.work"),
			goWorkUseDirs: []string{moduleDirA, moduleDirB},
		}, "demo")
		if err == nil {
			t.Fatal("resolveAppModulePath returned nil error, want non-nil")
		}
		if !strings.Contains(err.Error(), "go.work found, but cannot resolve module for apps/demo") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "go.work found, but cannot resolve module for apps/demo")
		}
	})

	t.Run("returns clear error when single workspace module is unrelated", func(t *testing.T) {
		rootDir := t.TempDir()
		appDir := filepath.Join(rootDir, "apps", "demo")
		moduleDir := filepath.Join(rootDir, "services", "api")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(moduleDir, 0o755); err != nil {
			t.Fatalf("mkdir module dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(moduleDir, "go.mod"), []byte("module example.com/api\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write module go.mod: %v", err)
		}

		_, err := resolveAppModulePath(projectContext{
			rootDir:       rootDir,
			goWorkPath:    filepath.Join(rootDir, "go.work"),
			goWorkUseDirs: []string{moduleDir},
		}, "demo")
		if err == nil {
			t.Fatal("resolveAppModulePath returned nil error, want non-nil")
		}
		if !strings.Contains(err.Error(), "go.work found, but cannot resolve module for apps/demo") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "go.work found, but cannot resolve module for apps/demo")
		}
	})
}

func TestResolveProjectCloneMode(t *testing.T) {
	t.Run("root go.mod uses root module mode", func(t *testing.T) {
		rootDir := t.TempDir()
		rootGoModPath := filepath.Join(rootDir, "go.mod")
		if err := os.WriteFile(rootGoModPath, []byte("module example.com/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}

		got := resolveProjectCloneMode(projectContext{
			rootDir:   rootDir,
			goModPath: rootGoModPath,
		})

		if got != projectCloneModeRootModule {
			t.Fatalf("clone mode = %q, want %q", got, projectCloneModeRootModule)
		}
	})

	t.Run("workspace only root uses workspace only mode", func(t *testing.T) {
		rootDir := t.TempDir()
		goWorkPath := filepath.Join(rootDir, "go.work")
		if err := os.WriteFile(goWorkPath, []byte("go 1.22\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}

		got := resolveProjectCloneMode(projectContext{
			rootDir:    rootDir,
			goWorkPath: goWorkPath,
		})

		if got != projectCloneModeWorkspaceOnly {
			t.Fatalf("clone mode = %q, want %q", got, projectCloneModeWorkspaceOnly)
		}
	})
}

func TestMaybeModifyGoMod(t *testing.T) {
	t.Run("missing go.mod skips without error", func(t *testing.T) {
		dstDir := t.TempDir()

		if err := maybeModifyGoMod(dstDir, "demo"); err != nil {
			t.Fatalf("maybeModifyGoMod returned error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(dstDir, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("go.mod stat err = %v, want not exist", err)
		}
	})

	t.Run("existing go.mod is updated", func(t *testing.T) {
		dstDir := t.TempDir()
		goModPath := filepath.Join(dstDir, "go.mod")
		if err := os.WriteFile(goModPath, []byte("module github.com/example/old\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		if err := maybeModifyGoMod(dstDir, "new"); err != nil {
			t.Fatalf("maybeModifyGoMod returned error: %v", err)
		}

		content, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("read go.mod: %v", err)
		}
		if !strings.Contains(string(content), "module github.com/example/new") {
			t.Fatalf("go.mod = %q, want updated module path", string(content))
		}
	})
}

func TestMaybeModifyAppGoMod(t *testing.T) {
	t.Run("missing app go.mod skips without error", func(t *testing.T) {
		appDir := t.TempDir()

		if err := maybeModifyAppGoMod(appDir, "oldapp", "newapp"); err != nil {
			t.Fatalf("maybeModifyAppGoMod returned error: %v", err)
		}

		if _, err := os.Stat(filepath.Join(appDir, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("go.mod stat err = %v, want not exist", err)
		}
	})

	t.Run("updates app module name only in module statement", func(t *testing.T) {
		appDir := t.TempDir()
		goModPath := filepath.Join(appDir, "go.mod")
		content := []byte("module github.com/example/apps/oldapp\n\ngo 1.22\n\nrequire github.com/example/shared-oldapp v0.0.0\n")
		if err := os.WriteFile(goModPath, content, 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		if err := maybeModifyAppGoMod(appDir, "oldapp", "newapp"); err != nil {
			t.Fatalf("maybeModifyAppGoMod returned error: %v", err)
		}

		updated, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("read go.mod: %v", err)
		}
		if !strings.Contains(string(updated), "module github.com/example/apps/newapp") {
			t.Fatalf("go.mod = %q, want updated module path", string(updated))
		}
		if !strings.Contains(string(updated), "require github.com/example/shared-oldapp v0.0.0") {
			t.Fatalf("go.mod = %q, want require unchanged", string(updated))
		}
	})

	t.Run("updates versioned app module and matching require replace entries", func(t *testing.T) {
		appDir := t.TempDir()
		goModPath := filepath.Join(appDir, "go.mod")
		content := []byte("module github.com/example/apps/oldapp/v2\n\ngo 1.22\n\nrequire (\n\tgithub.com/example/apps/oldapp/submod v0.0.0\n\tgithub.com/example/apps/oldapp/v2/internal v0.0.0\n\tgithub.com/example/apps/oldappx/submod v0.0.0\n)\n\nreplace (\n\tgithub.com/example/apps/oldapp/submod => ../submod\n\tgithub.com/example/apps/oldapp/v2/internal => github.com/example/apps/oldapp/v2/internal v0.0.1\n\tgithub.com/example/apps/oldappx/submod => github.com/example/apps/oldappx/submod v0.0.1\n)\n")
		if err := os.WriteFile(goModPath, content, 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		if err := maybeModifyAppGoMod(appDir, "oldapp", "newapp"); err != nil {
			t.Fatalf("maybeModifyAppGoMod returned error: %v", err)
		}

		updated, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("read go.mod: %v", err)
		}
		updatedContent := string(updated)
		if !strings.Contains(updatedContent, "module github.com/example/apps/newapp/v2") {
			t.Fatalf("go.mod = %q, want updated versioned app module path", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/oldapp/submod v0.0.0") {
			t.Fatalf("go.mod = %q, want non-prefix require unchanged", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/newapp/v2/internal v0.0.0") {
			t.Fatalf("go.mod = %q, want updated versioned require path", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/oldapp/submod => ../submod") {
			t.Fatalf("go.mod = %q, want non-prefix replace unchanged", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/newapp/v2/internal => github.com/example/apps/newapp/v2/internal v0.0.1") {
			t.Fatalf("go.mod = %q, want replace module paths updated", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/oldappx/submod v0.0.0") {
			t.Fatalf("go.mod = %q, want similar prefix require unchanged", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/oldappx/submod => github.com/example/apps/oldappx/submod v0.0.1") {
			t.Fatalf("go.mod = %q, want similar prefix replace unchanged", updatedContent)
		}
	})

	t.Run("only rewrites dependencies within current app module prefix", func(t *testing.T) {
		appDir := t.TempDir()
		goModPath := filepath.Join(appDir, "go.mod")
		content := []byte("module github.com/example/apps/oldapp/v2\n\ngo 1.22\n\nrequire (\n\tgithub.com/example/apps/oldapp/v2/submod v0.0.0\n\tgithub.com/acme/oldapp/sdk v1.2.3\n)\n\nreplace (\n\tgithub.com/example/apps/oldapp/v2/submod => github.com/example/apps/oldapp/v2/submod v0.0.1\n\tgithub.com/acme/oldapp/sdk => github.com/acme/oldapp/sdk v1.2.4\n)\n")
		if err := os.WriteFile(goModPath, content, 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}

		if err := maybeModifyAppGoMod(appDir, "oldapp", "newapp"); err != nil {
			t.Fatalf("maybeModifyAppGoMod returned error: %v", err)
		}

		updated, err := os.ReadFile(goModPath)
		if err != nil {
			t.Fatalf("read go.mod: %v", err)
		}
		updatedContent := string(updated)
		if !strings.Contains(updatedContent, "github.com/example/apps/newapp/v2/submod v0.0.0") {
			t.Fatalf("go.mod = %q, want current app require updated", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/example/apps/newapp/v2/submod => github.com/example/apps/newapp/v2/submod v0.0.1") {
			t.Fatalf("go.mod = %q, want current app replace updated", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/acme/oldapp/sdk v1.2.3") {
			t.Fatalf("go.mod = %q, want unrelated require unchanged", updatedContent)
		}
		if !strings.Contains(updatedContent, "github.com/acme/oldapp/sdk => github.com/acme/oldapp/sdk v1.2.4") {
			t.Fatalf("go.mod = %q, want unrelated replace unchanged", updatedContent)
		}
	})
}

func TestRenameAppModulePath(t *testing.T) {
	t.Run("rewrites exact app path segment before major version suffix", func(t *testing.T) {
		got := renameAppModulePath("github.com/example/apps/oldapp/v2", "github.com/example/apps/oldapp/v2", "oldapp", "newapp")
		if got != "github.com/example/apps/newapp/v2" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/apps/newapp/v2")
		}
	})

	t.Run("does not rewrite similar prefix segment", func(t *testing.T) {
		got := renameAppModulePath("github.com/example/apps/oldappx/v2", "github.com/example/apps/oldapp/v2", "oldapp", "newapp")
		if got != "github.com/example/apps/oldappx/v2" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/apps/oldappx/v2")
		}
	})

	t.Run("does not rewrite unrelated module path outside app prefix", func(t *testing.T) {
		got := renameAppModulePath("github.com/acme/oldapp/sdk", "github.com/example/apps/oldapp/v2", "oldapp", "newapp")
		if got != "github.com/acme/oldapp/sdk" {
			t.Fatalf("module path = %q, want %q", got, "github.com/acme/oldapp/sdk")
		}
	})
}

func TestCloneApp(t *testing.T) {
	t.Run("root go.mod single module rewrites app import without creating app go.mod", func(t *testing.T) {
		rootDir := t.TempDir()
		oldAppDir := filepath.Join(rootDir, "apps", "oldapp")
		internalDir := filepath.Join(oldAppDir, "internal", "pkg")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			t.Fatalf("mkdir internal dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module github.com/example/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(oldAppDir, "main.go"), []byte("package main\n\nimport _ \"github.com/example/root/apps/oldapp/internal/pkg\"\n"), 0o644); err != nil {
			t.Fatalf("write app go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(internalDir, "pkg.go"), []byte("package pkg\n"), 0o644); err != nil {
			t.Fatalf("write internal go file: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, rootDir)
		defer restoreWorkingDir()

		if err := cloneApp("oldapp", "newapp"); err != nil {
			t.Fatalf("cloneApp returned error: %v", err)
		}

		newAppDir := filepath.Join(rootDir, "apps", "newapp")
		if info, err := os.Stat(newAppDir); err != nil {
			t.Fatalf("stat cloned app dir: %v", err)
		} else if !info.IsDir() {
			t.Fatalf("cloned app path %q is not a directory", newAppDir)
		}

		newAppGoFile, err := os.ReadFile(filepath.Join(newAppDir, "main.go"))
		if err != nil {
			t.Fatalf("read cloned app go file: %v", err)
		}
		if !strings.Contains(string(newAppGoFile), "github.com/example/root/apps/newapp/internal/pkg") {
			t.Fatalf("main.go = %q, want updated root module app import path", string(newAppGoFile))
		}
		if strings.Contains(string(newAppGoFile), "github.com/example/root/apps/oldapp/internal/pkg") {
			t.Fatalf("main.go = %q, want old app import path removed", string(newAppGoFile))
		}

		if _, err := os.Stat(filepath.Join(newAppDir, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("app go.mod stat err = %v, want not exist", err)
		}
	})

	t.Run("app go.mod import prefix is rewritten from app module path", func(t *testing.T) {
		rootDir := t.TempDir()
		oldAppDir := filepath.Join(rootDir, "apps", "oldapp")
		internalDir := filepath.Join(oldAppDir, "internal", "pkg")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			t.Fatalf("mkdir internal dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "go.mod"), []byte("module github.com/example/root\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(oldAppDir, "go.mod"), []byte("module github.com/example/apps/oldapp\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write app go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(oldAppDir, "main.go"), []byte("package main\n\nimport _ \"github.com/example/apps/oldapp/internal/pkg\"\n"), 0o644); err != nil {
			t.Fatalf("write app go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(internalDir, "pkg.go"), []byte("package pkg\n"), 0o644); err != nil {
			t.Fatalf("write internal go file: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, rootDir)
		defer restoreWorkingDir()

		if err := cloneApp("oldapp", "newapp"); err != nil {
			t.Fatalf("cloneApp returned error: %v", err)
		}

		newAppGoMod, err := os.ReadFile(filepath.Join(rootDir, "apps", "newapp", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned app go.mod: %v", err)
		}
		if !strings.Contains(string(newAppGoMod), "module github.com/example/apps/newapp") {
			t.Fatalf("go.mod = %q, want updated app module path", string(newAppGoMod))
		}

		newAppGoFile, err := os.ReadFile(filepath.Join(rootDir, "apps", "newapp", "main.go"))
		if err != nil {
			t.Fatalf("read cloned app go file: %v", err)
		}
		if !strings.Contains(string(newAppGoFile), "github.com/example/apps/newapp/internal/pkg") {
			t.Fatalf("main.go = %q, want updated app import path", string(newAppGoFile))
		}
		if strings.Contains(string(newAppGoFile), "github.com/example/root/apps/newapp/internal/pkg") {
			t.Fatalf("main.go = %q, want app module import path instead of root module path", string(newAppGoFile))
		}
	})

	t.Run("workspace root with app go.mod clones app without creating root go.mod", func(t *testing.T) {
		rootDir := t.TempDir()
		oldAppDir := filepath.Join(rootDir, "apps", "oldapp")
		internalDir := filepath.Join(oldAppDir, "internal", "pkg")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			t.Fatalf("mkdir internal dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(rootDir, "go.work"), []byte("go 1.22\n\nuse ./apps/oldapp\n"), 0o644); err != nil {
			t.Fatalf("write root go.work: %v", err)
		}
		if err := os.WriteFile(filepath.Join(oldAppDir, "go.mod"), []byte("module github.com/example/apps/oldapp\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write app go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(oldAppDir, "main.go"), []byte("package main\n\nimport _ \"github.com/example/apps/oldapp/internal/pkg\"\n"), 0o644); err != nil {
			t.Fatalf("write app go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(internalDir, "pkg.go"), []byte("package pkg\n"), 0o644); err != nil {
			t.Fatalf("write internal go file: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, rootDir)
		defer restoreWorkingDir()

		if err := cloneApp("oldapp", "newapp"); err != nil {
			t.Fatalf("cloneApp returned error: %v", err)
		}

		newAppDir := filepath.Join(rootDir, "apps", "newapp")
		if info, err := os.Stat(newAppDir); err != nil {
			t.Fatalf("stat cloned app dir: %v", err)
		} else if !info.IsDir() {
			t.Fatalf("cloned app path %q is not a directory", newAppDir)
		}

		newAppGoMod, err := os.ReadFile(filepath.Join(newAppDir, "go.mod"))
		if err != nil {
			t.Fatalf("read cloned app go.mod: %v", err)
		}
		if !strings.Contains(string(newAppGoMod), "module github.com/example/apps/newapp") {
			t.Fatalf("go.mod = %q, want updated app module path", string(newAppGoMod))
		}

		newAppGoFile, err := os.ReadFile(filepath.Join(newAppDir, "main.go"))
		if err != nil {
			t.Fatalf("read cloned app go file: %v", err)
		}
		if !strings.Contains(string(newAppGoFile), "github.com/example/apps/newapp/internal/pkg") {
			t.Fatalf("main.go = %q, want updated app import path", string(newAppGoFile))
		}

		if _, err := os.Stat(filepath.Join(rootDir, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("root go.mod stat err = %v, want not exist", err)
		}
	})
}

func TestRewriteModulePathForProjectClone(t *testing.T) {
	mappings := map[string]string{
		"github.com/example/templateproj":     "github.com/example/newproj",
		"github.com/example/templateproj/pkg": "github.com/example/newproj/pkg",
	}

	t.Run("replaces matching path segment", func(t *testing.T) {
		got, changed := rewriteModulePathForProjectClone("github.com/example/templateproj/pkg", mappings)
		if !changed {
			t.Fatal("changed = false, want true")
		}
		if got != "github.com/example/newproj/pkg" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/newproj/pkg")
		}
	})

	t.Run("leaves unrelated module path unchanged", func(t *testing.T) {
		got, changed := rewriteModulePathForProjectClone("github.com/example/pkg", mappings)
		if changed {
			t.Fatal("changed = true, want false")
		}
		if got != "github.com/example/pkg" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/pkg")
		}
	})

	t.Run("does not rewrite similar prefix module path", func(t *testing.T) {
		got, changed := rewriteModulePathForProjectClone("github.com/example/templateprojx/pkg", mappings)
		if changed {
			t.Fatal("changed = true, want false")
		}
		if got != "github.com/example/templateprojx/pkg" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/templateprojx/pkg")
		}
	})

	t.Run("rewrites matching subpackage import by exact prefix", func(t *testing.T) {
		got, changed := rewriteModulePathForProjectClone("github.com/example/templateproj/pkg/subpkg", mappings)
		if !changed {
			t.Fatal("changed = false, want true")
		}
		if got != "github.com/example/newproj/pkg/subpkg" {
			t.Fatalf("module path = %q, want %q", got, "github.com/example/newproj/pkg/subpkg")
		}
	})
}

func TestRewriteWorkspaceGoMod(t *testing.T) {
	t.Run("rewrites replace target module path without version", func(t *testing.T) {
		modFile, err := modfile.Parse("go.mod", []byte("module github.com/example/api\n\ngo 1.22\n\nreplace github.com/example/templateproj/pkg => ../../pkg\n"), nil)
		if err != nil {
			t.Fatalf("parse go.mod: %v", err)
		}

		replaceStmt := modFile.Replace[0]
		replaceStmt.New.Path = "github.com/example/templateproj/altpkg"
		replaceStmt.Syntax.Token[3] = "github.com/example/templateproj/altpkg"

		changed, err := rewriteWorkspaceGoMod(modFile, map[string]string{
			"github.com/example/templateproj":     "github.com/example/newproj",
			"github.com/example/templateproj/pkg": "github.com/example/newproj/pkg",
		})
		if err != nil {
			t.Fatalf("rewriteWorkspaceGoMod returned error: %v", err)
		}
		if !changed {
			t.Fatal("changed = false, want true")
		}

		formatted, err := modFile.Format()
		if err != nil {
			t.Fatalf("format go.mod: %v", err)
		}
		if !strings.Contains(string(formatted), "replace github.com/example/newproj/pkg => github.com/example/newproj/altpkg") {
			t.Fatalf("go.mod = %q, want updated replace target", string(formatted))
		}
	})
}

func TestCloneProject(t *testing.T) {
	t.Run("rejects target directory inside source root", func(t *testing.T) {
		parentDir := t.TempDir()
		sourceRoot := filepath.Join(parentDir, "templateproj")
		if err := os.MkdirAll(sourceRoot, 0o755); err != nil {
			t.Fatalf("mkdir source root: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceRoot, "go.mod"), []byte("module github.com/example/templateproj\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, sourceRoot)
		defer restoreWorkingDir()

		err := cloneProject(filepath.Join(sourceRoot, "nested", "clone"))
		if err == nil {
			t.Fatal("cloneProject returned nil error, want non-nil")
		}
		if !strings.Contains(err.Error(), "is inside project root") {
			t.Fatalf("error = %q, want substring %q", err.Error(), "is inside project root")
		}
	})

	t.Run("root go.mod only clone rewrites root module and imports", func(t *testing.T) {
		parentDir := t.TempDir()
		sourceRoot := filepath.Join(parentDir, "templateproj")
		targetRoot := filepath.Join(parentDir, "newproj")
		internalDir := filepath.Join(sourceRoot, "internal", "pkg")
		cmdDir := filepath.Join(sourceRoot, "cmd", "demo")
		if err := os.MkdirAll(internalDir, 0o755); err != nil {
			t.Fatalf("mkdir internal dir: %v", err)
		}
		if err := os.MkdirAll(cmdDir, 0o755); err != nil {
			t.Fatalf("mkdir cmd dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceRoot, "go.mod"), []byte("module github.com/example/templateproj\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write root go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(cmdDir, "main.go"), []byte("package main\n\nimport _ \"github.com/example/templateproj/internal/pkg\"\n"), 0o644); err != nil {
			t.Fatalf("write main.go: %v", err)
		}
		if err := os.WriteFile(filepath.Join(internalDir, "pkg.go"), []byte("package pkg\n"), 0o644); err != nil {
			t.Fatalf("write pkg.go: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, sourceRoot)
		defer restoreWorkingDir()

		if err := cloneProject(targetRoot); err != nil {
			t.Fatalf("cloneProject returned error: %v", err)
		}

		if info, err := os.Stat(filepath.Join(targetRoot, "cmd", "demo")); err != nil {
			t.Fatalf("stat cloned cmd dir: %v", err)
		} else if !info.IsDir() {
			t.Fatalf("cloned cmd path is not a directory")
		}

		rootGoMod, err := os.ReadFile(filepath.Join(targetRoot, "go.mod"))
		if err != nil {
			t.Fatalf("read cloned root go.mod: %v", err)
		}
		if !strings.Contains(string(rootGoMod), "module github.com/example/newproj") {
			t.Fatalf("root go.mod = %q, want updated module path", string(rootGoMod))
		}

		mainGo, err := os.ReadFile(filepath.Join(targetRoot, "cmd", "demo", "main.go"))
		if err != nil {
			t.Fatalf("read cloned main.go: %v", err)
		}
		if !strings.Contains(string(mainGo), "github.com/example/newproj/internal/pkg") {
			t.Fatalf("main.go = %q, want updated import path", string(mainGo))
		}

		if _, err := os.Stat(filepath.Join(targetRoot, "go.work")); !os.IsNotExist(err) {
			t.Fatalf("root go.work stat err = %v, want not exist", err)
		}
	})

	t.Run("workspace only clone rewrites member module go mods", func(t *testing.T) {
		parentDir := t.TempDir()
		sourceRoot := filepath.Join(parentDir, "templateproj")
		targetRoot := filepath.Join(parentDir, "newproj")
		appDir := filepath.Join(sourceRoot, "apps", "demo")
		serviceDir := filepath.Join(sourceRoot, "services", "api")
		thirdPartyDir := filepath.Join(sourceRoot, "tools", "shim")
		pkgDir := filepath.Join(sourceRoot, "pkg")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(serviceDir, 0o755); err != nil {
			t.Fatalf("mkdir service dir: %v", err)
		}
		if err := os.MkdirAll(thirdPartyDir, 0o755); err != nil {
			t.Fatalf("mkdir third party dir: %v", err)
		}
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			t.Fatalf("mkdir pkg dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceRoot, "go.work"), []byte("go 1.22\n\nuse (\n\t./apps/demo\n\t./services/api\n\t./tools/shim\n\t./pkg\n)\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}
		if err := os.WriteFile(filepath.Join(appDir, "go.mod"), []byte("module github.com/example/templateproj\n\ngo 1.22\n\nrequire github.com/example/templateproj/pkg v0.0.0\n\nreplace github.com/example/templateproj/pkg => ../../pkg\n"), 0o644); err != nil {
			t.Fatalf("write app go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(serviceDir, "go.mod"), []byte("module github.com/example/api\n\ngo 1.22\n\nrequire (\n\tgithub.com/example/templateproj/pkg v0.0.0\n\tgithub.com/example/templateprojx/pkg v0.0.0\n)\n\nreplace github.com/example/templateproj/pkg => ../../pkg\n"), 0o644); err != nil {
			t.Fatalf("write service go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(serviceDir, "service.go"), []byte("package api\n\nimport (\n\t_ \"github.com/example/templateproj/pkg\"\n\t_ \"github.com/example/templateprojx/pkg\"\n)\n"), 0o644); err != nil {
			t.Fatalf("write service go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(thirdPartyDir, "go.mod"), []byte("module github.com/example/third-templateproj-tools\n\ngo 1.22\n\nrequire github.com/example/templateproj/pkg v0.0.0\n"), 0o644); err != nil {
			t.Fatalf("write third party go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(thirdPartyDir, "shim.go"), []byte("package shim\n\nimport (\n\t_ \"github.com/example/templateproj/pkg\"\n\t_ \"github.com/example/third-templateproj-tools/internal/x\"\n)\n"), 0o644); err != nil {
			t.Fatalf("write third party go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module github.com/example/templateproj/pkg\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write pkg go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "pkg.go"), []byte("package pkg\n"), 0o644); err != nil {
			t.Fatalf("write pkg go file: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, sourceRoot)
		defer restoreWorkingDir()

		if err := cloneProject(targetRoot); err != nil {
			t.Fatalf("cloneProject returned error: %v", err)
		}

		if info, err := os.Stat(filepath.Join(targetRoot, "go.work")); err != nil {
			t.Fatalf("stat cloned root go.work: %v", err)
		} else if info.IsDir() {
			t.Fatalf("cloned root go.work path is a directory")
		}

		appGoMod, err := os.ReadFile(filepath.Join(targetRoot, "apps", "demo", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned app go.mod: %v", err)
		}
		if !strings.Contains(string(appGoMod), "module github.com/example/newproj") {
			t.Fatalf("app go.mod = %q, want updated module path", string(appGoMod))
		}
		if !strings.Contains(string(appGoMod), "require github.com/example/newproj/pkg v0.0.0") {
			t.Fatalf("app go.mod = %q, want updated require path", string(appGoMod))
		}
		if !strings.Contains(string(appGoMod), "replace github.com/example/newproj/pkg => ../../pkg") {
			t.Fatalf("app go.mod = %q, want updated replace path", string(appGoMod))
		}

		serviceGoMod, err := os.ReadFile(filepath.Join(targetRoot, "services", "api", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned service go.mod: %v", err)
		}
		if !strings.Contains(string(serviceGoMod), "module github.com/example/api") {
			t.Fatalf("service go.mod = %q, want unchanged module path", string(serviceGoMod))
		}
		if !strings.Contains(string(serviceGoMod), "github.com/example/newproj/pkg v0.0.0") {
			t.Fatalf("service go.mod = %q, want updated require path", string(serviceGoMod))
		}
		if !strings.Contains(string(serviceGoMod), "replace github.com/example/newproj/pkg => ../../pkg") {
			t.Fatalf("service go.mod = %q, want updated replace path", string(serviceGoMod))
		}
		if !strings.Contains(string(serviceGoMod), "github.com/example/templateprojx/pkg v0.0.0") {
			t.Fatalf("service go.mod = %q, want similar prefix dependency unchanged", string(serviceGoMod))
		}

		serviceGoFile, err := os.ReadFile(filepath.Join(targetRoot, "services", "api", "service.go"))
		if err != nil {
			t.Fatalf("read cloned service go file: %v", err)
		}
		if !strings.Contains(string(serviceGoFile), "github.com/example/newproj/pkg") {
			t.Fatalf("service.go = %q, want updated project import", string(serviceGoFile))
		}
		if !strings.Contains(string(serviceGoFile), "github.com/example/templateprojx/pkg") {
			t.Fatalf("service.go = %q, want similar prefix third-party import unchanged", string(serviceGoFile))
		}

		thirdPartyGoMod, err := os.ReadFile(filepath.Join(targetRoot, "tools", "shim", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned third party go.mod: %v", err)
		}
		if !strings.Contains(string(thirdPartyGoMod), "module github.com/example/third-templateproj-tools") {
			t.Fatalf("third party go.mod = %q, want module path unchanged", string(thirdPartyGoMod))
		}
		if !strings.Contains(string(thirdPartyGoMod), "require github.com/example/newproj/pkg v0.0.0") {
			t.Fatalf("third party go.mod = %q, want project dependency updated", string(thirdPartyGoMod))
		}

		thirdPartyGoFile, err := os.ReadFile(filepath.Join(targetRoot, "tools", "shim", "shim.go"))
		if err != nil {
			t.Fatalf("read cloned third party go file: %v", err)
		}
		if !strings.Contains(string(thirdPartyGoFile), "github.com/example/newproj/pkg") {
			t.Fatalf("third party go file = %q, want project import updated", string(thirdPartyGoFile))
		}
		if !strings.Contains(string(thirdPartyGoFile), "github.com/example/third-templateproj-tools/internal/x") {
			t.Fatalf("third party go file = %q, want unrelated module import unchanged", string(thirdPartyGoFile))
		}

		pkgGoMod, err := os.ReadFile(filepath.Join(targetRoot, "pkg", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned pkg go.mod: %v", err)
		}
		if !strings.Contains(string(pkgGoMod), "module github.com/example/newproj/pkg") {
			t.Fatalf("pkg go.mod = %q, want updated module path", string(pkgGoMod))
		}

		if _, err := os.Stat(filepath.Join(targetRoot, "go.mod")); !os.IsNotExist(err) {
			t.Fatalf("root go.mod stat err = %v, want not exist", err)
		}
	})

	t.Run("workspace only app and pkg modules still build stable mappings", func(t *testing.T) {
		parentDir := t.TempDir()
		sourceRoot := filepath.Join(parentDir, "templateproj")
		targetRoot := filepath.Join(parentDir, "newproj")
		appDir := filepath.Join(sourceRoot, "apps", "demo")
		pkgDir := filepath.Join(sourceRoot, "pkg")
		if err := os.MkdirAll(appDir, 0o755); err != nil {
			t.Fatalf("mkdir app dir: %v", err)
		}
		if err := os.MkdirAll(pkgDir, 0o755); err != nil {
			t.Fatalf("mkdir pkg dir: %v", err)
		}
		if err := os.WriteFile(filepath.Join(sourceRoot, "go.work"), []byte("go 1.22\n\nuse (\n\t./apps/demo\n\t./pkg\n)\n"), 0o644); err != nil {
			t.Fatalf("write go.work: %v", err)
		}
		if err := os.WriteFile(filepath.Join(appDir, "go.mod"), []byte("module github.com/example/templateproj\n\ngo 1.22\n\nrequire github.com/example/templateproj/pkg v0.0.0\n\nreplace github.com/example/templateproj/pkg => ../../pkg\n"), 0o644); err != nil {
			t.Fatalf("write app go.mod: %v", err)
		}
		if err := os.WriteFile(filepath.Join(appDir, "main.go"), []byte("package main\n\nimport _ \"github.com/example/templateproj/pkg\"\n"), 0o644); err != nil {
			t.Fatalf("write app go file: %v", err)
		}
		if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module github.com/example/templateproj/pkg\n\ngo 1.22\n"), 0o644); err != nil {
			t.Fatalf("write pkg go.mod: %v", err)
		}

		restoreWorkingDir := chdirForTest(t, sourceRoot)
		defer restoreWorkingDir()

		if err := cloneProject(targetRoot); err != nil {
			t.Fatalf("cloneProject returned error: %v", err)
		}

		appGoMod, err := os.ReadFile(filepath.Join(targetRoot, "apps", "demo", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned app go.mod: %v", err)
		}
		if !strings.Contains(string(appGoMod), "module github.com/example/newproj") {
			t.Fatalf("app go.mod = %q, want updated module path", string(appGoMod))
		}
		if !strings.Contains(string(appGoMod), "require github.com/example/newproj/pkg v0.0.0") {
			t.Fatalf("app go.mod = %q, want updated require path", string(appGoMod))
		}
		if !strings.Contains(string(appGoMod), "replace github.com/example/newproj/pkg => ../../pkg") {
			t.Fatalf("app go.mod = %q, want updated replace path", string(appGoMod))
		}

		appGoFile, err := os.ReadFile(filepath.Join(targetRoot, "apps", "demo", "main.go"))
		if err != nil {
			t.Fatalf("read cloned app go file: %v", err)
		}
		if !strings.Contains(string(appGoFile), "github.com/example/newproj/pkg") {
			t.Fatalf("app go file = %q, want updated import path", string(appGoFile))
		}

		pkgGoMod, err := os.ReadFile(filepath.Join(targetRoot, "pkg", "go.mod"))
		if err != nil {
			t.Fatalf("read cloned pkg go.mod: %v", err)
		}
		if !strings.Contains(string(pkgGoMod), "module github.com/example/newproj/pkg") {
			t.Fatalf("pkg go.mod = %q, want updated module path", string(pkgGoMod))
		}
	})
}

func chdirForTest(t *testing.T, dir string) func() {
	t.Helper()

	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("get working directory: %v", err)
	}
	if err := os.Chdir(dir); err != nil {
		t.Fatalf("chdir to %q: %v", dir, err)
	}

	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory to %q: %v", originalDir, err)
		}
	}
}
