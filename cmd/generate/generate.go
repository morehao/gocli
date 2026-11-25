package generate

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

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

var undoCmd = &cobra.Command{
	Use:   "undo",
	Short: "Restore aggregate files (types/api/App.tsx/go) from the pre-generation snapshot",
	RunE:  runUndo,
}

func init() {
	backendCmd.AddCommand(backendModuleCmd, backendModelCmd, backendEndpointCmd)
	frontendCmd.AddCommand(frontendPageCmd)
	Cmd.AddCommand(backendCmd, frontendCmd, fullstackCmd, undoCmd)

	for _, subCmd := range []*cobra.Command{backendModuleCmd, backendModelCmd, backendEndpointCmd} {
		addCommonFlags(subCmd)
	}
	addCommonFlags(frontendPageCmd)
	addCommonFlags(fullstackCmd)
	undoCmd.Flags().StringP("app", "a", "", "App name (for locating the project root)")
}

func addCommonFlags(cmd *cobra.Command) {
	cmd.Flags().StringP("app", "a", "", "App name to generate code for (e.g., demoapp)")
	cmd.Flags().String("table", "", "Table name to generate code for (overrides code_gen.yaml)")
	cmd.Flags().String("package", "", "Package name to generate code for (overrides code_gen.yaml)")
	cmd.Flags().Bool("dry-run", false, "Print what would be generated without writing files")
	cmd.Flags().Bool("force", false, "Overwrite existing files")
	cmd.Flags().Bool("all", false, "Generate for all tables in the database (ignores --table)")
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
		return runMulti(cmd, mode)
	}
}

// runFrontend 返回用于前端/全栈叶子命令的 RunE（需要 DB 分析表结构）。
func runFrontend(mode codegen.Mode) func(cmd *cobra.Command, args []string) error {
	return func(cmd *cobra.Command, args []string) error {
		return runMulti(cmd, mode)
	}
}

// runMulti 支持 --table 逗号分隔多表 与 --all（枚举库内全部表），逐表生成。
func runMulti(cmd *cobra.Command, mode codegen.Mode) error {
	baseOpts, err := buildOptions(cmd, mode)
	if err != nil {
		return err
	}
	db, err := openDB(baseOpts.ProjectRoot, baseOpts.AppName)
	if err != nil {
		return fmt.Errorf("open database connection: %w", err)
	}

	all, _ := cmd.Flags().GetBool("all")
	tables := splitCSV(baseOpts.Table)
	if all {
		tables, err = listTables(db)
		if err != nil {
			return fmt.Errorf("list tables: %w", err)
		}
	}
	// 未显式给 --table 时，回退到 code_gen.yaml 中配置的表（保持 `make codegen APP=x COMMAND=module`
	// 这类「只靠 yaml」的旧用法可用）。
	if len(tables) == 0 {
		if t := yamlTable(baseOpts.ProjectRoot, baseOpts.AppName, mode); t != "" {
			tables = []string{t}
		}
	}
	if len(tables) == 0 {
		return fmt.Errorf("请用 --table 指定表、在 code_gen.yaml 配置 table_name，或用 --all 生成全部表")
	}

	var firstErr error
	for _, tbl := range tables {
		opts := baseOpts
		opts.Table = tbl
		if opts.PackageName == "" {
			opts.PackageName = tbl
		}
		gen, err := codegen.New(db, opts)
		if err != nil {
			fmt.Printf("[%s] 跳过：%v\n", tbl, err)
			if firstErr == nil {
				firstErr = err
			}
			continue
		}
		if err := execute(gen); err != nil {
			fmt.Printf("[%s] 失败：%v\n", tbl, err)
			if firstErr == nil {
				firstErr = err
			}
		}
	}
	return firstErr
}

func splitCSV(s string) []string {
	if s == "" {
		return nil
	}
	parts := strings.Split(s, ",")
	out := make([]string, 0, len(parts))
	for _, p := range parts {
		if v := strings.TrimSpace(p); v != "" {
			out = append(out, v)
		}
	}
	return out
}

// yamlTable 读取 code_gen.yaml 里与 mode 对应的表名，作为未指定 --table 时的回退。
func yamlTable(projectRoot, appName string, mode codegen.Mode) string {
	cfgPath := filepath.Join(projectRoot, "apps", appName, "config", "code_gen.yaml")
	if _, err := os.Stat(cfgPath); err != nil {
		return ""
	}
	var cfg codegen.Config
	gutil.LoadYamlConfig(cfgPath, &cfg)
	switch mode {
	case codegen.ModeBackendModel:
		if cfg.Model.TableName != "" {
			return cfg.Model.TableName
		}
	case codegen.ModeBackendEndpoint:
		// endpoint 用 module 表（api 段按 target_filename 命名，无独立表名）
		if cfg.Module.TableName != "" {
			return cfg.Module.TableName
		}
	}
	if cfg.Module.TableName != "" {
		return cfg.Module.TableName
	}
	return cfg.Model.TableName
}

// listTables 枚举当前库全部用户表（postgres: public；mysql: 当前 database）。
func listTables(db *gorm.DB) ([]string, error) {
	var names []string
	dialect := db.Dialector.Name()
	var sqlStr string
	if dialect == "postgres" {
		sqlStr = "SELECT table_name FROM information_schema.tables WHERE table_schema='public' AND table_type='BASE TABLE' ORDER BY table_name"
	} else {
		sqlStr = "SELECT table_name FROM information_schema.tables WHERE table_schema=DATABASE() AND table_type='BASE TABLE' ORDER BY table_name"
	}
	if err := db.Raw(sqlStr).Scan(&names).Error; err != nil {
		return nil, err
	}
	return names, nil
}

// runUndo 从快照还原聚合文件（types/api/App.tsx + router/table/code）。
func runUndo(cmd *cobra.Command, args []string) error {
	projectRoot, err := os.Getwd()
	if err != nil {
		return err
	}
	appName, _ := cmd.Flags().GetString("app")
	if appName == "" {
		return fmt.Errorf("please provide an app name using --app flag")
	}
	gen, err := codegen.New(nil, codegen.Options{AppName: appName, ProjectRoot: projectRoot})
	if err != nil {
		return err
	}
	restored, err := gen.Undo()
	if err != nil {
		return fmt.Errorf("undo failed: %w", err)
	}
	if len(restored) == 0 {
		fmt.Println("没有可还原的快照（未找到 .codegen/backup）")
		return nil
	}
	fmt.Println("已还原：")
	for _, f := range restored {
		fmt.Printf("  %s\n", f)
	}
	return nil
}

// execute 统一调用 Generator.Generate 并打印结果或错误。
func execute(gen *codegen.Generator) error {
	res, err := gen.Generate()
	if err != nil {
		return fmt.Errorf("generate failed: %w", err)
	}
	if len(res.WrittenFiles) == 0 {
		// 增量跳过等场景：无写入
		for _, w := range res.Warnings {
			fmt.Printf("  %s\n", w)
		}
		return nil
	}
	fmt.Println("Generated successfully")
	for _, f := range res.WrittenFiles {
		fmt.Printf("  wrote %s\n", f)
	}
	for _, w := range res.Warnings {
		fmt.Printf("  %s\n", w)
	}
	return nil
}
