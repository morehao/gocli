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
	if !g.opts.Web.Enabled {
		return &Result{Warnings: []string{"web.enabled=false，跳过前端生成"}}, nil
	}
	tm, analysisRes, err := g.buildTableMeta()
	if err != nil {
		return nil, err
	}
	// 增量：结构未变更则跳过
	if skip, why := g.skipIfUnchanged(analysisRes, ModeFrontendPage); skip {
		return &Result{Warnings: []string{why}}, nil
	}
	g.snapshot()

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
	g.recordManifest(analysisRes, ModeFrontendPage)
	return &Result{WrittenFiles: res.WrittenFiles, Warnings: res.Warnings}, nil
}

// buildTableMeta 连库分析表结构，构造 webgen.TableMeta。
func (g *Generator) buildTableMeta() (*webgen.TableMeta, *golibcodegen.ModuleTplAnalysisRes, error) {
	return g.buildTableMetaFor(g.cfg.Module.TableName, g.cfg.Module.PackageName)
}

// buildTableMetaFor 针对指定表/包分析结构（供前端渲染与后端增量指纹共用）。
func (g *Generator) buildTableMetaFor(tbl, pkgName string) (*webgen.TableMeta, *golibcodegen.ModuleTplAnalysisRes, error) {
	if g.db == nil {
		return nil, nil, fmt.Errorf("db is nil")
	}
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
			PackageName:       pkgName,
			TplDir:            tplDir,
			RootDir:           g.workDir,
			LayerParentDirMap: defaultLayerParentDirMap,
			LayerNameMap:      buildLayerNameMap(g.cfg.ServiceName),
			LayerPrefixMap:    defaultLayerPrefixMap,
			TplFuncMap:        moduleTplFuncMap(),
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
	ov := webgen.Overrides{
		Widgets: g.opts.Web.FieldWidgets,
		Options: convertFieldOptions(g.opts.Web.FieldOptions),
	}
	tm := webgen.FromModelFields(
		analysisRes.TableName,
		analysisRes.StructName,
		gutil.CamelToSnakeCase(pkgName),
		comment,
		adapters,
		ov,
	)
	// 选项的 TS 字面量风格按对应字段的 TS 类型推断
	inferOptionStyles(tm)
	return tm, analysisRes, nil
}

// convertFieldOptions 把 WebConfig.FieldOptions（[]KV）转成 webgen.Option。
func convertFieldOptions(in map[string][]KV) map[string][]webgen.Option {
	if len(in) == 0 {
		return nil
	}
	out := make(map[string][]webgen.Option, len(in))
	for k, list := range in {
		opts := make([]webgen.Option, 0, len(list))
		for _, kv := range list {
			opts = append(opts, webgen.Option{Value: fmt.Sprintf("%v", kv.Value), Label: kv.Label})
		}
		out[k] = opts
	}
	return out
}

// inferOptionStyles 依据字段 TSType 给其 Option 设置 TS 字面量风格（number/boolean/string）。
func inferOptionStyles(tm *webgen.TableMeta) {
	for i := range tm.Fields {
		style := "string"
		switch tm.Fields[i].TSType {
		case "number":
			style = "number"
		case "boolean":
			style = "boolean"
		}
		for j := range tm.Fields[i].Options {
			if tm.Fields[i].Options[j].Style == "" {
				tm.Fields[i].Options[j].Style = style
			}
		}
	}
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

// moduleTplFuncMap 返回 module 模板所需的全部模板函数（前端分析复用同一套，避免解析报错）。
func moduleTplFuncMap() template.FuncMap {
	return template.FuncMap{
		TplFuncIsBuiltInField:      IsBuiltInField,
		TplFuncIsSysField:          IsSysField,
		TplFuncIsDefaultModelLayer: IsDefaultModelLayer,
		TplFuncIsDefaultDaoLayer:   IsDefaultDaoLayer,
		TplFuncHasTimeField:        HasTimeField,
		TplFuncGetFieldImports:     GetFieldImports,
		TplFuncIsBasicType:         IsBasicType,
		TplFuncToKebabCase:         toKebabCase,
		TplFuncPluralize:           pluralize,
		TplFuncIsNumID:             IsNumID,
		TplFuncIsStringID:          IsStringID,
		TplFuncIsIntID:             IsIntID,
		TplFuncHasTimeFieldAny:     HasTimeFieldAny,
		TplFuncHasSysField:         HasSysField,
	}
}
