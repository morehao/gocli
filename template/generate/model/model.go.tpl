package {{.ModelLayerName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}
	{{- if isNumID .PKFieldType}}
	"gorm.io/gorm"
	{{- else if hasTimeFieldAny .ModelFields}}
	"time"
	{{- end}}
)

// {{.StructName}}Entity {{.Description}}表结构体
type {{.StructName}}Entity struct {
{{- if isNumID .PKFieldType}}
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
	{{- if $field.IsPrimaryKey}}{{$tagStr = printf "%s;primaryKey" $tagStr}}{{end}}
	{{.FieldName}} {{.FieldType}} `gorm:"{{$tagStr}}"`
	{{- end}}
{{- end}}
{{- else}}
{{- range .ModelFields}}
	{{- $field := .}}
	{{- $tagStr := ""}}
	{{- $tagStr = printf "%scolumn:%s" $tagStr $field.ColumnName}}
	{{- $tagStr = printf "%s;type:%s" $tagStr $field.ColumnType}}
	{{- if $field.NullableDesc}}{{$tagStr = printf "%s;%s" $tagStr $field.NullableDesc}}{{end}}
	{{- if $field.DefaultValue}}{{$tagStr = printf "%s;%s" $tagStr $field.DefaultValue}}{{end}}
	{{- if $field.IndexName}}{{$tagStr = printf "%s;index:%s" $tagStr $field.IndexName}}{{end}}
	{{- if and $field.IndexName $field.IsUniqueIndex}}{{$tagStr = printf "%s;uniqueIndex" $tagStr}}{{end}}
	{{- if $field.Comment}}{{$tagStr = printf "%s;comment:%s" $tagStr $field.Comment}}{{end}}
	{{- if $field.IsPrimaryKey}}{{$tagStr = printf "%s;primaryKey" $tagStr}}{{end}}
	{{.FieldName}} {{.FieldType}} `gorm:"{{$tagStr}}"`
{{- end}}
{{- end}}
}

type {{.StructName}}EntityList []{{.StructName}}Entity

func ({{.StructName}}Entity ) TableName() string {
  return TableName{{.StructName}}
}

func (l {{.StructName}}EntityList) ToMap() map[{{.PKFieldType}}]{{.StructName}}Entity {
	m := make(map[{{.PKFieldType}}]{{.StructName}}Entity)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}
