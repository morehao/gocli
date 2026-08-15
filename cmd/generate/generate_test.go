package generate

import (
	"bytes"
	"io"
	"io/fs"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"text/template"

	"github.com/morehao/gocli/internal/scaffold"
	"github.com/morehao/golib/gutil"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// 测试模板文件加载
func TestLoadTemplates(t *testing.T) {
	// 测试模板目录是否存在
	dirs := []string{"generate/module", "generate/model", "generate/api"}
	for _, dir := range dirs {
		entries, err := TemplatesFS.ReadDir(dir)
		if err != nil {
			t.Errorf("Failed to read directory %s: %v", dir, err)
			continue
		}
		if len(entries) == 0 {
			t.Errorf("Directory %s is empty", dir)
		}
		t.Logf("Directory %s is not empty", dir)
	}
}

// copyDir 递归复制目录，用于在临时目录中执行代码生成测试，
// 避免生成逻辑污染仓库内的示例项目。
func copyDir(t *testing.T, srcDir, dstDir string) {
	t.Helper()
	err := fs.WalkDir(os.DirFS(srcDir), ".", func(path string, d fs.DirEntry, err error) error {
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dstDir, path)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		data, readErr := os.ReadFile(filepath.Join(srcDir, path))
		if readErr != nil {
			return readErr
		}
		return os.WriteFile(targetPath, data, 0o644)
	})
	if err != nil {
		t.Fatalf("copy example dir: %v", err)
	}
}

// monorepoTemplateDir 返回 monorepo 模板目录（原 cmd/generate/example，已迁移至 template/monorepo）
func monorepoTemplateDir() string {
	abs, err := filepath.Abs(filepath.Join("..", "..", "template", "monorepo"))
	if err != nil {
		return ""
	}
	return abs
}

// chdirToExample 切换到 monorepo 模板目录（优先使用临时副本，避免生成代码污染仓库）
func chdirToExample(t *testing.T) func() {
	t.Helper()
	exampleDir := monorepoTemplateDir()
	if exampleDir == "" {
		t.Fatalf("resolve monorepo template dir fail")
	}
	originalDir, err := os.Getwd()
	if err != nil {
		t.Fatalf("getwd: %v", err)
	}
	if _, err := os.Stat(exampleDir); err != nil {
		t.Skipf("Skipping test: example directory not found: %v", err)
	}

	tmpExampleDir := filepath.Join(t.TempDir(), "example")
	copyDir(t, exampleDir, tmpExampleDir)

	// 模板中以 .tmpl 后缀存放 go.mod/go.sum/go.work，恢复为标准文件名
	if err := scaffold.RestoreTemplateFiles(tmpExampleDir); err != nil {
		t.Fatalf("restore template module files: %v", err)
	}

	if err := os.Chdir(tmpExampleDir); err != nil {
		t.Fatalf("chdir to example copy: %v", err)
	}
	return func() {
		if err := os.Chdir(originalDir); err != nil {
			t.Fatalf("restore working directory: %v", err)
		}
	}
}

// resetGenerateState 清空包级状态，避免不同测试的临时目录之间相互污染
func resetGenerateState() {
	cfg = nil
	workDir = ""
	DBClient = nil
}

// skipIfDBUnavailable 数据库不可达时跳过测试
func skipIfDBUnavailable(t *testing.T) {
	t.Helper()
	exampleDir := monorepoTemplateDir()
	configFilepath := filepath.Join(exampleDir, "apps", "demoapp", "config", "code_gen.yaml")
	if _, err := os.Stat(configFilepath); err != nil {
		t.Skipf("Skipping test: config file not found: %v", err)
		return
	}
	var localCfg Config
	gutil.LoadYamlConfig(configFilepath, &localCfg)
	dsn := localCfg.DatabaseDSN
	if dsn == "" {
		t.Skip("no database dsn configured")
	}
	dbCfg, err := ParseDatabaseDSN(dsn)
	if err != nil {
		t.Skipf("invalid database dsn: %v", err)
	}
	var db *gorm.DB
	switch dbCfg.Type {
	case DBTypeMySQL:
		db, err = gorm.Open(mysql.Open(dbCfg.ConnStr), &gorm.Config{})
	case DBTypePostgres:
		db, err = gorm.Open(postgres.Open(dbCfg.ConnStr), &gorm.Config{})
	default:
		t.Skipf("unsupported database type: %s", dbCfg.Type)
	}
	if err != nil {
		t.Skipf("Skipping test: database unavailable: %v", err)
	}
	if sqlDB, e := db.DB(); e == nil {
		_ = sqlDB.Close()
	}
}

// captureStdout 捕获执行期间写入 os.Stdout 的输出
func captureStdout(t *testing.T, fn func()) string {
	t.Helper()
	oldStdout := os.Stdout
	r, w, err := os.Pipe()
	if err != nil {
		t.Fatalf("create pipe: %v", err)
	}
	os.Stdout = w
	defer func() { os.Stdout = oldStdout }()

	fn()

	_ = w.Close()
	out, err := io.ReadAll(r)
	if err != nil {
		t.Fatalf("read stdout: %v", err)
	}
	_ = r.Close()
	return string(out)
}

// assertGenerateSuccess 断言生成命令成功：输出包含 "Generated successfully" 且不含 "Error generating"
func assertGenerateSuccess(t *testing.T, output string) {
	t.Helper()
	if strings.Contains(output, "Error generating") {
		t.Fatalf("generation failed:\n%s", output)
	}
	if !strings.Contains(output, "Generated successfully") {
		t.Fatalf("generation did not complete successfully:\n%s", output)
	}
}

// 测试配置加载
func TestConfigLoading(t *testing.T) {
	resetGenerateState()
	restore := chdirToExample(t)
	defer restore()

	// 执行命令，需要指定 app 名称
	_, err := ExecuteCommand(Cmd, "model", "--app", "demoapp")
	if err != nil {
		t.Errorf("Failed to execute command with config: %v", err)
	}
	if cfg == nil {
		t.Fatal("config was not loaded")
	}
	if cfg.ServiceName != "mysql" {
		t.Errorf("ServiceName = %q, want %q", cfg.ServiceName, "mysql")
	}
	if cfg.Module.TableName != "user" {
		t.Errorf("Module.TableName = %q, want %q", cfg.Module.TableName, "user")
	}
	if cfg.Model.TableName != "user_login_log" {
		t.Errorf("Model.TableName = %q, want %q", cfg.Model.TableName, "user_login_log")
	}
	t.Log(gutil.ToJsonString(cfg))
}

// TestGenerateModelCode 测试生成 model 层代码
// 依赖数据库（example/apps/demoapp/config/code_gen.yaml 中的 DSN），数据库不可达时自动跳过。
func TestGenerateModelCode(t *testing.T) {
	resetGenerateState()
	skipIfDBUnavailable(t)

	restore := chdirToExample(t)
	defer restore()

	output := captureStdout(t, func() {
		if _, err := ExecuteCommand(Cmd, "model", "--app", "demoapp"); err != nil {
			t.Errorf("Failed to execute model command: %v", err)
		}
	})
	assertGenerateSuccess(t, output)
	// 生成成功后应产出 model 层文件
	if _, statErr := os.Stat(filepath.Join("apps", "demoapp", "model", "user_login_log.go")); statErr != nil {
		t.Errorf("generated model file not found: %v", statErr)
	}
}

// TestGenerateModuleCode 测试生成完整模块代码
// 依赖数据库（example/apps/demoapp/config/code_gen.yaml 中的 DSN），数据库不可达时自动跳过。
func TestGenerateModuleCode(t *testing.T) {
	resetGenerateState()
	skipIfDBUnavailable(t)

	restore := chdirToExample(t)
	defer restore()

	output := captureStdout(t, func() {
		if _, err := ExecuteCommand(Cmd, "module", "--app", "demoapp"); err != nil {
			t.Errorf("Failed to execute module command: %v", err)
		}
	})
	assertGenerateSuccess(t, output)
	// 生成成功后应产出 controller 层文件
	if _, statErr := os.Stat(filepath.Join("apps", "demoapp", "internal", "controller", "ctruser", "user.go")); statErr != nil {
		t.Errorf("generated controller file not found: %v", statErr)
	}
}

// TestApiTemplateRestfulPath 验证 api 模板生成的路径为 kebab-case 复数资源 + 动作子路径（与 module 模板的 restful 风格一致）
func TestApiTemplateRestfulPath(t *testing.T) {
	tplFuncs := template.FuncMap{
		TplFuncToKebabCase: toKebabCase,
		TplFuncPluralize:   pluralize,
	}
	params := map[string]interface{}{
		"AppName":                "demoapp",
		"PackageName":            "user",
		"BaseModulePath":         "github.com/example",
		"AppModuleName":          "demoapp",
		"StructName":             "UserLoginLog",
		"StructNameLowerCamel":   "userLoginLog",
		"FunctionName":           "Delete",
		"FunctionNameLowerCamel": "delete",
		"HttpMethod":             "POST",
		"Description":            "删除登录记录",
		"ApiDocTag":              "用户登录记录",
		"TargetFileExist":        false,
		"IsNewRouter":            true,
	}

	controllerTpl, err := template.New("controller.go.tpl").Funcs(tplFuncs).ParseFS(TemplatesFS, "generate/api/controller.go.tpl")
	if err != nil {
		t.Fatalf("parse api controller template: %v", err)
	}
	var buf bytes.Buffer
	if err := controllerTpl.ExecuteTemplate(&buf, "controller.go.tpl", params); err != nil {
		t.Fatalf("render api controller template: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "@Router /v1/demoapp/user-login-logs/delete [post]") {
		t.Errorf("controller template missing restful POST route:\n%s", out)
	}

	// GET 分支：路由 restful 化，且函数定义与方法名之间应有空格
	params["HttpMethod"] = "GET"
	buf.Reset()
	if err := controllerTpl.ExecuteTemplate(&buf, "controller.go.tpl", params); err != nil {
		t.Fatalf("render api controller template (GET): %v", err)
	}
	out = buf.String()
	if !strings.Contains(out, "@Router /v1/demoapp/user-login-logs/delete [get]") {
		t.Errorf("controller template missing restful GET route:\n%s", out)
	}
	if !strings.Contains(out, ") Delete(ctx *gin.Context)") {
		t.Errorf("controller template GET method missing space before function name:\n%s", out)
	}

	routerTpl, err := template.New("router.go.tpl").Funcs(tplFuncs).ParseFS(TemplatesFS, "generate/api/router.go.tpl")
	if err != nil {
		t.Fatalf("parse api router template: %v", err)
	}
	buf.Reset()
	if err := routerTpl.ExecuteTemplate(&buf, "router.go.tpl", params); err != nil {
		t.Fatalf("render api router template: %v", err)
	}
	if out := buf.String(); !strings.Contains(out, `"/user-login-logs/delete"`) {
		t.Errorf("router template missing restful route:\n%s", out)
	}
}

// TestGenerateApiCode 测试生成 API 代码
// 依赖数据库（example/apps/demoapp/config/code_gen.yaml 中的 DSN），数据库不可达时自动跳过。
func TestGenerateApiCode(t *testing.T) {
	resetGenerateState()
	skipIfDBUnavailable(t)

	restore := chdirToExample(t)
	defer restore()

	output := captureStdout(t, func() {
		if _, err := ExecuteCommand(Cmd, "api", "--app", "demoapp"); err != nil {
			t.Errorf("Failed to execute api command: %v", err)
		}
	})
	assertGenerateSuccess(t, output)
	// 生成成功后应在包名对应的 router 文件中追加新路由函数，并在 router.go 中注册
	userRouterContent, err := os.ReadFile(filepath.Join("apps", "demoapp", "internal", "router", "user.go"))
	if err != nil {
		t.Fatalf("read generated router file: %v", err)
	}
	if !strings.Contains(string(userRouterContent), "userLoginLogRouter") {
		t.Errorf("router/user.go missing userLoginLogRouter function:\n%s", userRouterContent)
	}
	if !strings.Contains(string(userRouterContent), `"/user-login-logs/delete1"`) {
		t.Errorf("router/user.go missing generated route:\n%s", userRouterContent)
	}
	routerGoContent, err := os.ReadFile(filepath.Join("apps", "demoapp", "internal", "router", "router.go"))
	if err != nil {
		t.Fatalf("read router.go: %v", err)
	}
	if !strings.Contains(string(routerGoContent), "userLoginLogRouter(groups)") {
		t.Errorf("router.go missing userLoginLogRouter registration:\n%s", routerGoContent)
	}
}

// 测试数据库 DSN 解析
func TestParseDatabaseDSN(t *testing.T) {
	cases := []struct {
		name    string
		dsn     string
		wantErr bool
		want    *DatabaseConfig
	}{
		{"empty", "", true, nil},
		{"missing scheme", "root:pwd@tcp(127.0.0.1:3306)/demo", true, nil},
		{"unsupported type", "mongodb://127.0.0.1:27017/demo", true, nil},
		{"mysql", "mysql://root:pwd@tcp(127.0.0.1:3306)/demo?charset=utf8mb4", false,
			&DatabaseConfig{Type: DBTypeMySQL, ConnStr: "root:pwd@tcp(127.0.0.1:3306)/demo?charset=utf8mb4"}},
		{"postgres", "postgresql://host=127.0.0.1 user=root dbname=demo", false,
			&DatabaseConfig{Type: DBTypePostgres, ConnStr: "postgresql://host=127.0.0.1 user=root dbname=demo"}},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got, err := ParseDatabaseDSN(tc.dsn)
			if tc.wantErr {
				if err == nil {
					t.Fatalf("ParseDatabaseDSN(%q) = %+v, want error", tc.dsn, got)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParseDatabaseDSN(%q) error: %v", tc.dsn, err)
			}
			if got.Type != tc.want.Type || got.ConnStr != tc.want.ConnStr {
				t.Fatalf("ParseDatabaseDSN(%q) = %+v, want %+v", tc.dsn, got, tc.want)
			}
		})
	}
}

// 测试内置字段判断
func TestIsBuiltInField(t *testing.T) {
	for _, name := range []string{"ID", "CreatedAt", "UpdatedAt", "DeletedAt"} {
		if !IsBuiltInField(name) {
			t.Errorf("IsBuiltInField(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"Name", "CompanyID", "CreatedBy"} {
		if IsBuiltInField(name) {
			t.Errorf("IsBuiltInField(%q) = true, want false", name)
		}
	}
}

// 测试系统字段判断
func TestIsSysField(t *testing.T) {
	for _, name := range []string{"ID", "CreatedAt", "CreatedBy", "UpdatedAt", "UpdatedBy", "DeletedAt", "DeletedBy"} {
		if !IsSysField(name) {
			t.Errorf("IsSysField(%q) = false, want true", name)
		}
	}
	for _, name := range []string{"Name", "CompanyID"} {
		if IsSysField(name) {
			t.Errorf("IsSysField(%q) = true, want false", name)
		}
	}
}

// 测试默认层级判断
func TestIsDefaultLayer(t *testing.T) {
	if !IsDefaultModelLayer("model") {
		t.Error("IsDefaultModelLayer(model) = false, want true")
	}
	if IsDefaultModelLayer("dao") {
		t.Error("IsDefaultModelLayer(dao) = true, want false")
	}
	if !IsDefaultDaoLayer("dao") {
		t.Error("IsDefaultDaoLayer(dao) = false, want true")
	}
	if IsDefaultDaoLayer("model") {
		t.Error("IsDefaultDaoLayer(model) = true, want false")
	}
}

// 测试基础类型判断
func TestIsBasicType(t *testing.T) {
	for _, typ := range []string{"string", "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "time.Time"} {
		if !IsBasicType(typ) {
			t.Errorf("IsBasicType(%q) = false, want true", typ)
		}
	}
	for _, typ := range []string{"json.RawMessage", "[]byte", "map[string]string", "CustomStruct"} {
		if IsBasicType(typ) {
			t.Errorf("IsBasicType(%q) = true, want false", typ)
		}
	}
}

// 测试时间字段判断
func TestHasTimeField(t *testing.T) {
	builtInTimeField := []ModelField{{FieldName: "CreatedAt", FieldType: "time.Time"}}
	if HasTimeField(builtInTimeField) {
		t.Error("HasTimeField with only built-in time field = true, want false")
	}

	normalTimeField := []ModelField{{FieldName: "Birthday", FieldType: "time.Time"}}
	if !HasTimeField(normalTimeField) {
		t.Error("HasTimeField with normal time field = false, want true")
	}

	noTimeField := []ModelField{{FieldName: "Name", FieldType: "string"}}
	if HasTimeField(noTimeField) {
		t.Error("HasTimeField without time field = true, want false")
	}
}

// 测试字段 import 推导
func TestGetFieldImports(t *testing.T) {
	fields := []ModelField{
		{FieldName: "Birthday", FieldType: "time.Time"},
		{FieldName: "Extra", FieldType: "json.RawMessage"},
		{FieldName: "Name", FieldType: "string"},
	}
	imports := GetFieldImports(fields)
	for _, want := range []string{"time", "encoding/json"} {
		if _, ok := imports[want]; !ok {
			t.Errorf("GetFieldImports missing import %q, got %v", want, imports)
		}
	}
	if len(imports) != 2 {
		t.Errorf("GetFieldImports = %v, want exactly 2 imports", imports)
	}

	plainFields := []ModelField{{FieldName: "Name", FieldType: "string"}}
	if len(GetFieldImports(plainFields)) != 0 {
		t.Errorf("GetFieldImports(%v) should be empty", plainFields)
	}
}

// 测试 calcFieldImports（含内置字段剔除与 exclude）
func TestCalcFieldImports(t *testing.T) {
	fields := []ModelField{
		{FieldName: "CreatedAt", FieldType: "time.Time"}, // 内置字段，应剔除
		{FieldName: "Birthday", FieldType: "time.Time"},
		{FieldName: "Extra", FieldType: "json.RawMessage"},
	}
	imports := calcFieldImports(fields)
	if len(imports) != 2 {
		t.Fatalf("calcFieldImports = %v, want [encoding/json time]", imports)
	}
	if imports[0] != "encoding/json" || imports[1] != "time" {
		t.Fatalf("calcFieldImports = %v, want sorted [encoding/json time]", imports)
	}

	importsNoTime := calcFieldImports(fields, "time")
	if len(importsNoTime) != 1 || importsNoTime[0] != "encoding/json" {
		t.Fatalf("calcFieldImports with exclude time = %v, want [encoding/json]", importsNoTime)
	}
}

// 测试从结构体名去除表名前缀
func TestRemoveTablePrefixFromStructName(t *testing.T) {
	cases := []struct {
		name, structName, tableName, prefix, want string
	}{
		{"empty prefix", "IamUsers", "iam_users", "", "IamUsers"},
		{"strip prefix", "IamUsers", "iam_users", "iam_", "Users"},
		{"struct already without prefix", "Users", "iam_users", "iam_", "Users"},
		{"table without prefix", "IamUsers", "sys_users", "iam_", "IamUsers"},
		{"multi segment prefix", "SysLogUser", "sys_log_user", "sys_log_", "User"},
		{"prefix only underscore", "Users", "iam_users", "_", "Users"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveTablePrefixFromStructName(tc.structName, tc.tableName, tc.prefix)
			if got != tc.want {
				t.Fatalf("RemoveTablePrefixFromStructName(%q, %q, %q) = %q, want %q",
					tc.structName, tc.tableName, tc.prefix, got, tc.want)
			}
		})
	}
}

// 测试从文件名去除表名前缀
func TestRemoveTablePrefixFromFilename(t *testing.T) {
	cases := []struct {
		name, filename, tableName, prefix, want string
	}{
		{"empty prefix", "iam_user.go", "iam_users", "", "iam_user.go"},
		{"strip prefix", "iam_user.go", "iam_users", "iam_", "user.go"},
		{"filename already without prefix", "user.go", "iam_users", "iam_", "user.go"},
		{"table without prefix", "iam_user.go", "sys_users", "iam_", "iam_user.go"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			got := RemoveTablePrefixFromFilename(tc.filename, tc.tableName, tc.prefix)
			if got != tc.want {
				t.Fatalf("RemoveTablePrefixFromFilename(%q, %q, %q) = %q, want %q",
					tc.filename, tc.tableName, tc.prefix, got, tc.want)
			}
		})
	}
}

// 测试蛇形转小驼峰（_id 特殊处理）
func TestSnakeToLowerCamelWithID(t *testing.T) {
	cases := []struct{ in, want string }{
		{"tenant_id", "tenantID"},
		{"user_name", "userName"},
		{"id", "id"},
		{"_id", "id"},
		{"user", "user"},
		{"", ""},
	}
	for _, tc := range cases {
		if got := SnakeToLowerCamelWithID(tc.in); got != tc.want {
			t.Errorf("SnakeToLowerCamelWithID(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

// 测试 GetAppInfo 解析应用模块信息
func TestGetAppInfo(t *testing.T) {
	rootDir := t.TempDir()
	appDir := filepath.Join(rootDir, "apps", "demoapp")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "go.mod"),
		[]byte("module github.com/example/demoapp\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	info, err := GetAppInfo(appDir)
	if err != nil {
		t.Fatalf("GetAppInfo error: %v", err)
	}
	if info.AppName != "demoapp" {
		t.Errorf("AppName = %q, want %q", info.AppName, "demoapp")
	}
	if info.ProjectName != filepath.Base(rootDir) {
		t.Errorf("ProjectName = %q, want %q", info.ProjectName, filepath.Base(rootDir))
	}
	if info.ProjectRootPath != rootDir {
		t.Errorf("ProjectRootPath = %q, want %q", info.ProjectRootPath, rootDir)
	}
	if info.BaseModulePath != "github.com/example" {
		t.Errorf("BaseModulePath = %q, want %q", info.BaseModulePath, "github.com/example")
	}
	if info.AppModuleName != "demoapp" {
		t.Errorf("AppModuleName = %q, want %q", info.AppModuleName, "demoapp")
	}
}

// 测试 GetAppInfo 在存在 .git 目录时使用 git 根目录
func TestGetAppInfoWithGitRoot(t *testing.T) {
	rootDir := t.TempDir()
	if err := os.MkdirAll(filepath.Join(rootDir, ".git"), 0o755); err != nil {
		t.Fatalf("mkdir .git: %v", err)
	}
	appDir := filepath.Join(rootDir, "apps", "demoapp")
	if err := os.MkdirAll(appDir, 0o755); err != nil {
		t.Fatalf("mkdir app dir: %v", err)
	}
	if err := os.WriteFile(filepath.Join(appDir, "go.mod"),
		[]byte("module github.com/example/demoapp\n\ngo 1.26.1\n"), 0o644); err != nil {
		t.Fatalf("write go.mod: %v", err)
	}

	info, err := GetAppInfo(appDir)
	if err != nil {
		t.Fatalf("GetAppInfo error: %v", err)
	}
	if info.ProjectRootPath != rootDir {
		t.Errorf("ProjectRootPath = %q, want git root %q", info.ProjectRootPath, rootDir)
	}
}

// 测试读取模块信息
func TestGetModuleInfo(t *testing.T) {
	t.Run("missing go.mod errors", func(t *testing.T) {
		dir := t.TempDir()
		if _, _, err := getModuleInfo(dir); err == nil {
			t.Fatal("getModuleInfo on missing go.mod should return error")
		}
	})

	t.Run("valid module path", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"),
			[]byte("module github.com/example/demoapp\n\ngo 1.26.1\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		base, name, err := getModuleInfo(dir)
		if err != nil {
			t.Fatalf("getModuleInfo error: %v", err)
		}
		if base != "github.com/example" {
			t.Errorf("base = %q, want %q", base, "github.com/example")
		}
		if name != "demoapp" {
			t.Errorf("name = %q, want %q", name, "demoapp")
		}
	})

	t.Run("single segment module path errors", func(t *testing.T) {
		dir := t.TempDir()
		if err := os.WriteFile(filepath.Join(dir, "go.mod"), []byte("module demo\n"), 0o644); err != nil {
			t.Fatalf("write go.mod: %v", err)
		}
		if _, _, err := getModuleInfo(dir); err == nil {
			t.Fatal("single segment module path should return error")
		}
	})
}

// 测试嵌入模板复制到临时目录
func TestCopyEmbeddedTemplatesToTempDir(t *testing.T) {
	tempDir, err := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "generate/model")
	if err != nil {
		t.Fatalf("CopyEmbeddedTemplatesToTempDir error: %v", err)
	}
	defer os.RemoveAll(tempDir)

	for _, name := range []string{"model.go.tpl", "dao.go.tpl", "table.go.tpl", "object.go.tpl"} {
		if _, statErr := os.Stat(filepath.Join(tempDir, name)); statErr != nil {
			t.Errorf("missing copied template %s: %v", name, statErr)
		}
	}
}

// 测试命令执行辅助函数
func TestExecuteCommand(t *testing.T) {
	output, err := ExecuteCommand(Cmd)
	if err != nil {
		t.Fatalf("ExecuteCommand error: %v", err)
	}
	for _, sub := range []string{"module", "model", "api"} {
		if !strings.Contains(output, sub) {
			t.Errorf("help output missing subcommand %q", sub)
		}
	}

	if _, err := ExecuteCommand(Cmd, "--not-exist"); err == nil {
		t.Error("ExecuteCommand with unknown flag should return error")
	}
}
