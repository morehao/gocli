package svc{{.PackageName}}

import (
	"github.com/gin-gonic/gin"
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/dto/dto{{.PackageName}}"
)

{{if not .TargetFileExist}}
type {{.StructName}}Svc interface {
    {{.FunctionName}}(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req) (*dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp, error)
}

type {{.StructNameLowerCamel}}Svc struct {
}

var _ {{.StructName}}Svc = (*{{.StructNameLowerCamel}}Svc)(nil)

func New{{.StructName}}Svc() {{.StructName}}Svc {
    return &{{.StructNameLowerCamel}}Svc{
    }
}
{{end}}
func (svc *{{.StructNameLowerCamel}}Svc) {{.FunctionName}}(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req) (*dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp, error) {
    return &dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp{}, nil
}
