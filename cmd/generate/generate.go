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

	"github.com/morehao/golib/gutil"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

//go:embed template
var TemplatesFS embed.FS

var workDir string
var cfg *Config
var DBClient *gorm.DB

// Cmd represents the generate command
var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code based on templates",
	Long:  `Generate code for different layers like module, model, and API based on predefined templates.`,
}

var moduleCmd = &cobra.Command{
	Use:   "module",
	Short: "Generate module code",
	Run:   runGenerate(genModule),
}

var modelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate model code",
	Run:   runGenerate(genModel),
}

var apiCmd = &cobra.Command{
	Use:   "api",
	Short: "Generate api code",
	Run:   runGenerate(genApi),
}

func init() {
	Cmd.AddCommand(moduleCmd, modelCmd, apiCmd)
}

func runGenerate(genFn func() error) func(cmd *cobra.Command, args []string) {
	return func(cmd *cobra.Command, args []string) {
		projectRootDir, _ := os.Getwd()

		appName, _ := cmd.Flags().GetString("app")
		if appName == "" {
			fmt.Println("Please provide an app name using --app flag")
			return
		}

		workDir = filepath.Join(projectRootDir, "apps", appName)

		if _, err := os.Stat(workDir); os.IsNotExist(err) {
			fmt.Printf("App directory does not exist: %s\n", workDir)
			return
		}

		if cfg == nil {
			configFilepath := filepath.Join(workDir, "config", "code_gen.yaml")
			if _, err := os.Stat(configFilepath); os.IsNotExist(err) {
				fmt.Printf("Config file does not exist: %s\n", configFilepath)
				return
			}

			gutil.LoadYamlConfig(configFilepath, &cfg)
			appInfo, getAppInfoErr := GetAppInfo(workDir)
			if getAppInfoErr != nil {
				fmt.Printf("Get app info error: %v\n", getAppInfoErr)
				return
			}
			cfg.appInfo = *appInfo
		}

		if DBClient == nil {
			dbCfg, parseErr := ParseDatabaseDSN(cfg.DatabaseDSN)
			if parseErr != nil {
				fmt.Printf("Parse database dsn error: %v\n", parseErr)
				return
			}

			var dbClient *gorm.DB
			var openErr error
			switch dbCfg.Type {
			case DBTypeMySQL:
				dbClient, openErr = gorm.Open(mysql.Open(dbCfg.ConnStr), &gorm.Config{})
			case DBTypePostgres:
				dbClient, openErr = gorm.Open(postgres.Open(dbCfg.ConnStr), &gorm.Config{})
			default:
				fmt.Printf("Unsupported database type: %s\n", dbCfg.Type)
				return
			}
			if openErr != nil {
				fmt.Printf("Open database connection error: %v\n", openErr)
				return
			}
			DBClient = dbClient
		}

		if err := genFn(); err != nil {
			fmt.Printf("Error generating: %v\n", err)
			return
		}
		fmt.Println("Generated successfully")
	}
}

func init() {
	for _, subCmd := range []*cobra.Command{moduleCmd, modelCmd, apiCmd} {
		subCmd.Flags().StringP("app", "a", "", "App name to generate code for (e.g., demoapp)")
	}
}
