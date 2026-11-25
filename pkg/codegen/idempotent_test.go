package codegen

import (
	"os"
	"path/filepath"
	"testing"

	golibcodegen "github.com/morehao/golib/codegen"
)

func TestAddContentToFuncIdempotent(t *testing.T) {
	dir := t.TempDir()
	f := filepath.Join(dir, "x.go")
	content := `package x

func init() {
	registerError(demoErrorMsgMap)
}
`
	if err := os.WriteFile(f, []byte(content), 0o644); err != nil {
		t.Fatal(err)
	}
	// 已含相同内容：不应重复追加
	if err := addContentToFuncIdempotent(f, "init", "registerError(demoErrorMsgMap)"); err != nil {
		t.Fatal(err)
	}
	data, _ := os.ReadFile(f)
	if got := countSub(string(data), "registerError(demoErrorMsgMap)"); got != 1 {
		t.Fatalf("content count = %d, want 1", got)
	}
}

func TestRemoveExclusiveTargets(t *testing.T) {
	dir := t.TempDir()
	os.WriteFile(filepath.Join(dir, "model.go"), []byte("x"), 0o644)
	os.WriteFile(filepath.Join(dir, "router.go"), []byte("x"), 0o644) // 共享，应保留
	items := []golibcodegen.GenParamsItem{
		{TargetDir: dir, TargetFileName: "model.go"},
		{TargetDir: dir, TargetFileName: "router.go"},
	}
	removeExclusiveTargets(items)
	if _, err := os.Stat(filepath.Join(dir, "model.go")); !os.IsNotExist(err) {
		t.Fatal("exclusive file model.go should be removed")
	}
	if _, err := os.Stat(filepath.Join(dir, "router.go")); err != nil {
		t.Fatal("shared file router.go should be kept")
	}
}

func countSub(s, sub string) int {
	n := 0
	for i := 0; i+len(sub) <= len(s); i++ {
		if s[i:i+len(sub)] == sub {
			n++
		}
	}
	return n
}
