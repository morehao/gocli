package codegen

import (
	"errors"
	"fmt"
)

// Mode 生成模式，决定统一入口 Generate 的分发目标。
type Mode string

const (
	// ModeBackendModule 生成完整后端模块（model/dao/object/service/controller/dto/router/code）。
	ModeBackendModule Mode = "backend_module"
	// ModeBackendModel 生成后端 model/dao/object 层。
	ModeBackendModel Mode = "backend_model"
	// ModeBackendEndpoint 生成后端单个 API 接口。
	ModeBackendEndpoint Mode = "backend_endpoint"
	// ModeFrontendPage 生成前端页面（本期未实现）。
	ModeFrontendPage Mode = "frontend_page"
	// ModeFullstack 前后端全栈生成（本期未实现）。
	ModeFullstack Mode = "fullstack"
)

// ErrNotImplemented 表示该生成能力本期尚未实现（如前端 webgen）。
var ErrNotImplemented = errors.New("codegen: not implemented")

// KV 通用键值对，用于前端字段选项等。
type KV struct {
	Value any
	Label string
}

// WebConfig 前端生成配置（本期透传，由后续 webgen 任务消费）。
type WebConfig struct {
	Enabled      bool
	PageDir      string
	MenuGroup    string
	MenuLabel    string
	RegisterMenu bool
	FieldWidgets map[string]string
	FieldOptions map[string][]KV
}

// Options 生成选项。
type Options struct {
	AppName     string // 必填
	ProjectRoot string // 必填，目标工程根（含 apps/<AppName>）
	Table       string // 后端/前端必填（表名）
	PackageName string // 可选覆盖，缺省用 yaml
	Mode        Mode
	Force       bool
	DryRun      bool
	NoSnapshot  bool // 由批量调用方置位：批量前统一快照一次后，各表不再各自快照
	Web         WebConfig
}

// Result 生成结果。
type Result struct {
	// WrittenFiles 本次写入/追加的文件路径列表。
	WrittenFiles []string
	// Warnings 非致命提示（如页面已存在跳过、App.tsx 缺锚点等）。
	Warnings []string
}

// Generate 统一入口：按 g.opts.Mode 分发到对应的 GenerateXxx 方法。
func (g *Generator) Generate() (*Result, error) {
	switch g.opts.Mode {
	case ModeBackendModule:
		return g.GenerateBackendModule()
	case ModeBackendModel:
		return g.GenerateBackendModel()
	case ModeBackendEndpoint:
		return g.GenerateBackendEndpoint()
	case ModeFrontendPage:
		return g.GenerateFrontendPage()
	case ModeFullstack:
		return g.GenerateFullstack()
	default:
		return nil, errors.New("codegen: unknown mode: " + string(g.opts.Mode))
	}
}

// GenerateBackendModule 生成完整后端模块。
func (g *Generator) GenerateBackendModule() (*Result, error) {
	if g.opts.DryRun {
		return &Result{Warnings: []string{"[dry-run] 将执行后端完整模块生成，已跳过写盘"}}, nil
	}
	if _, res, err := g.buildTableMetaFor(g.cfg.Module.TableName, g.cfg.Module.PackageName); err == nil && res != nil {
		if skip, why := g.skipIfUnchanged(res, ModeBackendModule); skip {
			return &Result{Warnings: []string{why}}, nil
		}
		if err := g.genModule(); err != nil {
			return nil, err
		}
		g.recordManifest(res, ModeBackendModule)
		return &Result{}, nil
	}
	if err := g.genModule(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateBackendModel 生成后端 model/dao/object 层。
func (g *Generator) GenerateBackendModel() (*Result, error) {
	if g.opts.DryRun {
		return &Result{Warnings: []string{"[dry-run] 将执行后端数据层生成，已跳过写盘"}}, nil
	}
	if _, res, err := g.buildTableMetaFor(g.cfg.Model.TableName, g.cfg.Model.PackageName); err == nil && res != nil {
		if skip, why := g.skipIfUnchanged(res, ModeBackendModel); skip {
			return &Result{Warnings: []string{why}}, nil
		}
		if err := g.genModel(); err != nil {
			return nil, err
		}
		g.recordManifest(res, ModeBackendModel)
		return &Result{}, nil
	}
	if err := g.genModel(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateBackendEndpoint 生成后端单个 API 接口。
func (g *Generator) GenerateBackendEndpoint() (*Result, error) {
	if g.opts.DryRun {
		return &Result{Warnings: []string{"[dry-run] 将执行后端单接口生成，已跳过写盘"}}, nil
	}
	if err := g.genApi(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateFullstack 前后端全栈生成：先跑后端模块，再跑前端页面（共用同一 DB / 配置）。
// DryRun 时后端部分只打印将执行的步骤、不落盘，前端部分按 dry-run 收集文件清单。
func (g *Generator) GenerateFullstack() (*Result, error) {
	res := &Result{}

	// 一次性分析表结构（后端 + 前端 + 指纹共用）
	if g.db == nil {
		return nil, fmt.Errorf("db is nil")
	}
	_, analysisRes, err := g.buildTableMeta()
	if err != nil {
		return nil, err
	}
	// 增量：结构未变更则整体跳过
	if skip, why := g.skipIfUnchanged(analysisRes, ModeFullstack); skip {
		return &Result{Warnings: []string{why}}, nil
	}
	g.snapshot()

	// 后端模块（DryRun 时跳过实际写盘）
	if g.opts.DryRun {
		res.Warnings = append(res.Warnings, "[dry-run] 将执行后端模块生成（model/dao/object/service/controller/dto/router/code），已跳过写盘")
	} else {
		if err := g.genModule(); err != nil {
			return nil, err
		}
		res.WrittenFiles = append(res.WrittenFiles, "(backend module written)")
	}

	// 前端页面
	fres, err := g.GenerateFrontendPage()
	if err != nil {
		return nil, err
	}
	res.WrittenFiles = append(res.WrittenFiles, fres.WrittenFiles...)
	res.Warnings = append(res.Warnings, fres.Warnings...)

	g.recordManifest(analysisRes, ModeFullstack)
	return res, nil
}
