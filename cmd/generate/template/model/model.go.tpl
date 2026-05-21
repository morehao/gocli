package {{.ModelLayerName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}

	"gorm.io/gorm"
)

// {{.StructName}}Entity {{.Description}}表结构体
type {{.StructName}}Entity struct {
    gorm.Model
{{- range .ModelFields}}
    {{- if isBuiltInField .FieldName}}
        {{- continue}}
    {{- else}}
	{{- $field := .}}
	{{- $tagStr := ""}}
	{{- $tagStr = printf "%scolumn:%s" $tagStr $field.ColumnName}}
	{{- $tagStr = printf "%s;type:%s" $tagStr $field.ColumnType}}
	{{- if $field.NullableDesc}}{{$tagStr = printf "%s;%s" $tagStr $field.NullableDesc}}{{end}}
	{{- if $field.DefaultValue}}{{$tagStr = printf "%s;%s" $tagStr $field.DefaultValue}}{{end}}
	{{- if $field.IndexName}}{{$tagStr = printf "%s;index:%s" $tagStr $field.IndexName}}{{end}}
	{{- if and $field.IndexName $field.IsUniqueIndex}}{{$tagStr = printf "%s;uniqueIndex" $tagStr}}{{end}}
	{{- if $field.Comment}}{{$tagStr = printf "%s;comment:%s" $tagStr $field.Comment}}{{end}}
	{{.FieldName}} {{.FieldType}} `gorm:"{{$tagStr}}"`
	{{- end}}
{{- end}}
}

type {{.StructName}}EntityList []{{.StructName}}Entity

func ({{.StructName}}Entity ) TableName() string {
  return TableName{{.StructName}}
}

func (l {{.StructName}}EntityList) ToMap() map[uint]{{.StructName}}Entity {
	m := make(map[uint]{{.StructName}}Entity)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}