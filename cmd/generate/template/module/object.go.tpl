package obj{{.PackageName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}
)

type {{.StructName}}BaseInfo struct {
{{- range .ModelFields}}
{{- if isSysField .FieldName}}
    {{- continue}}
{{- end}}

{{- if eq .FieldType "time.Time"}}
    {{.FieldName}} int64 `json:"{{.JsonTagName}}" form:"{{.JsonTagName}}"` // {{.Comment}}
{{- else}}
    {{.FieldName}} {{.FieldType}} `json:"{{.JsonTagName}}" form:"{{.JsonTagName}}"` // {{.Comment}}
{{- end}}
{{- end}}
}
