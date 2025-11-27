package ctr{{.PackageName}}

import (
    "github.com/gin-gonic/gin"
    "{{.ModulePath}}/{{.AppPathInProject}}/internal/dto/dto{{.PackageName}}"
    "{{.ModulePath}}/{{.AppPathInProject}}/internal/service/svc{{.PackageName}}"
    "github.com/morehao/golib/biz/gcontext/gincontext"
)
{{if not .TargetFileExist}}
type {{.StructName}}Ctr interface {
	{{.FunctionName}}(ctx *gin.Context)
}

type {{.StructNameLowerCamel}}Ctr struct {
	{{.StructNameLowerCamel}}Svc svc{{.PackageName}}.{{.StructName}}Svc
}

var _ {{.StructName}}Ctr = (*{{.StructNameLowerCamel}}Ctr)(nil)

func New{{.StructName}}Ctr() {{.StructName}}Ctr {
	return &{{.StructNameLowerCamel}}Ctr{
		{{.StructNameLowerCamel}}Svc: svc{{.PackageName}}.New{{.StructName}}Svc(),
	}
}
{{end}}
{{if eq .HttpMethod "POST"}}
// {{.FunctionName}} {{.Description}}
// @Tags {{.ApiDocTag}}
// @Summary {{.Description}}
// @accept application/json
// @Produce application/json
// @Param req body dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req true "{{.Description}}"
// @Success 200 {object} gincontext.DtoRender{data=dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp} "{"code": 0,"data": "ok","msg": "success"}"
// @Router /{{.AppName}}/v1/{{.StructNameLowerCamel}}/{{.FunctionNameLowerCamel}} [post]
func (ctr *{{.StructNameLowerCamel}}Ctr) {{.FunctionName}}(ctx *gin.Context) {
	var req dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req
	if err := ctx.ShouldBindJSON(&req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	res, err := ctr.{{.StructNameLowerCamel}}Svc.{{.FunctionName}}(ctx, &req)
	if err != nil {
		gincontext.Fail(ctx, err)
		return
	} else {
		gincontext.Success(ctx, res)
	}
}
{{else if eq .HttpMethod "GET"}}
// {{.FunctionName}} {{.Description}}
// @Tags {{.ApiDocTag}}
// @Summary {{.Description}}
// @accept application/json
// @Produce application/json
// @Param req query dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req true "{{.Description}}"
// @Success 200 {object} gincontext.DtoRender{data=dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp} "{"code": 0,"data": "ok","msg": "success"}"
// @Router /{{.AppName}}/v1/{{.StructNameLowerCamel}}/{{.FunctionNameLowerCamel}} [get]
func (ctr *{{.StructNameLowerCamel}}Ctr){{.FunctionName}}(ctx *gin.Context) {
	var req dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req
	if err := ctx.ShouldBindQuery(&req); err != nil {
		gincontext.Fail(ctx, err)
		return
	}
	res, err := ctr.{{.StructNameLowerCamel}}Svc.{{.FunctionName}}(ctx, &req)
	if err != nil {
		gincontext.Fail(ctx, err)
		return
	} else {
		gincontext.Success(ctx, res)
	}
}
{{end}}
