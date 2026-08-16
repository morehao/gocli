package generate

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	tplPkg "github.com/morehao/gocli/template"
)

// renderTpl 渲染指定模板文件并返回输出。
func renderTpl(t *testing.T, fsPath string, params map[string]any) string {
	t.Helper()
	tpl, err := template.New("").Funcs(template.FuncMap{
		TplFuncIsBuiltInField:      IsBuiltInField,
		TplFuncIsSysField:          IsSysField,
		TplFuncIsDefaultModelLayer: IsDefaultModelLayer,
		TplFuncIsDefaultDaoLayer:   IsDefaultDaoLayer,
		TplFuncHasTimeField:        HasTimeField,
		TplFuncGetFieldImports:     GetFieldImports,
		TplFuncIsBasicType:         IsBasicType,
		TplFuncToKebabCase:        toKebabCase,
		TplFuncPluralize:          pluralize,
		TplFuncIsNumID:            IsNumID,
		TplFuncIsStringID:         IsStringID,
		TplFuncHasTimeFieldAny:    HasTimeFieldAny,
	}).ParseFS(TemplatesFS, fsPath)
	if err != nil {
		t.Fatalf("parse %s: %v", fsPath, err)
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, filepath.Base(fsPath), params); err != nil {
		t.Fatalf("render %s: %v", fsPath, err)
	}
	return buf.String()
}

// TestDaoTplRenderGenericID 验证 golib v1.32.5 gormdao 适配：
// dao 模板中 Dao/NewDao 携带主键类型参数（如 uint），而 BaseCond 为非泛型纯数据容器。
func TestDaoTplRenderGenericID(t *testing.T) {
	fields := []ModelField{
		{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id", ColumnType: "bigint unsigned"},
		{FieldName: "Name", FieldType: "string", ColumnName: "name", ColumnType: "varchar(255)"},
	}
	params := map[string]any{
		"PKFieldType":    "uint",
		"PackageName":    "user",
		"DaoPackageName": "dao",
		"ModelLayerName": "model",
		"StructName":     "User",
		"BaseModulePath": "github.com/example",
		"AppModuleName":  "demoapp",
		"DBName":         "DemoDB",
		"ModelFields":    fields,
		"FieldImports":   []string{"time"},
	}
	for _, fsPath := range []string{"generate/module/dao.go.tpl", "generate/model/dao.go.tpl"} {
		out := renderTpl(t, fsPath, params)
		for _, want := range []string{
			"*gormdao.BaseCond",
			"*gormdao.Dao[model.UserEntity, model.UserEntityList, uint]",
			"gormdao.NewDao[model.UserEntity, model.UserEntityList, uint](",
		} {
			if !strings.Contains(out, want) {
				t.Errorf("%s missing %q\n---\n%s", fsPath, want, out)
			}
		}
	}
	// service 模板的 PageList 构造 Cond 时使用非泛型 BaseCond。
	if out := renderTpl(t, "generate/module/service.go.tpl", params); !strings.Contains(out, "&gormdao.BaseCond{") {
		t.Errorf("generate/module/service.go.tpl missing &gormdao.BaseCond{:\n%s", out)
	}
}

// TestModelTplRenderPrimaryKeyTag 验证 model 模板依据 IsPrimaryKey 生成 primaryKey 标签，
// 非主键字段不携带该标签。
func TestModelTplRenderPrimaryKeyTag(t *testing.T) {
	fields := []ModelField{
		{IsPrimaryKey: true, FieldName: "Code", FieldType: "string", ColumnName: "code", ColumnType: "varchar(64)", NullableDesc: "not null", DefaultValue: "default ''"},
		{FieldName: "Name", FieldType: "string", ColumnName: "name", ColumnType: "varchar(255)"},
	}
	params := map[string]any{
		"ModelLayerName": "model",
		"StructName":     "Dict",
		"ModelFields":    fields,
		"FieldImports":   []string{},
		"PKFieldType":    "string",
	}
	out := renderTpl(t, "generate/module/model.go.tpl", params)
	if !strings.Contains(out, "Code string `gorm:\"column:code;type:varchar(64);not null;default '';primaryKey\"`") {
		t.Errorf("model template missing primaryKey tag:\n%s", out)
	}
	if !strings.Contains(out, "Name string `gorm:\"column:name;type:varchar(255)\"`") {
		t.Errorf("non-pk field should not get primaryKey tag:\n%s", out)
	}
	if !strings.Contains(out, "map[string]DictEntity") {
		t.Errorf("model template should use PKFieldType in ToMap when pk is string:\n%s", out)
	}
	if strings.Contains(out, "gorm.Model") {
		t.Errorf("model template should not embed gorm.Model for non-numeric pk:\n%s", out)
	}
}

// TestMonorepoTplRenderGenericID 验证 monorepo 模板（create 命令产物）适配 gormdao。
func TestMonorepoTplRenderGenericID(t *testing.T) {
	// 渲染 dao/user.go.tmpl（直接模板，无参数）
	tpl, err := template.New("").ParseFS(tplPkg.ArkFS, "ark/apps/demo/dao/user.go.tmpl")
	if err != nil {
		t.Fatalf("parse monorepo dao template: %v", err)
	}
	var buf bytes.Buffer
	if err := tpl.ExecuteTemplate(&buf, "user.go.tmpl", nil); err != nil {
		t.Fatalf("render monorepo dao template: %v", err)
	}
	out := buf.String()
	for _, want := range []string{
		"*gormdao.BaseCond",
		"*gormdao.Dao[model.UserEntity, model.UserEntityList, uint]",
		"gormdao.NewDao[model.UserEntity, model.UserEntityList, uint](",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("monorepo dao template missing %q\n---\n%s", want, out)
		}
	}

	// 渲染 svcuser/user.go.tmpl
	tpl2, err := template.New("").ParseFS(tplPkg.ArkFS, "ark/apps/demo/internal/service/svcuser/user.go.tmpl")
	if err != nil {
		t.Fatalf("parse svcuser template: %v", err)
	}
	buf.Reset()
	if err := tpl2.ExecuteTemplate(&buf, "user.go.tmpl", nil); err != nil {
		t.Fatalf("render svcuser template: %v", err)
	}
	if !strings.Contains(buf.String(), "&gormdao.BaseCond{") {
		t.Errorf("svcuser template missing BaseCond:\n%s", buf.String())
	}
}

// TestStringPKServiceTpl 验证字符串主键时 service 模板生成正确的零值判断与 deletedBy 类型，
// 数值主键时保持原行为。
func TestStringPKServiceTpl(t *testing.T) {
	fields := []ModelField{
		{IsPrimaryKey: true, FieldName: "ID", FieldType: "string", ColumnName: "id", ColumnType: "varchar(36)", NullableDesc: "not null"},
		{FieldName: "Name", FieldType: "string", ColumnName: "name", ColumnType: "varchar(255)"},
	}
	base := map[string]any{
		"PackageName":          "coreconfig",
		"DaoPackageName":       "dao",
		"ModelLayerName":       "model",
		"StructName":           "CoreConfig",
		"StructNameLowerCamel": "coreConfig",
		"BaseModulePath":       "github.com/example",
		"AppModuleName":        "demoapp",
		"DBName":               "DemoDB",
		"ModelFields":          fields,
		"FieldImports":         []string{},
	}

	// 字符串主键
	strParams := map[string]any{}
	for k, v := range base {
		strParams[k] = v
	}
	strParams["PKFieldType"] = "string"
	strOut := renderTpl(t, "generate/module/service.go.tpl", strParams)
	for _, want := range []string{
		"coreConfigEntity.ID == \"\"",
		"deletedBy := gutil.ToString(userID)",
	} {
		if !strings.Contains(strOut, want) {
			t.Errorf("string-pk service missing %q\n---\n%s", want, strOut)
		}
	}
	if strings.Contains(strOut, "coreConfigEntity.ID == 0") || strings.Contains(strOut, "deletedBy := uint(") {
		t.Errorf("string-pk service should not use numeric zero/uint cast:\n%s", strOut)
	}

	// 数值主键（uint）保持原行为
	uintParams := map[string]any{}
	for k, v := range base {
		uintParams[k] = v
	}
	uintParams["PKFieldType"] = "uint"
	uintOut := renderTpl(t, "generate/module/service.go.tpl", uintParams)
	for _, want := range []string{
		"coreConfigEntity.ID == 0",
		"deletedBy := uint(gutil.VToUint64(userID))",
	} {
		if !strings.Contains(uintOut, want) {
			t.Errorf("uint-pk service missing %q\n---\n%s", want, uintOut)
		}
	}
}
