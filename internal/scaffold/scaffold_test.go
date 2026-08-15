package scaffold

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestShouldIgnore(t *testing.T) {
	cases := []struct {
		name     string
		path     string
		expected bool
	}{
		{name: "git dir", path: ".git/config", expected: true},
		{name: "idea dir", path: ".idea/workspace.xml", expected: true},
		{name: "vendor dir", path: "vendor/github.com/x/y/file.go", expected: true},
		{name: "ds store", path: "dir/.DS_Store", expected: true},
		{name: "log file", path: "logs/app.log", expected: true},
		{name: "normal file", path: "apps/demoapp/main.go", expected: false},
		{name: "normal dir", path: "pkg/code", expected: false},
	}
	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := ShouldIgnore(c.path); got != c.expected {
				t.Errorf("ShouldIgnore(%q) = %v, want %v", c.path, got, c.expected)
			}
		})
	}
}

func TestRestoreTemplateFiles(t *testing.T) {
	root := t.TempDir()
	files := map[string]string{
		"go.work.tmpl":         "go 1.26.1\nuse (./pkg)\n",
		"apps/a/go.mod.tmpl":   "module example.com/a\n",
		"apps/a/go.sum.tmpl":   "hash line\n",
		"apps/a/main.go.tmpl":  "package main\n",
		"apps/a/internal/x.go": "package x\n",
		"README.md":            "readme\n",
	}
	for rel, content := range files {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := RestoreTemplateFiles(root); err != nil {
		t.Fatalf("RestoreTemplateFiles: %v", err)
	}

	expected := map[string]bool{
		"go.work":              true,
		"apps/a/go.mod":        true,
		"apps/a/go.sum":        true,
		"apps/a/main.go":       true,
		"apps/a/internal/x.go": true, // 非 .go.tmpl 后缀文件不受影响
		"README.md":            true,
		"go.work.tmpl":         false,
		"apps/a/go.mod.tmpl":   false,
		"apps/a/go.sum.tmpl":   false,
		"apps/a/main.go.tmpl":  false,
	}
	for rel, wantExists := range expected {
		_, err := os.Stat(filepath.Join(root, rel))
		got := err == nil
		if got != wantExists {
			t.Errorf("%s exists = %v, want %v", rel, got, wantExists)
		}
	}
}

func TestRewriteGoContent(t *testing.T) {
	src := `package demoappdao

import (
	"github.com/example/demoapp/model"
	"github.com/example/pkg/dbclient"
	"github.com/gin-gonic/gin"
)

// @Router /v1/demoapp/user/create [post]
func Demo() { _ = gin.Default }
`
	mappings := map[string]string{
		"github.com/example/demoapp": "github.com/acme/backend/userapp",
		"github.com/example/pkg":     "github.com/acme/backend/pkg",
	}
	out, err := RewriteGoContent("user.go", []byte(src), "demoapp", "userapp", mappings)
	if err != nil {
		t.Fatalf("RewriteGoContent: %v", err)
	}
	result := string(out)
	for _, want := range []string{
		"package userappdao",
		"\"github.com/acme/backend/userapp/model\"",
		"\"github.com/acme/backend/pkg/dbclient\"",
		"@Router /v1/demoapp/user/create", // 注释文本由调用方另行替换，此处不变
	} {
		if !strings.Contains(result, want) {
			t.Errorf("result missing %q:\n%s", want, result)
		}
	}
	if strings.Contains(result, "github.com/example") {
		t.Errorf("result still contains github.com/example:\n%s", result)
	}
}

func TestRewriteGoModsInTree(t *testing.T) {
	root := t.TempDir()
	mods := map[string]string{
		"pkg/go.mod":          "module github.com/example/pkg\n\ngo 1.26.1\n",
		"apps/demoapp/go.mod": "module github.com/example/demoapp\n\ngo 1.26.1\n",
	}
	for rel, content := range mods {
		path := filepath.Join(root, rel)
		if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
			t.Fatal(err)
		}
		if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
			t.Fatal(err)
		}
	}

	if err := RewriteGoModsInTree(root, "github.com/example", "github.com/acme/backend"); err != nil {
		t.Fatalf("RewriteGoModsInTree: %v", err)
	}

	gotPkg, err := ReadModulePath(filepath.Join(root, "pkg", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if gotPkg != "github.com/acme/backend/pkg" {
		t.Errorf("pkg module = %q, want github.com/acme/backend/pkg", gotPkg)
	}
	gotApp, err := ReadModulePath(filepath.Join(root, "apps", "demoapp", "go.mod"))
	if err != nil {
		t.Fatal(err)
	}
	if gotApp != "github.com/acme/backend/demoapp" {
		t.Errorf("demoapp module = %q, want github.com/acme/backend/demoapp", gotApp)
	}
}

func TestAddGoWorkUse(t *testing.T) {
	goWorkPath := filepath.Join(t.TempDir(), "go.work")
	content := "go 1.26.1\n\nuse (\n\t./apps/demoapp\n\t./pkg\n)\n"
	if err := os.WriteFile(goWorkPath, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}

	if err := AddGoWorkUse(goWorkPath, "apps/userapp"); err != nil {
		t.Fatalf("AddGoWorkUse: %v", err)
	}
	got, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatal(err)
	}
	text := string(got)
	if !strings.Contains(text, "./apps/userapp") {
		t.Errorf("go.work missing ./apps/userapp:\n%s", text)
	}
	if !strings.Contains(text, "./apps/demoapp") {
		t.Errorf("go.work lost ./apps/demoapp:\n%s", text)
	}

	// 幂等：重复添加不产生重复条目
	before := string(got)
	if err := AddGoWorkUse(goWorkPath, "apps/userapp"); err != nil {
		t.Fatalf("AddGoWorkUse second: %v", err)
	}
	after, err := os.ReadFile(goWorkPath)
	if err != nil {
		t.Fatal(err)
	}
	if string(after) != before {
		t.Errorf("AddGoWorkUse not idempotent:\nbefore:\n%s\nafter:\n%s", before, after)
	}
}

func TestInferBaseModule(t *testing.T) {
	root := t.TempDir()
	if _, err := InferBaseModule(root); err == nil {
		t.Error("expected error for empty dir")
	}

	// 通过 pkg/go.mod 推断
	pkgDir := filepath.Join(root, "pkg")
	if err := os.MkdirAll(pkgDir, 0o755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(filepath.Join(pkgDir, "go.mod"), []byte("module github.com/acme/backend/pkg\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatal(err)
	}
	base, err := InferBaseModule(root)
	if err != nil {
		t.Fatalf("InferBaseModule: %v", err)
	}
	if base != "github.com/acme/backend" {
		t.Errorf("base = %q, want github.com/acme/backend", base)
	}
}

func TestValidateAppName(t *testing.T) {
	for _, name := range []string{"userapp", "demoapp", "a1"} {
		if err := ValidateAppName(name); err != nil {
			t.Errorf("ValidateAppName(%q) unexpected error: %v", name, err)
		}
	}
	for _, name := range []string{"", "UserApp", "user-app", "user_app", "1app", "user app"} {
		if err := ValidateAppName(name); err == nil {
			t.Errorf("ValidateAppName(%q) should error", name)
		}
	}
}

func TestValidateModulePath(t *testing.T) {
	for _, path := range []string{"github.com/acme/backend", "example.com/x/y", "myproject"} {
		if err := ValidateModulePath(path); err != nil {
			t.Errorf("ValidateModulePath(%q) unexpected error: %v", path, err)
		}
	}
	for _, path := range []string{"", "github.com/acme/backend/", "has space", "a/b/"} {
		if err := ValidateModulePath(path); err == nil {
			t.Errorf("ValidateModulePath(%q) should error", path)
		}
	}
}
