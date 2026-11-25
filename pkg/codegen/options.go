package codegen

import "errors"

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
	Web         WebConfig
}

// Result 生成结果。
type Result struct {
	// WrittenFiles 本次写入的文件列表；本期暂不精确收集，后续任务补充。
	WrittenFiles []string
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
	if err := g.genModule(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateBackendModel 生成后端 model/dao/object 层。
func (g *Generator) GenerateBackendModel() (*Result, error) {
	if err := g.genModel(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateBackendEndpoint 生成后端单个 API 接口。
func (g *Generator) GenerateBackendEndpoint() (*Result, error) {
	if err := g.genApi(); err != nil {
		return nil, err
	}
	return &Result{}, nil
}

// GenerateFrontendPage 生成前端页面。本期未实现，返回 ErrNotImplemented。
// TODO: 由后续前端 webgen 任务实现。
func (g *Generator) GenerateFrontendPage() (*Result, error) {
	return nil, ErrNotImplemented
}

// GenerateFullstack 前后端全栈生成。本期未实现，返回 ErrNotImplemented。
// TODO: 由后续前端 webgen 任务实现。
func (g *Generator) GenerateFullstack() (*Result, error) {
	return nil, ErrNotImplemented
}
