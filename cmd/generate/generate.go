/*
 * @Author: morehao morehao@qq.com
 * @Date: 2024-11-30 11:42:59
 * @LastEditors: morehao morehao@qq.com
 * @LastEditTime: 2025-05-18 21:09:10
 * @FilePath: /gocli/cmd/generate/generate.go
 * @Description: 这是默认设置,请设置`customMade`, 打开koroFileHeader查看配置 进行设置: https://github.com/OBKoro1/koro1FileHeader/wiki/%E9%85%8D%E7%BD%AE
 */
package generate

import (
	"embed"
	"fmt"
	"os"
	"path/filepath"

	"github.com/morehao/golib/conf"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
)

//go:embed template
var TemplatesFS embed.FS

var workDir string
var cfg *Config
var MysqlClient *gorm.DB

// Cmd represents the generate command
var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code based on templates",
	Long:  `Generate code for different layers like module, model, and API based on predefined templates.`,
	Run: func(cmd *cobra.Command, args []string) {
		// 获取当前工作目录（项目根目录）
		projectRootDir, _ := os.Getwd()
		
		// 获取 app 名称
		appName, _ := cmd.Flags().GetString("app")
		if appName == "" {
			fmt.Println("Please provide an app name using --app flag")
			return
		}

		// 构建 app 的完整路径
		workDir = filepath.Join(projectRootDir, "apps", appName)
		
		// 检查 app 目录是否存在
		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			fmt.Printf("App directory does not exist: %s\n", workDir)
			return
		}

		// 初始化配置和 MySQL 客户端
		if cfg == nil {
			// 配置文件路径：apps/{appName}/config/code_gen.yaml
			configFilepath := filepath.Join(workDir, "config", "code_gen.yaml")
			if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
				fmt.Printf("Config file does not exist: %s\n", configFilepath)
				return
			}
			
			conf.LoadConfig(configFilepath, &cfg)
			appInfo, getAppInfoErr := GetAppInfo(workDir)
			if getAppInfoErr != nil {
				fmt.Printf("Get app info error: %v\n", getAppInfoErr)
				return
			}
			cfg.appInfo = *appInfo
		}
		
		// 延迟初始化 Mysql 客户端
		if MysqlClient == nil {
			mysqlClient, getMysqlClientErr := gorm.Open(mysql.Open(cfg.MysqlDSN), &gorm.Config{})
			if getMysqlClientErr != nil {
				fmt.Printf("Get mysql client error: %v\n", getMysqlClientErr)
				return
			}
			MysqlClient = mysqlClient
		}

		mode, _ := cmd.Flags().GetString("mode")

		switch mode {
		case "module":
			if err := genModule(); err != nil {
				fmt.Printf("Error generating module: %v\n", err)
				return
			}
			fmt.Println("Module generated successfully")
		case "model":
			if err := genModel(); err != nil {
				fmt.Printf("Error generating model: %v\n", err)
				return
			}
			fmt.Println("Model generated successfully")
		case "api":
			if err := genApi(); err != nil {
				fmt.Printf("Error generating api: %v\n", err)
				return
			}
			fmt.Println("API generated successfully")

		// 这里可以添加其他模式的处理逻辑
		default:
			fmt.Println("Invalid mode. Available modes are: module, model, api")
		}
	},
}

func init() {
	// 定义 generate 命令的参数
	Cmd.Flags().StringP("mode", "m", "", "Mode of code generation (module, model, api)")
	Cmd.Flags().StringP("app", "a", "", "App name to generate code for (e.g., demoapp)")
}
