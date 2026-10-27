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
		TplFuncToKebabCase:         toKebabCase,
		TplFuncPluralize:           pluralize,
		TplFuncIsNumID:             IsNumID,
		TplFuncIsStringID:          IsStringID,
		TplFuncIsIntID:             IsIntID,
		TplFuncHasTimeFieldAny:     HasTimeFieldAny,
		TplFuncHasSysField:         HasSysField,
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
		"PKFieldType":      "uint",
		"PackageName":      "user",
		"DaoPackageName":   "dao",
		"ModelLayerName":   "model",
		"StructName":       "User",
		"BaseModulePath":   "github.com/example",
		"AppModuleName":    "demoapp",
		"DBName":           "DemoDB",
		"ModelFields":      fields,
		"FieldImports":     []string{"time"},
		"NeedClauseImport": true,
	}
	for _, fsPath := range []string{"generate/module/dao.go.tpl", "generate/model/dao.go.tpl"} {
		out := renderTpl(t, fsPath, params)
		for _, want := range []string{
			"*gormdao.BaseCond",
			"*gormdao.Dao[model.UserEntity, model.UserEntityList, uint]",
			"gormdao.NewDao[model.UserEntity, model.UserEntityList, uint](",
			`clause.Eq{Column: clause.Column{Table: tableName, Name: "name"}, Value: c.Name}`,
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
		"ctx context.Context",
		"deletedBy := gutil.VToStr(req.Operator.UserID)",
	} {
		if !strings.Contains(strOut, want) {
			t.Errorf("string-pk service missing %q\n---\n%s", want, strOut)
		}
	}
	if strings.Contains(strOut, "coreConfigEntity.ID == 0") || strings.Contains(strOut, "deletedBy := uint(") {
		t.Errorf("string-pk service should not use numeric zero/uint cast:\n%s", strOut)
	}
	// service 不再依赖 gin：操作人由 controller 填进 DTO。
	if strings.Contains(strOut, "gincontext") || strings.Contains(strOut, "gin-gonic") || strings.Contains(strOut, "gin.Context") {
		t.Errorf("string-pk service should not reference gin:\n%s", strOut)
	}
	// DTO：Create/Update/Delete 都内嵌 operator.Operator（字段均 json:"-"，客户端不可伪造）。
	strReqOut := renderTpl(t, "generate/module/request.go.tpl", strParams)
	assertOperatorReqTpl(t, "string-pk", strReqOut)
	// controller 仍是 *gin.Context，调用处转纯 ctx，写操作前统一 FillOperator。
	strCtrOut := renderTpl(t, "generate/module/controller.go.tpl", strParams)
	assertOperatorCtrTpl(t, "string-pk", strCtrOut)

	// 数值主键（uint）保持原行为
	uintParams := map[string]any{}
	for k, v := range base {
		uintParams[k] = v
	}
	uintParams["PKFieldType"] = "uint"
	uintOut := renderTpl(t, "generate/module/service.go.tpl", uintParams)
	for _, want := range []string{
		"coreConfigEntity.ID == 0",
		"ctx context.Context",
		"deletedBy := req.Operator.UserID",
	} {
		if !strings.Contains(uintOut, want) {
			t.Errorf("uint-pk service missing %q\n---\n%s", want, uintOut)
		}
	}
	if strings.Contains(uintOut, "gincontext") || strings.Contains(uintOut, "gin.Context") {
		t.Errorf("uint-pk service should not reference gin:\n%s", uintOut)
	}
	uintReqOut := renderTpl(t, "generate/module/request.go.tpl", uintParams)
	assertOperatorReqTpl(t, "uint-pk", uintReqOut)
	uintCtrOut := renderTpl(t, "generate/module/controller.go.tpl", uintParams)
	assertOperatorCtrTpl(t, "uint-pk", uintCtrOut)
}

// assertOperatorReqTpl 断言 DTO 模板把操作人做成内嵌 operator.Operator（而非按主键类型分叉的标量字段），
// 且 Create / Update / Delete 三个写请求都带上它。
func assertOperatorReqTpl(t *testing.T, label, out string) {
	t.Helper()
	if !strings.Contains(out, `"github.com/example/pkg/operator"`) {
		t.Errorf("%s DTO should import pkg/operator:\n%s", label, out)
	}
	if got := strings.Count(out, "\toperator.Operator"); got != 3 {
		t.Errorf("%s DTO should embed operator.Operator in Create/Update/Delete, got %d:\n%s", label, got, out)
	}
	// Detail / PageList 是只读请求，不该携带身份。
	detail := out[strings.Index(out, "CoreConfigDetailReq"):strings.Index(out, "CoreConfigPageListReq")]
	if strings.Contains(detail, "operator.Operator") {
		t.Errorf("%s DetailReq should not embed operator.Operator:\n%s", label, detail)
	}
}

// assertOperatorCtrTpl 断言 controller 在 Create / Update / Delete 三处都用 ginhelper.FillOperator
// 填请求 DTO，并把纯 ctx 交给 service。
func assertOperatorCtrTpl(t *testing.T, label, out string) {
	t.Helper()
	for _, want := range []string{
		"Delete(ctx *gin.Context)",
		"ginhelper.FillOperator(ctx, &req.Operator)",
		"ctr.coreConfigSvc.Create(ctx.Request.Context(), &req)",
		"ctr.coreConfigSvc.Delete(ctx.Request.Context(), &req)",
		"ctr.coreConfigSvc.Update(ctx.Request.Context(), &req)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("%s controller missing %q\n---\n%s", label, want, out)
		}
	}
	if got := strings.Count(out, "ginhelper.FillOperator(ctx, &req.Operator)"); got != 3 {
		t.Errorf("%s controller should FillOperator in Create/Update/Delete, got %d", label, got)
	}
	// 身份读取收敛到 ginhelper：controller 不再自己调 getter。
	if strings.Contains(out, "gincontext.GetUserID") {
		t.Errorf("%s controller should not call gincontext.GetUserID directly:\n%s", label, out)
	}
}

// TestIntPKServiceTpl 验证有符号整型主键下 deletedBy 仍按 PK 类型转换（dao.Delete 形参为主键类型），
// 操作人同样取自 DTO。
func TestIntPKServiceTpl(t *testing.T) {
	params := map[string]any{
		"PKFieldType":          "int64",
		"PackageName":          "coreconfig",
		"DaoPackageName":       "dao",
		"ModelLayerName":       "model",
		"StructName":           "CoreConfig",
		"StructNameLowerCamel": "coreConfig",
		"BaseModulePath":       "github.com/example",
		"AppModuleName":        "demoapp",
		"DBName":               "DemoDB",
		"ModelFields": []ModelField{
			{IsPrimaryKey: true, FieldName: "ID", FieldType: "int64", ColumnName: "id", ColumnType: "bigint"},
		},
		"FieldImports": []string{},
	}
	out := renderTpl(t, "generate/module/service.go.tpl", params)
	for _, want := range []string{
		"ctx context.Context",
		"deletedBy := int64(req.Operator.UserID)",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("int64-pk service missing %q\n---\n%s", want, out)
		}
	}
	if strings.Contains(out, "gincontext") || strings.Contains(out, "gin.Context") {
		t.Errorf("int64-pk service should not reference gin:\n%s", out)
	}
}

// TestServiceTplAuditColumns 验证 service 模板按表里是否真的有审计列决定是否写 CreatedBy / UpdatedBy：
// 表没有这些列时不能生成赋值，否则生成物编译失败。
func TestServiceTplAuditColumns(t *testing.T) {
	params := func(fields []ModelField) map[string]any {
		return map[string]any{
			"PKFieldType":          "uint",
			"PackageName":          "user",
			"DaoPackageName":       "dao",
			"ModelLayerName":       "model",
			"StructName":           "User",
			"StructNameLowerCamel": "user",
			"BaseModulePath":       "github.com/example",
			"AppModuleName":        "demoapp",
			"DBName":               "DemoDB",
			"ModelFields":          fields,
			"FieldImports":         []string{},
		}
	}
	pk := ModelField{IsPrimaryKey: true, FieldName: "ID", FieldType: "uint", ColumnName: "id"}

	// 有审计列：Create 写 created_by/updated_by，Update 写 updated_by，详情/列表回显。
	withAudit := renderTpl(t, "generate/module/service.go.tpl", params([]ModelField{
		pk,
		{FieldName: "Name", FieldType: "string", ColumnName: "name"},
		{FieldName: "CreatedBy", FieldType: "uint", ColumnName: "created_by"},
		{FieldName: "UpdatedBy", FieldType: "uint", ColumnName: "updated_by"},
	}))
	for _, want := range []string{
		"CreatedBy: req.Operator.UserID,",
		"UpdatedBy: req.Operator.UserID,",
		`updateMap["updated_by"] = req.Operator.UserID`,
		"CreatedBy: userEntity.CreatedBy,",
		"UpdatedBy: userEntity.UpdatedBy,",
		"CreatedBy: v.CreatedBy,",
		"UpdatedBy: v.UpdatedBy,",
	} {
		if !strings.Contains(withAudit, want) {
			t.Errorf("service with audit columns missing %q\n---\n%s", want, withAudit)
		}
	}

	// 无审计列：不得出现任何审计人赋值。
	withoutAudit := renderTpl(t, "generate/module/service.go.tpl", params([]ModelField{
		pk,
		{FieldName: "Name", FieldType: "string", ColumnName: "name"},
	}))
	for _, forbidden := range []string{"CreatedBy:", "UpdatedBy:", `updateMap["updated_by"]`} {
		if strings.Contains(withoutAudit, forbidden) {
			t.Errorf("service without audit columns should not write %q:\n%s", forbidden, withoutAudit)
		}
	}
}

// TestApiTplServiceContext 验证 api 生成器的 service 收 context.Context，
// controller 保持 *gin.Context 并在调用处转换。
func TestApiTplServiceContext(t *testing.T) {
	params := map[string]any{
		"PackageName":            "user",
		"StructName":             "User",
		"StructNameLowerCamel":   "user",
		"FunctionName":           "Login",
		"FunctionNameLowerCamel": "login",
		"HttpMethod":             "GET",
		"BaseModulePath":         "github.com/example",
		"AppModuleName":          "demoapp",
		"AppName":                "demo",
		"ApiDocTag":              "用户",
		"TargetFileExist":        false,
	}
	svcOut := renderTpl(t, "generate/api/service.go.tpl", params)
	if !strings.Contains(svcOut, "Login(ctx context.Context, req *dtouser.UserLoginReq)") {
		t.Errorf("api service should take context.Context:\n%s", svcOut)
	}
	if strings.Contains(svcOut, "gin.Context") || strings.Contains(svcOut, "gin-gonic") {
		t.Errorf("api service should not reference gin:\n%s", svcOut)
	}
	ctrOut := renderTpl(t, "generate/api/controller.go.tpl", params)
	for _, want := range []string{
		"Login(ctx *gin.Context)",
		"ctr.userSvc.Login(ctx.Request.Context(), &req)",
	} {
		if !strings.Contains(ctrOut, want) {
			t.Errorf("api controller missing %q\n---\n%s", want, ctrOut)
		}
	}
}
