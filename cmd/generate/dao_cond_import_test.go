package generate

import (
	"go/format"
	"strings"
	"testing"
)

// TestHasDaoCondBranch 钉住 dao 层 clause import 的有无：
// 没有条件分支却 import 了 clause，生成物会因 import 未使用而编译失败。
func TestHasDaoCondBranch(t *testing.T) {
	basic := ModelField{FieldName: "Name", FieldType: "string", ColumnName: "name"}
	nonBasic := ModelField{FieldName: "Extra", FieldType: "datatypes.JSON", ColumnName: "extra"}

	cases := []struct {
		name   string
		fields []ModelField
		want   bool
	}{
		{
			name:   "只有内建字段时不生成条件分支",
			fields: []ModelField{{FieldName: "ID", FieldType: "uint"}, {FieldName: "CreatedAt", FieldType: "time.Time"}},
			want:   false,
		},
		{
			name:   "只有非基本类型字段时不生成条件分支",
			fields: []ModelField{nonBasic},
			want:   false,
		},
		{
			name:   "有基本类型字段时生成条件分支",
			fields: []ModelField{{FieldName: "ID", FieldType: "uint"}, basic},
			want:   true,
		},
		{
			name:   "内建字段之后的非基本类型字段不算条件分支",
			fields: []ModelField{{FieldName: "ID", FieldType: "uint"}, nonBasic},
			want:   false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			if got := hasDaoCondBranch(c.fields); got != c.want {
				t.Errorf("hasDaoCondBranch() = %v, want %v", got, c.want)
			}
		})
	}
}

// importGroups 从渲染出的源码里提取 import 分组：import 块内以空行分隔的连续块。
// 先把源码过一遍 format.Source（与 codegen 落盘前的处理一致），因此拿到的是规范形态，
// 顺带也就校验了模板渲染出的确实是合法 Go 源码。
func importGroups(t *testing.T, src string) [][]string {
	t.Helper()

	formatted, err := format.Source([]byte(src))
	if err != nil {
		t.Fatalf("rendered template is not valid Go source: %v\n---\n%s", err, src)
	}
	text := string(formatted)

	start := strings.Index(text, "import (")
	if start < 0 {
		t.Fatalf("no import block in rendered source:\n%s", text)
	}
	rest := text[start+len("import ("):]
	end := strings.Index(rest, "\n)")
	if end < 0 {
		t.Fatalf("unterminated import block:\n%s", text)
	}

	var groups [][]string
	var current []string
	for _, line := range strings.Split(rest[:end], "\n") {
		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			if len(current) > 0 {
				groups = append(groups, current)
				current = nil
			}
			continue
		}
		current = append(current, strings.Trim(trimmed, `"`))
	}
	if len(current) > 0 {
		groups = append(groups, current)
	}
	return groups
}

// groupOf 返回 path 所在的分组下标，不存在时返回 -1。
func groupOf(groups [][]string, path string) int {
	for i, group := range groups {
		for _, item := range group {
			if item == path {
				return i
			}
		}
	}
	return -1
}

// TestDaoTplClauseImportMatchesUsage 验证 dao 模板的 clause import 满足两点：
//  1. 与实际使用严格同步——import 了却没有条件分支会因未使用而编译失败，漏 import 则直接编译不过；
//  2. 落在第三方 import 分组，而不是混进 FieldImports 所在的标准库分组。
func TestDaoTplClauseImportMatchesUsage(t *testing.T) {
	const clausePath = "gorm.io/gorm/clause"

	cases := []struct {
		name       string
		fields     []ModelField
		wantClause bool
		wantStdlib string // 期望出现在 FieldImports 分组里的标准库 import，空表示没有
	}{
		{
			name: "含基本类型字段",
			fields: []ModelField{
				{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id", ColumnType: "bigint unsigned"},
				{FieldName: "Name", FieldType: "string", ColumnName: "name", ColumnType: "varchar(255)"},
			},
			wantClause: true,
		},
		{
			name: "基本类型字段与时间字段并存",
			fields: []ModelField{
				{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id", ColumnType: "bigint unsigned"},
				{FieldName: "Name", FieldType: "string", ColumnName: "name", ColumnType: "varchar(255)"},
				{FieldName: "LoginTime", FieldType: "time.Time", ColumnName: "login_time"},
			},
			wantClause: true,
			wantStdlib: "time",
		},
		{
			name: "只有内建字段",
			fields: []ModelField{
				{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id", ColumnType: "bigint unsigned"},
				{FieldName: "CreatedAt", FieldType: "time.Time", ColumnName: "created_at"},
				{FieldName: "UpdatedAt", FieldType: "time.Time", ColumnName: "updated_at"},
				{FieldName: "DeletedAt", FieldType: "time.Time", ColumnName: "deleted_at"},
			},
			wantClause: false,
		},
		{
			name: "除内建字段外只有非基本类型字段",
			fields: []ModelField{
				{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id", ColumnType: "bigint unsigned"},
				{FieldName: "Extra", FieldType: "datatypes.JSON", ColumnName: "extra"},
			},
			wantClause: false,
		},
	}

	for _, c := range cases {
		t.Run(c.name, func(t *testing.T) {
			params := map[string]any{
				"PKFieldType":      "uint",
				"PackageName":      "user",
				"DaoPackageName":   "dao",
				"ModelLayerName":   "model",
				"StructName":       "User",
				"BaseModulePath":   "github.com/example",
				"AppModuleName":    "demoapp",
				"DBName":           "DemoDB",
				"ModelFields":      c.fields,
				"FieldImports":     calcFieldImports(c.fields),
				"NeedClauseImport": hasDaoCondBranch(c.fields),
			}

			for _, fsPath := range []string{"generate/module/dao.go.tpl", "generate/model/dao.go.tpl"} {
				out := renderTpl(t, fsPath, params)
				imported := strings.Contains(out, `"`+clausePath+`"`)
				used := strings.Contains(out, "clause.Eq{")
				if imported != used {
					t.Errorf("%s: import clause=%v 但使用=%v，生成物会编译失败\n---\n%s", fsPath, imported, used, out)
				}
				if imported != c.wantClause {
					t.Errorf("%s: 期望 clause=%v，实际=%v\n---\n%s", fsPath, c.wantClause, imported, out)
				}

				groups := importGroups(t, out)
				clauseGroup := groupOf(groups, clausePath)
				if c.wantClause && clauseGroup < 0 {
					t.Fatalf("%s: 期望 import clause，但分组里没有找到\n---\n%s", fsPath, out)
				}
				if !c.wantClause && clauseGroup >= 0 {
					t.Fatalf("%s: 不该 import clause，却出现在第 %d 组\n---\n%s", fsPath, clauseGroup, out)
				}
				// clause 必须与标准库分开分组：与 time 同组就说明又混回了 FieldImports。
				if c.wantStdlib != "" {
					stdlibGroup := groupOf(groups, c.wantStdlib)
					if stdlibGroup < 0 {
						t.Fatalf("%s: 期望 FieldImports 里有 %q\n---\n%s", fsPath, c.wantStdlib, out)
					}
					if clauseGroup == stdlibGroup {
						t.Errorf("%s: clause 与标准库 %q 同在第 %d 组，应放进第三方分组\n---\n%s",
							fsPath, c.wantStdlib, clauseGroup, out)
					}
				}
			}
		})
	}
}
