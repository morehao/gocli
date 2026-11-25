// ==== {{.Comment}} {{.StructName}} ====
// 由 gocli 生成（generate frontend page / fullstack），区块内可安全重生成；请勿手改区块内容。

/** {{.Comment}}列表项 / 详情（对应后端 {{.StructName}}DetailResp） */
export interface {{.StructName}}PageListItem {
{{- range .Fields}}
  // {{.Label}}
  {{.JsonName}}{{if .IsPK}}: {{.TSType}}{{else}}?: {{.TSType}}{{end}}
{{- end}}
}

/** {{.Comment}}详情（与列表项形状一致） */
export type {{.StructName}}Detail = {{.StructName}}PageListItem

/** 新建{{.Comment}}请求体 */
export interface {{.StructName}}CreateReq {
{{- range .Fields}}
{{- if .IsPK}}{{else}}
  // {{.Label}}{{if .IsRequired}}（必填）{{end}}
  {{.JsonName}}{{if not .IsRequired}}?{{end}}: {{.TSType}}
{{- end}}
{{- end}}
}

/** 更新{{.Comment}}请求体（局部更新，未传字段不修改） */
export interface {{.StructName}}UpdateReq {
{{- range .Fields}}
{{- if .IsPK}}{{else}}
  // {{.Label}}
  {{.JsonName}}?: {{.TSType}}
{{- end}}
{{- end}}
}
