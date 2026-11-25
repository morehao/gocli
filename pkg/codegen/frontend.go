package codegen

import (
	"fmt"
	"path/filepath"
	"text/template"

	golibcodegen "github.com/morehao/golib/codegen"
	"github.com/morehao/golib/gutil"

	"github.com/morehao/gocli/pkg/codegen/webgen"
)

// GenerateFrontendPage 生成前端页面（types + api 函数 + 页面组件 + 可选菜单注册）。
func (g *Generator) GenerateFrontendPage() (*Result, error) {
	tm, analysisRes, err := g.buildTableMeta()
	if err != nil {
		return nil, err
	}
	// 用真实字段（golib ModelField）再组装一次 webgen.TableMeta（结构名/包名以 analysisRes 为准）
	_ = analysisRes

	wopts := webgen.Options{
		ProjectRoot:  g.opts.ProjectRoot,
		PageDir:      g.opts.Web.PageDir,
		MenuGroup:    g.opts.Web.MenuGroup,
		MenuLabel:    g.opts.Web.MenuLabel,
		RegisterMenu: g.opts.Web.RegisterMenu,
		Force:        g.opts.Force,
		DryRun:       g.opts.DryRun,
	}
	r := webgen.New(TemplatesFS, wopts)
	res, err := r.Render(tm)
	if err != nil {
		return nil, err
	}
	if err := r.RegisterMenu(tm, "", "", res); err != nil {
		return nil, err
	}
	return &Result{WrittenFiles: res.WrittenFiles, Warnings: res.Warnings}, nil
}

// buildTableMeta 连库分析表结构，构造 webgen.TableMeta。
func (g *Generator) buildTableMeta() (*webgen.TableMeta, *golibcodegen.ModuleTplAnalysisRes, error) {
	if g.db == nil {
		return nil, nil, fmt.Errorf("db is nil")
	}
	tbl := g.cfg.Module.TableName
	if tbl == "" {
		return nil, nil, fmt.Errorf("table name is required (set --table or code_gen.yaml module.table_name)")
	}
	tplDir, err := CopyEmbeddedTemplatesToTempDir(TemplatesFS, "generate/module")
	if err != nil {
		return nil, nil, err
	}
	// 注意：tplDir 仅用于分析元数据，渲染前端不依赖它，函数返回后可留待进程回收。
	analysisCfg := &golibcodegen.ModuleCfg{
		CommonConfig: golibcodegen.CommonConfig{
			PackageName:       g.cfg.Module.PackageName,
			TplDir:            tplDir,
			RootDir:           g.workDir,
			LayerParentDirMap: defaultLayerParentDirMap,
			LayerNameMap:      buildLayerNameMap(g.cfg.ServiceName),
			LayerPrefixMap:    defaultLayerPrefixMap,
			TplFuncMap:        template.FuncMap{},
		},
		TableName: tbl,
	}
	gen := golibcodegen.NewGenerator()
	analysisRes, err := gen.AnalysisModuleTpl(g.db, analysisCfg)
	if err != nil {
		return nil, nil, fmt.Errorf("analysis module tpl error: %w", err)
	}

	// 取 model 层的字段列表（各层字段一致，优先取 model 层）
	fields := analysisRes.TplAnalysisList[0].ModelFields
	for _, v := range analysisRes.TplAnalysisList {
		if v.OriginLayerName == golibcodegen.LayerNameModel {
			fields = v.ModelFields
			break
		}
	}

	adapters := make([]webgen.ModelFieldLike, 0, len(fields))
	for _, f := range fields {
		adapters = append(adapters, fieldAdapter{f})
	}

	comment := g.cfg.Module.Description
	if comment == "" {
		comment = analysisRes.StructName
	}
	tm := webgen.FromModelFields(
		analysisRes.TableName,
		analysisRes.StructName,
		gutil.CamelToSnakeCase(g.cfg.Module.PackageName),
		comment,
		adapters,
	)
	// 应用 web.field_options 覆盖（enum 选项）——简单起见，本期先留空实现。
	return tm, analysisRes, nil
}

// fieldAdapter 把 golib ModelField 适配为 webgen.ModelFieldLike。
type fieldAdapter struct {
	f golibcodegen.ModelField
}

func (a fieldAdapter) FieldName() string  { return gutil.ReplaceIdToID(a.f.FieldName) }
func (a fieldAdapter) JsonName() string   { return SnakeToLowerCamelWithID(a.f.ColumnName) }
func (a fieldAdapter) FieldType() string  { return a.f.FieldType }
func (a fieldAdapter) ColumnType() string { return a.f.ColumnType }
func (a fieldAdapter) Comment() string    { return a.f.Comment }
func (a fieldAdapter) IsNullable() bool   { return a.f.IsNullable }
func (a fieldAdapter) IsPrimaryKey() bool { return a.f.IsPrimaryKey }

var _ = filepath.Join
