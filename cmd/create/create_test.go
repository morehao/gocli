package create

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestCreateProject 验证 create project 生成完整 monorepo 并正确重写模块路径。
func TestCreateProject(t *testing.T) {
	parent := t.TempDir()
	modulePath := "github.com/acme/backend"
	if err := createProject(filepath.Join(parent, "my-backend"), modulePath, false, false); err != nil {
		t.Fatalf("createProject: %v", err)
	}
	root := filepath.Join(parent, "my-backend")

	assertFileContent(t, filepath.Join(root, "go.work"), "use (")
	assertFileContent(t, filepath.Join(root, "pkg", "go.mod"), "module github.com/acme/backend/pkg")
	assertFileContent(t, filepath.Join(root, "apps", "demoapp", "go.mod"), "module github.com/acme/backend/demoapp")
	assertFileContent(t, filepath.Join(root, "apps", "demoapp", "main.go"), "github.com/acme/backend/demoapp/internal/router")
	assertFileContent(t, filepath.Join(root, "apps", "demoapp", "dao", "base.go"), "github.com/acme/backend/pkg/dbclient")

	// 不应残留占位符或模板后缀文件
	assertNoContent(t, root, "github.com/example")
	assertNoFile(t, root, "go.work.tmpl")
	assertNoFile(t, root, "go.mod.tmpl")
	assertNoFile(t, root, "main.go.tmpl")
}

// TestCreateApp 验证在既有 monorepo 中新增 app：内容替换 + go.work 注册。
func TestCreateApp(t *testing.T) {
	parent := t.TempDir()
	modulePath := "github.com/acme/backend"
	if err := createProject(filepath.Join(parent, "my-backend"), modulePath, false, false); err != nil {
		t.Fatalf("createProject: %v", err)
	}
	root := filepath.Join(parent, "my-backend")

	// 在项目根执行 create app
	wd, err := os.Getwd()
	if err != nil {
		t.Fatal(err)
	}
	if err := os.Chdir(root); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if err := createApp("userapp", "", false); err != nil {
		t.Fatalf("createApp: %v", err)
	}

	assertFileContent(t, filepath.Join(root, "apps", "userapp", "go.mod"), "module github.com/acme/backend/userapp")
	assertFileContent(t, filepath.Join(root, "apps", "userapp", "main.go"), "github.com/acme/backend/userapp/internal/router")
	assertFileContent(t, filepath.Join(root, "apps", "userapp", "main.go"), `"userapp"`)
	assertFileContent(t, filepath.Join(root, "apps", "userapp", "dao", "base.go"), "github.com/acme/backend/pkg/dbclient")
	assertFileContent(t, filepath.Join(root, "go.work"), "./apps/userapp")

	assertNoContent(t, filepath.Join(root, "apps", "userapp"), "github.com/example")
	assertNoContent(t, filepath.Join(root, "apps", "userapp"), "demoapp")
	assertNoFile(t, filepath.Join(root, "apps", "userapp"), "go.mod.tmpl")
	assertNoFile(t, filepath.Join(root, "apps", "userapp"), "main.go.tmpl")

	// 重复创建应报错
	if err := createApp("userapp", "", false); err == nil {
		t.Error("createApp on existing app should error")
	}
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
