package obj{{.PackageName}}

type {{.StructName}}BaseInfo struct {
{{- range .ModelFields}}
{{- if isSysField .FieldName}}
    {{- continue}}
{{- end}}
{{- if eq .FieldType "time.Time"}}
    // {{.FieldName}} {{.Comment}}
    {{.FieldName}} int64 `json:"{{.FieldLowerCaseName}}" form:"{{.FieldLowerCaseName}}"`
{{- else}}
    // {{.FieldName}} {{.Comment}}
    {{.FieldName}} {{.FieldType}} `json:"{{.FieldLowerCaseName}}" form:"{{.FieldLowerCaseName}}"`
{{- end}}
{{- end}}
}
