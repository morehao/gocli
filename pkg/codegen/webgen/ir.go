// Package webgen 负责从表元数据生成前端产物（TS 类型 / axios api 函数 / Ant Design
// 页面组件 / App.tsx 菜单注册片段）。它是 pkg/codegen 的前端渲染子模块，供 SDK 与 CLI 复用。
package webgen

import (
	"strings"

	gutil "github.com/morehao/golib/gutil"
)

// frontendSysFields 是前端默认不在表单/列表里展示的系统字段（与后端 sysField 对齐）。
var frontendSysFields = map[string]struct{}{
	"CreatedAt": {}, "UpdatedAt": {}, "DeletedAt": {},
	"CreatedBy": {}, "UpdatedBy": {}, "DeletedBy": {},
}

// TableMeta 是从后端元数据派生、专供前端模板渲染的中间模型。
// 字段直接复用 golib/codegen.ModelField（见 FromFields）。
type TableMeta struct {
	TableName        string  // 表名，如 product
	StructName       string  // PascalCase，如 Product
	PackageName      string  // 生成器包名（用于 api url 前缀资源名），如 product
	Comment          string  // 表/模块描述（页面 title/description）
	PKName           string  // 主键 Go 字段名（沿用改名后的形式，如 ProductID / ID）
	PKJsonName       string  // 主键 JSON 字段名（如 productID / id）
	PKTSType         string  // 主键 TS 类型（number / string）
	Fields           []Field // 业务字段（去掉 sys 字段、转换好 TS 类型与控件）
	ResourceKebab    string  // 资源路径，kebab 复数，如 products
	StructLowerCamel string  // 小驼峰，如 product
}

// Field 是单个前端字段的渲染信息。
type Field struct {
	GoFieldName  string // Go/Pascal 字段名，如 CategoryID
	JsonName     string // JSON 字段名，如 categoryID
	Label        string // 中文标签（取注释，缺省用字段名）
	TSType       string // TS 类型，如 number / string
	Widget       string // antd 控件：input / textarea / number / switch / date / select
	IsPK         bool   // 是否主键
	IsRequired   bool   // 是否表单必填（NOT NULL 且非 sys、非自增主键）
	IsSelectable bool   // 是否作为列表筛选/搜索条件（默认：短字符串、枚举）
	Options      []Option
}

// Option 是 Select 选项。Value/Label 为原始字符串，TSValue() 依据 Style 生成 TS 字面量。
type Option struct {
	Value string
	Label string
	Style string // TS 值字面量风格：number / string / boolean（缺省 string）
}

// TSValue 返回该选项在 TSX 中的字面量表示（数字/布尔裸值，字符串加引号）。
func (o Option) TSValue() string {
	switch o.Style {
	case "number", "boolean":
		return o.Value
	default:
		return "'" + strings.ReplaceAll(o.Value, "'", "\\'") + "'"
	}
}

// Overrides 是字段级覆盖（来自 code_gen.yaml 的 web 段）。
type Overrides struct {
	Widgets map[string]string   // 列 JSON 名 → 控件（input/textarea/number/switch/date/select）
	Options map[string][]Option // 列 JSON 名 → Select 选项
}

// FromModelFields 把后端 []ModelFieldLike 组装成 TableMeta。
// ov 提供字段级覆盖（web.field_widgets / web.field_options）；可为零值。
func FromModelFields(tableName, structName, packageName, comment string, fields []ModelFieldLike, ov Overrides) *TableMeta {
	structLowerCamel := gutil.FirstLetterToLower(structName)
	resourceKebab := toKebabCase(pluralize(structLowerCamel))

	tm := &TableMeta{
		TableName:        tableName,
		StructName:       structName,
		PackageName:      packageName,
		Comment:          comment,
		StructLowerCamel: structLowerCamel,
		ResourceKebab:    resourceKebab,
	}

	for _, f := range fields {
		if _, isSys := frontendSysFields[f.FieldName()]; isSys {
			continue
		}
		jsonName := f.JsonName()
		tsType := TSForField(f.ColumnType(), f.FieldType())
		widget := WidgetForField(f.ColumnType(), f.FieldType(), jsonName)
		// 覆盖：web.field_options 提供选项时强制 select
		var opts []Option
		if ov.Options != nil {
			if o, ok := ov.Options[jsonName]; ok && len(o) > 0 {
				opts = o
				widget = "select"
			}
		}
		// 覆盖：web.field_widgets 显式指定控件
		if ov.Widgets != nil {
			if w, ok := ov.Widgets[jsonName]; ok && w != "" {
				widget = w
			}
		}
		label := f.Comment()
		if label == "" {
			label = f.FieldName()
		}
		fld := Field{
			GoFieldName:  f.FieldName(),
			JsonName:     jsonName,
			Label:        label,
			TSType:       tsType,
			Widget:       widget,
			IsPK:         f.IsPrimaryKey(),
			IsRequired:   !f.IsNullable() && !f.IsPrimaryKey(),
			IsSelectable: (widget == "input" || widget == "select") && !f.IsPrimaryKey(),
			Options:      opts,
		}
		tm.Fields = append(tm.Fields, fld)
		if f.IsPrimaryKey() {
			tm.PKName = f.FieldName()
			tm.PKJsonName = jsonName
			tm.PKTSType = tsType
		}
	}
	return tm
}

// ModelFieldLike 收窄对 golib ModelField 的依赖，便于 webgen 单测与解耦。
type ModelFieldLike interface {
	FieldName() string
	JsonName() string
	FieldType() string
	ColumnType() string
	Comment() string
	IsNullable() bool
	IsPrimaryKey() bool
}

// TSForField 依据列类型 + Go 类型给出前端 TS 类型。
// 特殊：numeric/decimal 在 Go 侧映射为 string（保精度），但前端表单/展示更宜用 number；
// 这里按列类型优先判定为 number。
func TSForField(columnType, goType string) string {
	lower := strings.ToLower(columnType)
	if strings.HasPrefix(lower, "numeric") || strings.HasPrefix(lower, "decimal") {
		return "number"
	}
	return GoTypeToTS(goType)
}

// GoTypeToTS 把 Go 字段类型映射为 TS 类型。
func GoTypeToTS(goType string) string {
	switch goType {
	case "string":
		return "string"
	case "bool":
		return "boolean"
	case "int", "int8", "int16", "int32", "int64", "uint", "uint8", "uint16", "uint32", "uint64", "float32", "float64", "time.Time", "time.Duration":
		return "number"
	case "json.RawMessage":
		return "Record<string, unknown>"
	case "[]byte":
		return "string"
	default:
		if strings.HasPrefix(goType, "[]") {
			return "unknown[]"
		}
		return "string"
	}
}

// WidgetForField 依据列类型/Go 类型推断 antd 控件种类。
func WidgetForField(columnType, goType, jsonName string) string {
	lower := strings.ToLower(columnType)
	switch {
	case strings.HasPrefix(lower, "enum"):
		return "select"
	case strings.HasPrefix(lower, "numeric"), strings.HasPrefix(lower, "decimal"):
		return "number"
	case goType == "bool":
		return "switch"
	case goType == "time.Time":
		return "date"
	case goType == "json.RawMessage":
		return "textarea"
	case goType == "string":
		if strings.Contains(lower, "text") {
			return "textarea"
		}
		// 命名约定：*status / *type / *state / *kind / *level 视作枚举（后续可由 web.field_options 覆盖）
		ln := strings.ToLower(jsonName)
		for _, s := range []string{"status", "state", "type", "kind", "level"} {
			if strings.HasSuffix(ln, s) {
				return "select"
			}
		}
		return "input"
	default:
		return "number"
	}
}

// ---- 命名工具（与 pkg/codegen 包内的 toKebabCase/pluralize 保持一致，此处本地实现避免包循环） ----

func toKebabCase(s string) string {
	return strings.ReplaceAll(gutil.CamelToSnakeCase(s), "_", "-")
}

func pluralize(s string) string {
	if s == "" {
		return s
	}
	switch {
	case strings.HasSuffix(s, "y") && len(s) > 2 && !isVowel(s[len(s)-2]):
		return s[:len(s)-1] + "ies"
	case strings.HasSuffix(s, "s"), strings.HasSuffix(s, "x"), strings.HasSuffix(s, "z"), strings.HasSuffix(s, "ch"), strings.HasSuffix(s, "sh"):
		return s + "es"
	default:
		return s + "s"
	}
}

func isVowel(c byte) bool {
	switch c {
	case 'a', 'e', 'i', 'o', 'u':
		return true
	}
	return false
}
