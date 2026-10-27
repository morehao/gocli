package svc{{.PackageName}}

import (
	"context"

	"{{.BaseModulePath}}/{{.AppModuleName}}/internal/dto/dto{{.PackageName}}"
)

{{if not .TargetFileExist}}
type {{.StructName}}Svc interface {
    {{.FunctionName}}(ctx context.Context, req *dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req) (*dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp, error)
}

type {{.StructNameLowerCamel}}Svc struct {
}

var _ {{.StructName}}Svc = (*{{.StructNameLowerCamel}}Svc)(nil)

func New{{.StructName}}Svc() {{.StructName}}Svc {
    return &{{.StructNameLowerCamel}}Svc{
    }
}
{{end}}
func (svc *{{.StructNameLowerCamel}}Svc) {{.FunctionName}}(ctx context.Context, req *dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Req) (*dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp, error) {
    return &dto{{.PackageName}}.{{.StructName}}{{.FunctionName}}Resp{}, nil
}
