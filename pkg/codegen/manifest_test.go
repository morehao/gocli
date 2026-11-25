package codegen

import (
	"os"
	"path/filepath"
	"testing"

	golibcodegen "github.com/morehao/golib/codegen"
)

func fakeRes(table string, cols ...golibcodegen.ModelField) *golibcodegen.ModuleTplAnalysisRes {
	return &golibcodegen.ModuleTplAnalysisRes{
		TableName:  table,
		StructName: "Demo",
		TplAnalysisList: []golibcodegen.ModuleTplAnalysisItem{{
			TplAnalysisItem: golibcodegen.TplAnalysisItem{OriginLayerName: golibcodegen.LayerNameModel},
			ModelFields:     cols,
		}},
	}
}

func TestTableFingerprintStable(t *testing.T) {
	a := fakeRes("t", golibcodegen.ModelField{ColumnName: "id", FieldType: "uint", IsPrimaryKey: true})
	b := fakeRes("t", golibcodegen.ModelField{ColumnName: "id", FieldType: "uint", IsPrimaryKey: true})
	if tableFingerprint(a, "t") != tableFingerprint(b, "t") {
		t.Fatal("fingerprint should be stable")
	}
	c := fakeRes("t", golibcodegen.ModelField{ColumnName: "id", FieldType: "uint", IsPrimaryKey: true},
		golibcodegen.ModelField{ColumnName: "name", FieldType: "string"})
	if tableFingerprint(a, "t") == tableFingerprint(c, "t") {
		t.Fatal("fingerprint should change when columns change")
	}
}

func TestSkipIfUnchanged(t *testing.T) {
	root := t.TempDir()
	res := fakeRes("t", golibcodegen.ModelField{ColumnName: "id", FieldType: "uint", IsPrimaryKey: true})
	g := &Generator{opts: Options{ProjectRoot: root}, cfg: &Config{}}
	// 首次：不跳过
	if skip, _ := g.skipIfUnchanged(res, ModeFrontendPage); skip {
		t.Fatal("first time should not skip")
	}
	g.recordManifest(res, ModeFrontendPage)
	// 有记录后：跳过
	if skip, _ := g.skipIfUnchanged(res, ModeFrontendPage); !skip {
		t.Fatal("after record with same fingerprint should skip")
	}
	// force 不跳过
	g.opts.Force = true
	if skip, _ := g.skipIfUnchanged(res, ModeFrontendPage); skip {
		t.Fatal("force should not skip")
	}
	// 表结构变化：不跳过
	g.opts.Force = false
	res2 := fakeRes("t", golibcodegen.ModelField{ColumnName: "id", FieldType: "uint", IsPrimaryKey: true},
		golibcodegen.ModelField{ColumnName: "x", FieldType: "string"})
	if skip, _ := g.skipIfUnchanged(res2, ModeFrontendPage); skip {
		t.Fatal("changed schema should not skip")
	}
	_ = os.RemoveAll(root)
}

func TestUndoRestores(t *testing.T) {
	root := t.TempDir()
	dir := filepath.Join(root, "packages/types/src")
	os.MkdirAll(dir, 0o755)
	target := filepath.Join(dir, "index.ts")
	os.WriteFile(target, []byte("ORIGINAL"), 0o644)
	g := &Generator{opts: Options{ProjectRoot: root}}
	g.snapshot()
	os.WriteFile(target, []byte("MODIFIED"), 0o644)
	restored, err := g.Undo()
	if err != nil {
		t.Fatal(err)
	}
	if len(restored) == 0 {
		t.Fatal("expected restore")
	}
	b, _ := os.ReadFile(target)
	if string(b) != "ORIGINAL" {
		t.Fatalf("undo = %q, want ORIGINAL", b)
	}
}
