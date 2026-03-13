package generate

import "github.com/morehao/golib/codegen"

var (
	defaultLayerParentDirMap = map[codegen.LayerName]string{
		codegen.LayerNameController: "internal",
		codegen.LayerNameService:    "internal",
		codegen.LayerNameDto:        "internal",
	}

	defaultLayerNameMap = map[codegen.LayerName]codegen.LayerName{}

	defaultLayerPrefixMap = map[codegen.LayerName]codegen.LayerPrefix{}
)

func buildLayerNameMap(serviceName string) map[codegen.LayerName]codegen.LayerName {
	return map[codegen.LayerName]codegen.LayerName{
		codegen.LayerNameModel: codegen.LayerName(serviceName + "model"),
		codegen.LayerNameDao:   codegen.LayerName(serviceName + "dao"),
	}
}

type Config struct {
	MysqlDSN    string       `yaml:"mysql_dsn"`    // MySQL 连接字符串
	ServiceName string       `yaml:"service_name"` // 服务名
	Module      ModuleConfig `yaml:"module"`       // 模块生成配置
	Model       ModelConfig  `yaml:"model"`        // 模型生成配置
	Api         ApiConfig    `yaml:"api"`          // 控制器生成配置
	appInfo     AppInfo
}

type AppInfo struct {
	AppPathInProject string
	ProjectName      string
	AppName          string
	ProjectRootPath  string
	ModulePath       string
}

type ModuleConfig struct {
	PackageName string `yaml:"package_name"` // 包名
	Description string `yaml:"description"`  // 描述
	TableName   string `yaml:"table_name"`   // 表名
	TablePrefix string `yaml:"table_prefix"` // 表名前缀，生成结构体名时会去除此前缀，如 iam_
}

type ModelConfig struct {
	PackageName string `yaml:"package_name"` // 包名
	Description string `yaml:"description"`  // 描述
	TableName   string `yaml:"table_name"`   // 表名
	TablePrefix string `yaml:"table_prefix"` // 表名前缀，生成结构体名时会去除此前缀，如 iam_
}

type ApiConfig struct {
	PackageName    string `yaml:"package_name"`    // 包名，如user
	TargetFilename string `yaml:"target_filename"` // 目标文件名，生成的代码写入的目标文件名
	FunctionName   string `yaml:"function_name"`   // 函数名
	HttpMethod     string `yaml:"http_method"`     // http方法
	ApiDocTag      string `yaml:"api_doc_tag"`     // api文档tag
	Description    string `yaml:"description"`     // 描述
}
