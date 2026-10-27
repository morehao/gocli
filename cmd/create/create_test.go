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

	assertFileContent(t, filepath.Join(root, "go.work"), "./backend/apps/demo")
	assertFileContent(t, filepath.Join(root, "go.work"), "./backend/pkg")
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "go.mod"), "module github.com/acme/backend/pkg")
	assertFileContent(t, filepath.Join(root, "backend", "apps", "demo", "go.mod"), "module github.com/acme/backend/demo")
	// backend/ 内不再有嵌套 go.work
	assertNoFile(t, filepath.Join(root, "backend"), "go.work")

	// 不应残留占位符或模板后缀文件
	assertNoContent(t, root, "github.com/morehao/go-ark-template")

	// 生成物的 service 层不依赖 gin：service 收 context.Context，操作人由 controller 填进 DTO。
	backend := filepath.Join(root, "backend")
	assertNoServiceGin(t, filepath.Join(backend, "apps", "demo", "internal", "service"))
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "internal", "service", "svcuser", "user.go"),
		"func (svc *userSvc) PageList(ctx context.Context, req *dtouser.UserPageListReq)")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "internal", "controller", "ctruser", "user.go"),
		"ctr.userSvc.Delete(ctx.Request.Context(), &req)")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "internal", "controller", "ctruser", "user.go"),
		"ginhelper.FillOperator(ctx, &req.Operator)")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "internal", "dto", "dtouser", "request.go"),
		"operator.Operator")
	// 操作人定义在 pkg/operator（不依赖 gin），由 pkg/ginhelper 从登录态填充。
	assertFileContent(t, filepath.Join(backend, "pkg", "operator", "operator.go"), "package operator")
	assertFileContent(t, filepath.Join(backend, "pkg", "operator", "operator.go"), "UserID   uint")
	assertFileContent(t, filepath.Join(backend, "pkg", "operator", "operator.go"), "`json:\"-\" form:\"-\"`")
	assertFileContent(t, filepath.Join(backend, "pkg", "ginhelper", "operator.go"), "func FillOperator(ctx *gin.Context, op *operator.Operator)")
	assertNoFile(t, filepath.Join(backend, "pkg", "operator"), "operator.go.tmpl")

	// 示例库为 PostgreSQL：config 与 code_gen 都用 postgres DSN，model tag 用 PG 类型。
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "config", "config.yaml"),
		"postgres://postgres:123456@127.0.0.1:5432/demo")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "config", "code_gen.yaml"),
		"postgresql://postgres:123456@127.0.0.1:5432/demo")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "model", "user.go"), "type:bigint;not null;default:0")
	// 只校验 gorm tag 行本身：注释/守护用例里对 `type:bigint unsigned` 反例的说明是允许的。
	assertNoGormUnsignedTag(t, filepath.Join(backend, "apps", "demo", "model"))
	// health 只检查 postgres/redis；建表脚本是 PG 语法
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "internal", "dto", "dtohealth", "health.go"), `json:"postgres"`)
	assertFileContent(t, filepath.Join(backend, "scripts", "sql", "table_init.sql"), "BIGSERIAL")
	assertNoContent(t, filepath.Join(backend, "scripts", "sql"), "ENGINE=InnoDB")

	// 依赖 pin 到 golib v1.33.0（gincontext 身份 getter + gcontext 包注释）。
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "go.mod"), "github.com/morehao/golib v1.33.0")
	assertFileContent(t, filepath.Join(backend, "apps", "demo", "go.sum"),
		"github.com/morehao/golib v1.33.0 h1:JGNNdOxDL2q8Y2S0RHIKGBKZE1aXHVdFLvztt5FU3v4=")
	assertFileContent(t, filepath.Join(backend, "pkg", "go.mod"), "github.com/morehao/golib v1.33.0")
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
	// 根 go.work 是唯一 workspace，指向 backend 下子模块
	assertFileContent(t, filepath.Join(root, "go.work"), "./backend/apps/demo")
	assertFileContent(t, filepath.Join(root, "go.work"), "./backend/pkg")
	// backend/ 内不再有嵌套 go.work
	assertNoFile(t, filepath.Join(root, "backend"), "go.work")
}

// TestCreateAppInvalidName 非法 app 名应被拒绝。
func TestCreateAppInvalidName(t *testing.T) {
	if err := createAppX("User-App", "", false); err == nil {
		t.Error("createAppX with invalid name should error")
	}
}

// TestCreateAppFromArk 验证 create app 从 backend/apps/demo 复制并做 token 替换 + pkg 联动。
func TestCreateAppFromArk(t *testing.T) {
	parent := t.TempDir()
	if err := createProjectWithOpts(CreateOptions{
		Dir:         filepath.Join(parent, "my-ark"),
		ModulePath:  "github.com/acme/my-ark",
		ProjectName: "my-ark",
		GitInit:     false,
		Tidy:        false,
	}); err != nil {
		t.Fatalf("createProjectWithOpts: %v", err)
	}
	root := filepath.Join(parent, "my-ark")
	wd, _ := os.Getwd()
	if err := os.Chdir(filepath.Join(root, "backend")); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { _ = os.Chdir(wd) })

	if err := createAppX("user", "", false); err != nil {
		t.Fatalf("createAppX: %v", err)
	}
	assertFileContent(t, filepath.Join(root, "go.work"), "./backend/apps/user")
	assertFileContent(t, filepath.Join(root, "backend", "apps", "user", "go.mod"), "module github.com/acme/my-ark/user")
	assertNoContent(t, filepath.Join(root, "backend", "apps", "user"), "demo")
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "testsetup", "constant.go"), "AppNameUser")
	// 编译关键：新 app 引用的 dbclient 连接件必须补齐（UserDB / UserappInitializer / AppNameUser）
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "dbclient", "gorm.go"), "func UserDB")
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "dbclient", "gorm.go"), "dbNameUser")
	assertFileContent(t, filepath.Join(root, "backend", "pkg", "testsetup", "init.go"), "newUserappInitializer")
	// 本项目不连接 Elasticsearch：新 app 不再需要 ES 连接件，pkg 里也不应有 es.go
	assertNoFile(t, filepath.Join(root, "backend", "pkg", "dbclient"), "es.go")
	assertNoContent(t, filepath.Join(root, "backend"), "elasticsearch")
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

// assertNoServiceGin 断言 service 目录下没有任何 gin 依赖（import、*gin.Context 或 gincontext 调用）。
func assertNoServiceGin(t *testing.T, dir string) {
	t.Helper()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
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
		for _, forbidden := range []string{"gin-gonic", "gin.Context", "gincontext"} {
			if strings.Contains(string(content), forbidden) {
				t.Errorf("%s contains forbidden %q; service must take context.Context and read operator from DTO", path, forbidden)
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}

// assertNoGormUnsignedTag 断言目录下没有 gorm tag 使用 MySQL 的 unsigned 字面量。
func assertNoGormUnsignedTag(t *testing.T, dir string) {
	t.Helper()
	err := filepath.Walk(dir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}
		if info.IsDir() {
			return nil
		}
		for _, line := range strings.Split(string(mustRead(t, path)), "\n") {
			if strings.Contains(line, `gorm:"`) && strings.Contains(line, "unsigned") {
				t.Errorf("%s has MySQL-only gorm tag: %s", path, strings.TrimSpace(line))
			}
		}
		return nil
	})
	if err != nil {
		t.Fatalf("walk %s: %v", dir, err)
	}
}

// mustRead 读取文件，失败即终止用例。
func mustRead(t *testing.T, path string) []byte {
	t.Helper()
	content, err := os.ReadFile(path)
	if err != nil {
		t.Fatalf("read %s: %v", path, err)
	}
	return content
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
