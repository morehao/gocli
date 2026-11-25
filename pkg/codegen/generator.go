// Package codegen 提供 gocli 代码生成能力，供 CLI 与外部 Go 程序 import 调用。
//
// 核心入口是 Generator：由 New 构造，持有数据库连接、生成选项与解析后的配置，
// 通过 Generate 或各 GenerateXxx 方法完成具体生成任务。
package codegen

import (
	"fmt"
	"os"
	"path/filepath"

	tpl "github.com/morehao/gocli/template"
	"github.com/morehao/golib/gutil"
	"gorm.io/gorm"
)

// TemplatesFS 代码生成模板（内嵌于顶层 template 包，见 template/generate）
var TemplatesFS = tpl.GenerateFS

// Generator 代码生成器。它把原先依赖包级全局（workDir/cfg/DBClient）的生成逻辑
// 收敛到实例字段上，使生成过程可复用、可并发构造多个实例。
type Generator struct {
	db      *gorm.DB // 数据库连接，由外部传入，New 不负责建立连接
	opts    Options  // 生成选项
	workDir string   // 目标应用目录 = filepath.Join(opts.ProjectRoot, "apps", opts.AppName)
	cfg     *Config  // 由 opts + code_gen.yaml 合并得到
	appInfo *AppInfo // 应用模块信息
}

// New 构造一个 Generator。
//
// workDir 由 opts.ProjectRoot 与 opts.AppName 组合得到；随后读取
// <workDir>/config/code_gen.yaml 到 cfg，并解析 appInfo。若 opts.PackageName/opts.Table
// 非空则覆盖 cfg 对应字段。db 由外部传入，New 不建立数据库连接。
func New(db *gorm.DB, opts Options) (*Generator, error) {
	if opts.AppName == "" {
		return nil, fmt.Errorf("codegen: AppName is required")
	}
	if opts.ProjectRoot == "" {
		return nil, fmt.Errorf("codegen: ProjectRoot is required")
	}

	workDir := filepath.Join(opts.ProjectRoot, "apps", opts.AppName)
	if _, err := os.Stat(workDir); err != nil {
		return nil, fmt.Errorf("codegen: app directory does not exist: %s: %w", workDir, err)
	}

	configFilepath := filepath.Join(workDir, "config", "code_gen.yaml")
	cfg := &Config{}
	if _, err := os.Stat(configFilepath); err != nil {
		return nil, fmt.Errorf("codegen: config file does not exist: %s: %w", configFilepath, err)
	}
	gutil.LoadYamlConfig(configFilepath, cfg)

	appInfo, err := GetAppInfo(workDir)
	if err != nil {
		return nil, fmt.Errorf("codegen: get app info: %w", err)
	}
	cfg.appInfo = *appInfo

	// Options 覆盖：PackageName / Table
	if opts.PackageName != "" {
		cfg.Module.PackageName = opts.PackageName
		cfg.Model.PackageName = opts.PackageName
		cfg.Api.PackageName = opts.PackageName
	}
	if opts.Table != "" {
		cfg.Module.TableName = opts.Table
		cfg.Model.TableName = opts.Table
	}

	return &Generator{
		db:      db,
		opts:    opts,
		workDir: workDir,
		cfg:     cfg,
		appInfo: appInfo,
	}, nil
}
