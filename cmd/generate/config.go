package generate

import "github.com/morehao/golib/codegen"

type Config struct {
	MysqlDSN          string                                    `yaml:"mysql_dsn"`            // MySQL 连接字符串
	LayerParentDirMap map[codegen.LayerName]string              `yaml:"layer_parent_dir_map"` // 模块父目录映射
	LayerNameMap      map[codegen.LayerName]codegen.LayerName   `yaml:"layer_name_map"`       // 模块层名称映射
	LayerPrefixMap    map[codegen.LayerName]codegen.LayerPrefix `yaml:"layer_prefix_map"`     // 模块层前缀映射
	Module            ModuleConfig                              `yaml:"module"`               // 模块生成配置
	Model             ModelConfig                               `yaml:"model"`                // 模型生成配置
	Api               ApiConfig                                 `yaml:"api"`                  // 控制器生成配置
	appInfo           AppInfo
}

// AppInfo 应用信息，示例路径：go-gin-web/internal/apps/demoapp
type AppInfo struct {
	AppPathInProject string // go-gin-web/internal/apps/demoapp
	ProjectName      string // go-gin-web
	AppName          string // demoapp
	ProjectRootPath  string // /Users/morehao/xxx/go-gin-web (项目根目录绝对路径)
}

type CodeGen struct {
	ServiceName string       `yaml:"service_name"` // 服务名
	AppName     string       `yaml:"app_name"`     // 应用名
	Mode        string       `yaml:"mode"`         // 生成模式，支持：module、model、api
	Module      ModuleConfig `yaml:"module"`       // 模块生成配置
	Model       ModelConfig  `yaml:"model"`        // 模型生成配置
	Api         ApiConfig    `yaml:"api"`          // 控制器生成配置
}

type ModuleConfig struct {
	PackageName string `yaml:"package_name"` // 包名
	Description string `yaml:"description"`  // 描述
	TableName   string `yaml:"table_name"`   // 表名
}

type ModelConfig struct {
	PackageName string `yaml:"package_name"` // 包名
	Description string `yaml:"description"`  // 描述
	TableName   string `yaml:"table_name"`   // 表名
}

type ApiConfig struct {
	PackageName    string `yaml:"package_name"`    // 包名，如user
	TargetFilename string `yaml:"target_filename"` // 目标文件名，生成的代码写入的目标文件名
	FunctionName   string `yaml:"function_name"`   // 函数名
	HttpMethod     string `yaml:"http_method"`     // http方法
	ApiDocTag      string `yaml:"api_doc_tag"`     // api文档tag
	Description    string `yaml:"description"`     // 描述
}
