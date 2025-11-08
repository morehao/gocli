package svc{{.PackageName}}

import (
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/dto/dto{{.PackageName}}"

	"github.com/gin-gonic/gin"
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
