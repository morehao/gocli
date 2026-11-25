// ==== {{.Comment}} {{.StructName}} ====
// 由 gocli 生成（generate frontend page / fullstack），区块内可安全重生成；请勿手改区块内容。
import type {
  {{.StructName}}PageListItem,
  {{.StructName}}Detail,
  {{.StructName}}CreateReq,
  {{.StructName}}UpdateReq,
} from '@ark-go-starter/types'

/** {{.Comment}}列表筛选参数：分页 + 可选筛选字段 */
export interface {{.StructName}}ListQuery extends PageQuery {
{{- range .Fields}}
{{- if and .IsSelectable (not .IsPK)}}
  // {{.Label}}（可选）
  {{.JsonName}}?: {{.TSType}}
{{- end}}
{{- end}}
}

/** {{.Comment}}分页列表 GET /{{.PackageName}}/{{.ResourceKebab}} */
export function {{.ListFn}}(params: {{.StructName}}ListQuery = {}) {
  return request<PageResult<{{.StructName}}PageListItem>>({
    url: '/{{.PackageName}}/{{.ResourceKebab}}',
    method: 'get',
    params: withPaging(params),
  })
}

/** {{.Comment}}详情 GET /{{.PackageName}}/{{.ResourceKebab}}/{id} */
export function {{.DetailFn}}(id: {{.PKTSType}}) {
  return request<{{.StructName}}Detail>({
    url: `/{{.PackageName}}/{{.ResourceKebab}}/${id}`,
    method: 'get',
  })
}

/** 新建{{.Comment}} POST /{{.PackageName}}/{{.ResourceKebab}} */
export function create{{.StructName}}(body: {{.StructName}}CreateReq) {
  return request<{{.StructName}}Detail>({
    url: '/{{.PackageName}}/{{.ResourceKebab}}',
    method: 'post',
    data: body,
  })
}

/** 更新{{.Comment}} PUT /{{.PackageName}}/{{.ResourceKebab}}/{id} */
export function update{{.StructName}}(id: {{.PKTSType}}, body: {{.StructName}}UpdateReq) {
  return request<unknown>({
    url: `/{{.PackageName}}/{{.ResourceKebab}}/${id}`,
    method: 'put',
    data: body,
  })
}

/** 删除{{.Comment}} DELETE /{{.PackageName}}/{{.ResourceKebab}}/{id} */
export function delete{{.StructName}}(id: {{.PKTSType}}) {
  return request<unknown>({
    url: `/{{.PackageName}}/{{.ResourceKebab}}/${id}`,
    method: 'delete',
  })
}
