package codegen

import (
	"os"
	"path/filepath"
	"testing"
)

// TestSDKImportable 验证外部可 import 的公开 SDK 契约：
//   - codegen.New 能构造出 *Generator
//   - GenerateBackendModule 在 db=nil 时返回明确错误而非 panic
func TestSDKImportable(t *testing.T) {
	// 构造一个最小的目标工程：ProjectRoot/apps/demoapp/config/code_gen.yaml + go.mod
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "demoapp")
	if err := os.MkdirAll(filepath.Join(appDir, "config"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "go.mod"),
		[]byte("module github.com/example/demoapp\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "config", "code_gen.yaml"),
		[]byte("service_name: demo\nmodule:\n  table_name: demo\n"), 0o644); err != nil {
		t.Fatalf("write code_gen.yaml: %v", err)
	}

	g, err := New(nil, Options{
		AppName:     "demoapp",
		ProjectRoot: root,
		Table:       "demo",
		Mode:        ModeBackendModule,
	})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if g == nil {
		t.Fatal("New returned nil Generator")
	}

	// db=nil 时 genModule 应在分析阶段返回明确错误（而非 panic）。
	_, err = g.GenerateBackendModule()
	if err == nil {
		t.Fatal("GenerateBackendModule with nil db should return error, got nil")
	}
	t.Logf("GenerateBackendModule(nil db) error (expected): %v", err)

	// 前端页面已实现：nil db 时应返回明确错误（而非 panic）。
	if _, err := g.GenerateFrontendPage(); err == nil {
		t.Fatal("GenerateFrontendPage with nil db should return error, got nil")
	} else {
		t.Logf("GenerateFrontendPage(nil db) error (expected): %v", err)
	}
	// 全栈已实现：nil db 时也应返回明确错误。
	if _, err := g.GenerateFullstack(); err == nil {
		t.Fatal("GenerateFullstack with nil db should return error, got nil")
	} else {
		t.Logf("GenerateFullstack(nil db) error (expected): %v", err)
	}
}

// TestGenerateDispatch 验证统一入口按 Mode 分发；未知 Mode 返回错误。
func TestGenerateDispatch(t *testing.T) {
	root := t.TempDir()
	appDir := filepath.Join(root, "apps", "demoapp")
	if err := os.MkdirAll(filepath.Join(appDir, "config"), 0o755); err != nil {
		t.Fatalf("mkdir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "go.mod"),
		[]byte("module github.com/example/demoapp\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "config", "code_gen.yaml"),
		[]byte("service_name: demo\n"), 0o644); err != nil {
		t.Fatalf("write code_gen.yaml: %v", err)
	}

	g, err := New(nil, Options{AppName: "demoapp", ProjectRoot: root, Mode: Mode("bogus")})
	if err != nil {
		t.Fatalf("New: %v", err)
	}
	if _, err := g.Generate(); err == nil {
		t.Fatal("Generate with unknown mode should return error")
	}
}
