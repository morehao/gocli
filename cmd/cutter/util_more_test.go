package cutter

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	cases := []struct {
		name string
		path string
		want bool
	}{
		{"ignore git dir", ".git/config", true},
		{"ignore idea dir", ".idea/workspace.xml", true},
		{"ignore vscode dir", ".vscode/settings.json", true},
		{"ignore history dir", ".history/cmd/x.go", true},
		{"ignore nested node_modules", "apps/web/node_modules/x/y.js", true},
		{"ignore vendor dir", "vendor/github.com/x/y", true},
		{"ignore log dir", "logs/app.log", true},
		{"ignore tmp dir", "tmp/scratch.txt", true},
		{"ignore DS_Store file", "a/.DS_Store", true},
		{"ignore log file", "app.log", true},
		{"ignore tmp file", "scratch.tmp", true},
		{"keep go file", "main.go", false},
		{"keep nested go file", "internal/pkg/x.go", false},
		{"keep similar dir name", "logical", false},
		{"keep similar file name", "logs.txt", false},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			if got := shouldIgnore(tc.path); got != tc.want {
				t.Fatalf("shouldIgnore(%q) = %v, want %v", tc.path, got, tc.want)
			}
		})
	}
}

func TestIsGoProject(t *testing.T) {
	dir := t.TempDir()
	if isGoProject(dir) {
		t.Fatal("empty dir should not be a go project")
	}
	if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module example.com/x\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if !isGoProject(dir) {
		t.Fatal("dir with go.mod should be a go project")
	}
}

func TestIsGoWork(t *testing.T) {
	dir := t.TempDir()
	if isGoWork(dir) {
		t.Fatal("empty dir should not be a go workspace")
	}
	if err := os.WriteFile(filepath.Join(dir, "go.work"), []byte("go 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.work: %v", err)
	}
	if !isGoWork(dir) {
		t.Fatal("dir with go.work should be a go workspace")
	}
}

func TestReadModulePath(t *testing.T) {
	goModPath := filepath.Join(t.TempDir(), "go.mod")
	if _, err := readModulePath(goModPath); err == nil {
		t.Fatal("readModulePath on missing go.mod should error")
	}
	if err := os.WriteFile(goModPath, []byte("module github.com/example/demo\n\ngo 1.22\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	got, err := readModulePath(goModPath)
	if err != nil {
		t.Fatalf("readModulePath error: %v", err)
	}
	if got != "github.com/example/demo" {
		t.Fatalf("module path = %q, want %q", got, "github.com/example/demo")
	}
}

func TestHasModulePathPrefix(t *testing.T) {
	cases := []struct {
		path, prefix string
		want         bool
	}{
		{"github.com/a/b", "github.com/a/b", true},
		{"github.com/a/b/c", "github.com/a/b", true},
		{"github.com/a/bc", "github.com/a/b", false},
		{"github.com/a", "github.com/a/b", false},
		{"", "github.com/a/b", false},
	}
	for _, tc := range cases {
		if got := hasModulePathPrefix(tc.path, tc.prefix); got != tc.want {
			t.Errorf("hasModulePathPrefix(%q, %q) = %v, want %v", tc.path, tc.prefix, got, tc.want)
		}
	}
}

func TestBelongsToAppModulePrefix(t *testing.T) {
	oldModulePath := "github.com/example/apps/demo"
	cases := []struct {
		path string
		want bool
	}{
		{"github.com/example/apps/demo", true},
		{"github.com/example/apps/demo/internal/pkg", true},
		{"github.com/example/apps/demo2", false},
		{"github.com/example/pkg", false},
	}
	for _, tc := range cases {
		if got := belongsToAppModulePrefix(tc.path, oldModulePath); got != tc.want {
			t.Errorf("belongsToAppModulePrefix(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestIsModulePathReplacement(t *testing.T) {
	cases := []struct {
		path string
		want bool
	}{
		{"github.com/example/pkg", true},
		{"example.com/x/y", true},
		{"./pkg", false},
		{"../pkg", false},
		{"/abs/pkg", false},
		{"C:/windows/pkg", false},
		{"", false},
	}
	for _, tc := range cases {
		if got := isModulePathReplacement(tc.path); got != tc.want {
			t.Errorf("isModulePathReplacement(%q) = %v, want %v", tc.path, got, tc.want)
		}
	}
}

func TestRenameImportedAppPath(t *testing.T) {
	oldPrefix := "github.com/example/apps/oldapp"
	cases := []struct {
		importPath string
		oldName    string
		newName    string
		want       string
	}{
		{"github.com/example/apps/oldapp", "oldapp", "newapp", "github.com/example/apps/newapp"},
		{"github.com/example/apps/oldapp/internal/pkg", "oldapp", "newapp", "github.com/example/apps/newapp/internal/pkg"},
		{"github.com/example/apps/oldappx", "oldapp", "newapp", "github.com/example/apps/oldappx"},
		{"github.com/acme/other", "oldapp", "newapp", "github.com/acme/other"},
	}
	for _, tc := range cases {
		got := renameImportedAppPath(tc.importPath, oldPrefix, tc.oldName, tc.newName)
		if got != tc.want {
			t.Errorf("renameImportedAppPath(%q) = %q, want %q", tc.importPath, got, tc.want)
		}
	}
}

func TestIsPathWithinDir(t *testing.T) {
	parent := t.TempDir()
	baseDir := filepath.Join(parent, "base")
	nestedDir := filepath.Join(baseDir, "apps", "demo")
	otherDir := filepath.Join(parent, "other")
	for _, dir := range []string{baseDir, nestedDir, otherDir} {
		if err := os.MkdirAll(dir, 0o755); err != nil {
			t.Fatalf("mkdir %s: %v", dir, err)
		}
	}

	ok, err := isPathWithinDir(baseDir, nestedDir)
	if err != nil {
		t.Fatalf("isPathWithinDir error: %v", err)
	}
	if !ok {
		t.Fatal("nested dir should be within base dir")
	}

	ok, err = isPathWithinDir(nestedDir, baseDir)
	if err != nil {
		t.Fatalf("isPathWithinDir error: %v", err)
	}
	if ok {
		t.Fatal("base dir should not be within nested dir")
	}

	ok, err = isPathWithinDir(baseDir, otherDir)
	if err != nil {
		t.Fatalf("isPathWithinDir error: %v", err)
	}
	if ok {
		t.Fatal("sibling dir should not be within base dir")
	}
}

func TestDirContainsAndModuleCoversAppDir(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "demo")
	otherDir := filepath.Join(root, "services", "api")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir app dir: %v", err)
	}
	if err := os.MkdirAll(otherDir, 0o755); err != nil {
		t.Fatalf("mkdir other dir: %v", err)
	}

	if !dirContains(root, appDir) {
		t.Fatal("root should contain appDir")
	}
	if dirContains(appDir, otherDir) {
		t.Fatal("appDir should not contain otherDir")
	}
	if dirContains(otherDir, appDir) {
		t.Fatal("otherDir should not contain appDir")
	}

	if !moduleCoversAppDir(root, appDir) {
		t.Fatal("root module should cover appDir")
	}
	if moduleCoversAppDir(otherDir, appDir) {
		t.Fatal("other module should not cover appDir")
	}
}
