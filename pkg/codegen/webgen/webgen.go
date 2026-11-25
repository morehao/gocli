package webgen

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"text/template"

	"github.com/morehao/golib/gutil"
)

// TplFS 由调用方注入的模板文件系统（顶层 template.GenerateFS）。用 fs.FS 抽象避免循环依赖。
type TplFS interface {
	ReadFile(name string) ([]byte, error)
}

// Options 前端生成选项（由 pkg/codegen.Options 映射而来）。
type Options struct {
	ProjectRoot  string // 目标工程根
	PageDir      string // 页面目录，缺省 apps/web/src/pages
	TypesFile    string // types 汇总文件，缺省 packages/types/src/index.ts
	ApiFile      string // api 汇总文件，缺省 packages/api/src/index.ts
	AppFile      string // App.tsx，缺省 apps/web/src/App.tsx
	MenuGroup    string
	MenuLabel    string
	RegisterMenu bool
	Force        bool
	DryRun       bool
}

// Result 前端生成结果。
type Result struct {
	WrittenFiles []string // 实际写入/追加的文件
	Warnings     []string // 非致命提示（如无锚点跳过菜单注册）
}

// Renderer 前端渲染器。
type Renderer struct {
	fs   TplFS
	opts Options
}

// New 构造 Renderer。
func New(fs TplFS, opts Options) *Renderer {
	if opts.PageDir == "" {
		opts.PageDir = filepath.Join("apps", "web", "src", "pages")
	}
	if opts.TypesFile == "" {
		opts.TypesFile = filepath.Join("packages", "types", "src", "index.ts")
	}
	if opts.ApiFile == "" {
		opts.ApiFile = filepath.Join("packages", "api", "src", "index.ts")
	}
	if opts.AppFile == "" {
		opts.AppFile = filepath.Join("apps", "web", "src", "App.tsx")
	}
	return &Renderer{fs: fs, opts: opts}
}

// Render 生成 types / api / page 三件产物（menu 由 RegisterMenu 控制，另行处理）。
func (r *Renderer) Render(tm *TableMeta) (*Result, error) {
	res := &Result{}

	// 1) types 追加
	typesBlock, err := r.renderTpl("types.tpl", tm)
	if err != nil {
		return nil, err
	}
	if err := r.appendTypes(typesBlock, tm, res); err != nil {
		return nil, err
	}

	// 2) api 追加
	apiBlock, err := r.renderTpl("api.tpl", tm)
	if err != nil {
		return nil, err
	}
	if err := r.appendApi(apiBlock, tm, res); err != nil {
		return nil, err
	}

	// 3) page 整文件
	pageBody, err := r.renderTpl("page.tpl", buildPageCtx(tm))
	if err != nil {
		return nil, err
	}
	if err := r.writePage(pageBody, tm, res); err != nil {
		return nil, err
	}

	return res, nil
}

// RegisterMenu 把页面注册进 App.tsx 的锚点（imports / pages / menu）。
func (r *Renderer) RegisterMenu(tm *TableMeta, menuKey, menuLabel string, res *Result) error {
	if !r.opts.RegisterMenu {
		return nil
	}
	if menuKey == "" {
		menuKey = tm.ResourceKebab
	}
	if menuLabel == "" {
		menuLabel = r.opts.MenuLabel
	}
	if menuLabel == "" {
		menuLabel = tm.Comment
	}
	appPath := filepath.Join(r.opts.ProjectRoot, r.opts.AppFile)
	data, err := os.ReadFile(appPath)
	if err != nil {
		res.Warnings = append(res.Warnings, fmt.Sprintf("跳过菜单注册：读取 %s 失败：%v", appPath, err))
		return nil
	}
	src := string(data)

	// 幂等：已存在该 key / import 则跳过
	if strings.Contains(src, "key: '"+menuKey+"'") || strings.Contains(src, "./pages/"+tm.StructName+`'`) {
		res.Warnings = append(res.Warnings, fmt.Sprintf("菜单项 %s 已存在，跳过注册", menuKey))
		return nil
	}

	imps, pages, menus, err := r.renderMenuFragments(tm, menuKey, menuLabel)
	if err != nil {
		return err
	}

	const (
		anchorImports = "// <codegen:imports>"
		anchorPages   = "// <codegen:pages>"
		anchorMenu    = "// <codegen:menu>"
	)
	if !strings.Contains(src, anchorImports) || !strings.Contains(src, anchorPages) || !strings.Contains(src, anchorMenu) {
		res.Warnings = append(res.Warnings, "App.tsx 缺少 <codegen:*> 锚点，跳过菜单注册（请按文档补锚点）")
		return nil
	}

	out := insertAfterLine(src, anchorImports, imps)
	out = insertAfterLine(out, anchorPages, pages)
	out = insertAfterLine(out, anchorMenu, menus)

	if r.opts.DryRun {
		res.WrittenFiles = append(res.WrittenFiles, appPath+" (dry-run)")
		return nil
	}
	if err := os.WriteFile(appPath, []byte(out), 0o644); err != nil {
		return fmt.Errorf("写入 App.tsx 失败: %w", err)
	}
	res.WrittenFiles = append(res.WrittenFiles, appPath)
	return nil
}

// ---- 内部实现 ----

func (r *Renderer) renderTpl(name string, data any) (string, error) {
	raw, err := r.fs.ReadFile("generate/web/" + name)
	if err != nil {
		return "", fmt.Errorf("读取模板 %s 失败: %w", name, err)
	}
	tpl, err := template.New(name).Funcs(template.FuncMap{
		"toLower": strings.ToLower,
	}).Parse(string(raw))
	if err != nil {
		return "", fmt.Errorf("解析模板 %s 失败: %w", name, err)
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, data); err != nil {
		return "", fmt.Errorf("渲染模板 %s 失败: %w", name, err)
	}
	return buf.String(), nil
}

func (r *Renderer) appendTypes(block string, tm *TableMeta, res *Result) error {
	path := filepath.Join(r.opts.ProjectRoot, r.opts.TypesFile)
	return r.appendBlock(path, "interface "+tm.StructName+"PageListItem", block, res)
}

func (r *Renderer) appendApi(block string, tm *TableMeta, res *Result) error {
	path := filepath.Join(r.opts.ProjectRoot, r.opts.ApiFile)
	return r.appendBlock(path, "function get"+tm.StructName+"s", block, res)
}

// appendBlock 幂等追加：若 marker 已存在则跳过，否则在文件末尾追加 block。
func (r *Renderer) appendBlock(path, marker, block string, res *Result) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("读取 %s 失败: %w", path, err)
	}
	if strings.Contains(string(data), marker) {
		res.Warnings = append(res.Warnings, fmt.Sprintf("%s 中已存在 %q，跳过追加", filepath.Base(path), marker))
		return nil
	}
	if r.opts.DryRun {
		res.WrittenFiles = append(res.WrittenFiles, path+" (dry-run)")
		return nil
	}
	content := string(data)
	if !strings.HasSuffix(content, "\n") {
		content += "\n"
	}
	content += "\n" + block
	if err := os.WriteFile(path, []byte(content), 0o644); err != nil {
		return fmt.Errorf("写入 %s 失败: %w", path, err)
	}
	res.WrittenFiles = append(res.WrittenFiles, path)
	return nil
}

func (r *Renderer) writePage(body string, tm *TableMeta, res *Result) error {
	dir := filepath.Join(r.opts.ProjectRoot, r.opts.PageDir, tm.StructName)
	path := filepath.Join(dir, "index.tsx")
	if gutil.FileExists(path) && !r.opts.Force {
		res.Warnings = append(res.Warnings, fmt.Sprintf("页面已存在，跳过（--force 覆盖）: %s", path))
		return nil
	}
	if r.opts.DryRun {
		res.WrittenFiles = append(res.WrittenFiles, path+" (dry-run)")
		return nil
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		return fmt.Errorf("创建页面目录失败: %w", err)
	}
	if err := os.WriteFile(path, []byte(body), 0o644); err != nil {
		return fmt.Errorf("写入页面失败: %w", err)
	}
	res.WrittenFiles = append(res.WrittenFiles, path)
	return nil
}

func (r *Renderer) renderMenuFragments(tm *TableMeta, menuKey, menuLabel string) (imports, pages, menus string, err error) {
	raw, err := r.fs.ReadFile("generate/web/menu.tpl")
	if err != nil {
		return "", "", "", fmt.Errorf("读取 menu 模板失败: %w", err)
	}
	tpl, err := template.New("menu").Parse(string(raw))
	if err != nil {
		return "", "", "", err
	}
	var buf bytes.Buffer
	if err := tpl.Execute(&buf, map[string]string{
		"StructName": tm.StructName,
		"MenuKey":    menuKey,
		"MenuLabel":  menuLabel,
	}); err != nil {
		return "", "", "", err
	}
	var imps, pgs, mns []string
	for _, line := range strings.Split(buf.String(), "\n") {
		switch {
		case strings.HasPrefix(line, "#IMPORT#"):
			imps = append(imps, strings.TrimPrefix(line, "#IMPORT#"))
		case strings.HasPrefix(line, "#PAGES#"):
			pgs = append(pgs, strings.TrimPrefix(line, "#PAGES#"))
		case strings.HasPrefix(line, "#MENU#"):
			mns = append(mns, strings.TrimPrefix(line, "#MENU#"))
		}
	}
	return strings.Join(imps, "\n"), strings.Join(pgs, "\n"), strings.Join(mns, "\n"), nil
}

// insertAfterLine 在包含 anchor 的那一行之后插入 content。
func insertAfterLine(src, anchor, content string) string {
	lines := strings.Split(src, "\n")
	for i, ln := range lines {
		if strings.Contains(ln, anchor) {
			insert := content
			newLines := make([]string, 0, len(lines)+2)
			newLines = append(newLines, lines[:i+1]...)
			newLines = append(newLines, insert)
			newLines = append(newLines, lines[i+1:]...)
			return strings.Join(newLines, "\n")
		}
	}
	return src
}

// pageCtx 是渲染 page.tpl 的上下文：TableMeta + 依字段推导出的「需要哪些 antd 组件」标记，
// 用于条件裁剪 import，避免未使用 import 触发 TS 报错。
type pageCtx struct {
	*TableMeta
	NeedDatePicker  bool
	NeedSelect      bool
	NeedSwitch      bool
	NeedInputNumber bool
	NeedTextArea    bool
	NeedFmtTime     bool
	NeedTag         bool
	NeedSearch      bool
}

func buildPageCtx(tm *TableMeta) *pageCtx {
	c := &pageCtx{TableMeta: tm}
	for _, f := range tm.Fields {
		switch f.Widget {
		case "date":
			c.NeedDatePicker = true
			c.NeedFmtTime = true
		case "select":
			c.NeedTag = true
			if !f.IsPK {
				c.NeedSelect = true
			}
		case "switch":
			if !f.IsPK {
				c.NeedSwitch = true
			}
		case "number":
			// InputNumber 只在表单里用；主键不进表单
			if !f.IsPK {
				c.NeedInputNumber = true
			}
		case "textarea":
			if !f.IsPK {
				c.NeedTextArea = true
			}
		}
	}
	for _, f := range tm.Fields {
		if f.IsSelectable && !f.IsPK {
			c.NeedSearch = true
			break
		}
	}
	// 列表列渲染里 date 列才会用到 fmtTime（同上，date→NeedFmtTime）
	return c
}
