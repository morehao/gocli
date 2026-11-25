package webgen

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	tpl "github.com/morehao/gocli/template"
)

type fakeField struct {
	n, j, ftype, col, comment string
	nullable, pk              bool
}

func (f fakeField) FieldName() string  { return f.n }
func (f fakeField) JsonName() string   { return f.j }
func (f fakeField) FieldType() string  { return f.ftype }
func (f fakeField) ColumnType() string { return f.col }
func (f fakeField) Comment() string    { return f.comment }
func (f fakeField) IsNullable() bool   { return f.nullable }
func (f fakeField) IsPrimaryKey() bool { return f.pk }

func setupProject(t *testing.T) string {
	t.Helper()
	root := t.TempDir()
	for _, d := range []string{"packages/types/src", "packages/api/src", "apps/web/src"} {
		if err := os.MkdirAll(filepath.Join(root, d), 0o755); err != nil {
			t.Fatal(err)
		}
	}
	os.WriteFile(filepath.Join(root, "packages/types/src/index.ts"), []byte("export interface PageQuery { page?: number }\n"), 0o644)
	os.WriteFile(filepath.Join(root, "packages/api/src/index.ts"), []byte("export function request<T>(): Promise<T> { return null as T }\n"), 0o644)
	os.WriteFile(filepath.Join(root, "apps/web/src/App.tsx"), []byte(
		"// <codegen:imports>\nconst A = 1\n// <codegen:pages>\nconst P = {}\n// <codegen:menu>\nconst M = []\n"), 0o644)
	return root
}

func sampleMeta() *TableMeta {
	return FromModelFields("product", "Product", "product", "商品管理", []ModelFieldLike{
		fakeField{"ProductID", "productID", "uint", "bigserial", "自增ID", false, true},
		fakeField{"Name", "name", "string", "varchar(255)", "商品名称", false, false},
		fakeField{"Price", "price", "string", "numeric(10,2)", "价格", false, false},
		fakeField{"Status", "status", "int32", "integer", "状态", false, false},
		fakeField{"CreatedAt", "createdAt", "time.Time", "timestamptz", "创建时间", false, false},
	}, Overrides{})
}

func TestRenderProducesFiles(t *testing.T) {
	root := setupProject(t)
	tm := sampleMeta()
	if tm.PKJsonName != "productID" || tm.PKTSType != "number" || tm.ResourceKebab != "products" {
		t.Fatalf("meta wrong: %+v", tm)
	}
	r := New(tpl.GenerateFS, Options{ProjectRoot: root, RegisterMenu: true, MenuGroup: "数据管理", MenuLabel: "商品管理"})
	res, err := r.Render(tm)
	if err != nil {
		t.Fatalf("Render: %v", err)
	}
	if err := r.RegisterMenu(tm, "", "", res); err != nil {
		t.Fatalf("RegisterMenu: %v", err)
	}
	// types
	tb, _ := os.ReadFile(filepath.Join(root, "packages/types/src/index.ts"))
	if !strings.Contains(string(tb), "interface ProductPageListItem") || !strings.Contains(string(tb), "productID: number") {
		t.Fatalf("types missing: %s", tb)
	}
	if !strings.Contains(string(tb), "price?: number") {
		t.Fatalf("numeric price should be number: %s", tb)
	}
	// sys 字段不应出现
	if strings.Contains(string(tb), "createdAt") {
		t.Fatalf("sys field createdAt should be excluded from types")
	}
	// api
	ab, _ := os.ReadFile(filepath.Join(root, "packages/api/src/index.ts"))
	if !strings.Contains(string(ab), "function getProducts") || !strings.Contains(string(ab), "/product/products") {
		t.Fatalf("api missing: %s", ab)
	}
	// page
	pb, _ := os.ReadFile(filepath.Join(root, "apps/web/src/pages/Product/index.tsx"))
	if !strings.Contains(string(pb), "export default function ProductPage") {
		t.Fatalf("page missing component: %s", pb)
	}
	if strings.Contains(string(pb), "{{") {
		t.Fatalf("page has unrendered template syntax")
	}
	// App.tsx 菜单
	app, _ := os.ReadFile(filepath.Join(root, "apps/web/src/App.tsx"))
	if !strings.Contains(string(app), "import ProductPage from './pages/Product'") ||
		!strings.Contains(string(app), "key: 'products'") {
		t.Fatalf("menu not registered: %s", app)
	}
}

func TestIdempotent(t *testing.T) {
	root := setupProject(t)
	tm := sampleMeta()
	r := New(tpl.GenerateFS, Options{ProjectRoot: root, RegisterMenu: true})
	r.Render(tm)
	res2, _ := r.Render(tm)
	if err := r.RegisterMenu(tm, "", "", res2); err != nil {
		t.Fatalf("RegisterMenu2: %v", err)
	}
	tb, _ := os.ReadFile(filepath.Join(root, "packages/types/src/index.ts"))
	if strings.Count(string(tb), "interface ProductPageListItem") != 1 {
		t.Fatalf("types duplicated:\n%s", tb)
	}
	app, _ := os.ReadFile(filepath.Join(root, "apps/web/src/App.tsx"))
	if strings.Count(string(app), "import ProductPage") != 1 || strings.Count(string(app), "key: 'products'") != 1 {
		t.Fatalf("app duplicated:\n%s", app)
	}
}

func TestSelectOptionsAndSearch(t *testing.T) {
	root := setupProject(t)
	fields := []ModelFieldLike{
		fakeField{"ProductID", "productID", "uint", "bigserial", "自增ID", false, true},
		fakeField{"Name", "name", "string", "varchar(255)", "商品名称", false, false},
		fakeField{"Status", "status", "int32", "integer", "状态", false, false},
	}
	ov := Overrides{Options: map[string][]Option{
		"status": {{Value: "1", Label: "启用", Style: "number"}, {Value: "0", Label: "停用", Style: "number"}},
	}}
	tm := FromModelFields("product", "Product", "product", "商品", fields, ov)
	// status 应因 options 变为 select
	var statusWidget string
	for _, f := range tm.Fields {
		if f.JsonName == "status" {
			statusWidget = f.Widget
		}
	}
	if statusWidget != "select" {
		t.Fatalf("status widget = %q, want select", statusWidget)
	}
	r := New(tpl.GenerateFS, Options{ProjectRoot: root})
	if _, err := r.Render(tm); err != nil {
		t.Fatalf("Render: %v", err)
	}
	pb, _ := os.ReadFile(filepath.Join(root, "apps/web/src/pages/Product/index.tsx"))
	pg := string(pb)
	if !strings.Contains(pg, "SELECT_OPTIONS_Product_status") {
		t.Fatalf("missing select options const")
	}
	if !strings.Contains(pg, "label: '启用', value: 1") {
		t.Fatalf("missing option literal: %s", grepLine(pg, "启用"))
	}
	if !strings.Contains(pg, "toolbar={") || !strings.Contains(pg, "Input.Search") {
		t.Fatalf("missing search toolbar")
	}
}

func grepLine(s, sub string) string {
	for _, l := range strings.Split(s, "\n") {
		if strings.Contains(l, sub) {
			return l
		}
	}
	return "<not found: " + sub + ">"
}
