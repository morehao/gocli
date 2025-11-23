package obj{{.PackageName}}

type {{.StructName}}BaseInfo struct {
{{- range .ModelFields}}
{{- if isSysField .FieldName}}
    {{- continue}}
{{- end}}
{{- if eq .FieldType "time.Time"}}
    // {{.FieldName}} {{.Comment}}
    {{.FieldName}} int64 `json:"{{.JsonTagName}}" form:"{{.JsonTagName}}"`
{{- else}}
    // {{.FieldName}} {{.Comment}}
    {{.FieldName}} {{.FieldType}} `json:"{{.JsonTagName}}" form:"{{.JsonTagName}}"`
{{- end}}
{{- end}}
}
