package generate

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/morehao/gocli/pkg/codegen"
	"github.com/morehao/golib/gutil"
	"github.com/spf13/cobra"
	"gorm.io/driver/mysql"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// Cmd represents the generate command
var Cmd = &cobra.Command{
	Use:   "generate",
	Short: "Generate code based on templates",
	Long:  `Generate code for different layers like module, model, and endpoint based on predefined templates.`,
}

// backendCmd is the parent command for backend generation.
var backendCmd = &cobra.Command{
	Use:   "backend",
	Short: "Generate backend code",
}

// frontendCmd is the parent command for frontend generation.
var frontendCmd = &cobra.Command{
	Use:   "frontend",
	Short: "Generate frontend code",
}

var backendModuleCmd = &cobra.Command{
	Use:   "module",
	Short: "Generate a backend module (model/dao/object/service/controller/dto/router/code)",
	RunE:  runBackend(codegen.ModeBackendModule),
}

var backendModelCmd = &cobra.Command{
	Use:   "model",
	Short: "Generate backend model/dao/object layers",
	RunE:  runBackend(codegen.ModeBackendModel),
}

var backendEndpointCmd = &cobra.Command{
	Use:   "endpoint",
	Short: "Generate a backend API endpoint",
	RunE:  runBackend(codegen.ModeBackendEndpoint),
}

var frontendPageCmd = &cobra.Command{
	Use:   "page",
	Short: "Generate a frontend page",
	RunE:  runFrontend(codegen.ModeFrontendPage),
}

var fullstackCmd = &cobra.Command{
	Use:   "fullstack",
	Short: "Generate backend and frontend code together",
	RunE:  runFrontend(codegen.ModeFullstack),
}

func init() {
	backendCmd.AddCommand(backendModuleCmd, backendModelCmd, backendEndpointCmd)
	frontendCmd.AddCommand(frontendPageCmd)
	Cmd.AddCommand(backendCmd, frontendCmd, fullstackCmd)

	for _, subCmd := range []*cobra.Command{backendModuleCmd, backendModelCmd, backendEndpointCmd} {
		addCommonFlags(subCmd)
	}
	addCommonFlags(frontendPageCmd)
	addCommonFlags(fullstackCmd)
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("app", "a", "", "App name to generate code for (e.g., demoapp)")
	cmd.Flags().String("table", "", "Table name to generate code for (overrides code_gen.yaml)")
	cmd.Flags().String("package", "", "Package name to generate code for (overrides code_gen.yaml)")
	cmd.Flags().Bool("dry-run", false, "Print what would be generated without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")
}

// buildOptions 从命令行 flag 构造 codegen.Options。
func buildOptions(cmd *cobra.Command, mode codegen.Mode) (codegen.Options, error) {
	projectRoot, err := os.Getwd()
	if err != nil {
		return codegen.Options{}, fmt.Errorf("getwd: %w", err)
	}
	appName, _ := cmd.Flags().GetString("app")
	if appName == "" {
		return codegen.Options{}, fmt.Errorf("please provide an app name using --app flag")
	}
	table, _ := cmd.Flags().GetString("table")
	packageName, _ := cmd.Flags().GetString("package")
	dryRun, _ := cmd.Flags().GetBool("dry-run")
	force, _ := cmd.Flags().GetBool("force")

	return codegen.Options{
		AppName:     appName,
		ProjectRoot: projectRoot,
		Table:       table,
		PackageName: packageName,
		Mode:        mode,
		Force:       force,
		DryRun:      dryRun,
	}, nil
}

// openDB 从目标应用 config/code_gen.yaml 读取 DSN 并建立数据库连接。
func openDB(projectRoot, appName string) (*gorm.DB, error) {
	configFilepath := filepath.Join(projectRoot, "apps", appName, "config", "code_gen.yaml")
	if _, err := os.Stat(configFilepath); err != nil {
		return nil, fmt.Errorf("config file does not exist: %s: %w", configFilepath, err)
	}
	var cfg codegen.Config
	gutil.LoadYamlConfig(configFilepath, &cfg)

	dbCfg, err := codegen.ParseDatabaseDSN(cfg.DatabaseDSN)
	if err != nil {
		return nil, fmt.Errorf("parse database dsn: %w", err)
	}
	switch dbCfg.Type {
	case codegen.DBTypeMySQL:
		return gorm.Open(mysql.Open(dbCfg.ConnStr), &gorm.Config{})
	case codegen.DBTypePostgres:
		return gorm.Open(postgres.Open(dbCfg.ConnStr), &gorm.Config{})
	default:
		return nil, fmt.Errorf("unsupported database type: %s", dbCfg.Type)
	}
}

// runBackend 返回用于后端叶子命令的 RunE，按 mode 构造 Generator（需要 DB）并调用 Generate。
func runBackend(mode codegen.Mode) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		opts, err := buildOptions(cmd, mode)
		if err != nil {
			return err
		}
		db, err := openDB(opts.ProjectRoot, opts.AppName)
		if err != nil {
			return fmt.Errorf("open database connection: %w", err)
		}
		gen, err := codegen.New(db, opts)
		if err != nil {
			return err
		}
		return execute(gen)
	}
}

// runFrontend 返回用于前端/全栈叶子命令的 RunE（本期返回 ErrNotImplemented）。
func runFrontend(mode codegen.Mode) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		opts, err := buildOptions(cmd, mode)
		if err != nil {
			return err
		}
		gen, err := codegen.New(nil, opts)
		if err != nil {
			return err
		}
		return execute(gen)
	}
}

// execute 统一调用 Generator.Generate 并打印结果或错误。
func execute(gen *codegen.Generator) error {
	res, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generate failed: %w", err)
	}
	fmt.Println("Generated successfully")
	for _, f := range res.WrittenFiles {
		fmt.Printf("  wrote %s\n", f)
	}
	return nil
}
