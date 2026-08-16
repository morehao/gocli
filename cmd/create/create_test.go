package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateProject 验证 create project 基于 ark 模板生成仓库根与 backend，并正确重写模块路径。
func TestCreateProject(t *testing.T) {
	parent := t.TempDir()
	modulePath := "github.com/acme/backend"
	if err := createProject(filepath.Join(parent, "my-backend"), modulePath, false, false); err != nil {
		t.Fatalf("createProject: %v", err)
	}
	root := filepath.Join(parent, "my-backend")

	assertFileContent(t, filepath.Join(root, "go.work"), "./backend")
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "go.mod"), "module github.com/acme/backend/pkg")
	assertFileContent(t, filepath.Join(root, "backend", "apps", "demo", "go.mod"), "module github.com/acme/backend/demo")
	assertFileContent(t, filepath.Join(root, "backend", "go.work"), "./apps/demo")
	assertFileContent(t, filepath.Join(root, "backend", "go.work"), "./pkg")

	// 不应残留占位符或模板后缀文件
	assertNoContent(t, root, "github.com/morehao/go-ark-template")
	assertNoFile(t, root, "go.work.tmpl")
	assertNoFile(t, root, "go.mod.tmpl")
	assertNoFile(t, root, "go.work.sum.tmpl")
	assertNoFile(t, root, "main.go.tmpl")
}

// TestCreateProjectProjectName 验证给定项目名时生成仓库根与 ark backend。
func TestCreateProjectProjectName(t *testing.T) {
	parent := t.TempDir()
	if err := createProjectWithOpts(CreateOptions{
		Dir:         filepath.Join(parent, "my-ark"),
		ModulePath:  "github.com/acme/my-ark",
		ProjectName: "my-ark",
	}); err != nil {
		t.Fatalf("createProjectWithOpts: %v", err)
	}
	root := filepath.Join(parent, "my-ark")
	assertFileContent(t, filepath.Join(root, "go.mod"), "module github.com/acme/my-ark")
	assertFileContent(t, filepath.Join(root, "go.work"), "./backend")
	assertFileContent(t, filepath.Join(root, "backend", "go.work"), "./apps/demo")
	assertFileContent(t, filepath.Join(root, "backend", "go.work"), "./pkg")
}

// TestCreateAppInvalidName 非法 app 名应被拒绝。
func TestCreateAppInvalidName(t *testing.T) {
	if err := createApp("User-App", "", false); err == nil {
		t.Error("createApp with invalid name should error")
	}
}

func assertFileContent(t *testing.T, path, want string) {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	if !strings.Contains(string(content), want) {
		t.Errorf("%s missing %q, got:\n%s", path, want, content)
	}
}

func assertNoContent(t *testing.T, root, forbidden string) {
	t.Helper()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		content, err := os.ReadFile(path)
		if err != nil {
			return err
		}
		if strings.Contains(string(content), forbidden) {
			t.Errorf("%s contains forbidden %q", path, forbidden)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}

func assertNoFile(t *testing.T, root, name string) {
	t.Helper()
	err := filepath.Walk(root, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if !info.IsDir() && info.Name() == name {
			t.Errorf("unexpected file found: %s", path)
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", root, err)
	}
}
