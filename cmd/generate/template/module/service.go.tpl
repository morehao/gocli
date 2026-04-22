package svc{{.PackageName}}

import (
	{{- range .FieldImports}}
	"{{.}}"
	{{- end}}

	"github.com/gin-gonic/gin"
	"github.com/morehao/golib/biz/gcontext/gincontext"
	"github.com/morehao/golib/biz/genericdao"
	"github.com/morehao/golib/biz/gobject"
	"github.com/morehao/golib/glog"
	"github.com/morehao/golib/gutil"
	"{{.ModulePath}}/{{.AppPathInProject}}/{{.DaoPackageName}}"
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/dto/dto{{.PackageName}}"
    {{- if isDefaultModelLayer .ModelLayerName}}
    "{{.ModulePath}}/{{.AppPathInProject}}/model"
    {{- else}}
    "{{.ModulePath}}/{{.AppPathInProject}}/{{.ModelLayerName}}"
    {{- end}}
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"{{.ModulePath}}/pkg/code"
)

type {{.StructName}}Svc interface {
	Create(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}CreateReq) (*dto{{.PackageName}}.{{.StructName}}CreateResp, error)
	Delete(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DeleteReq) error
	Update(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}UpdateReq) error
	Detail(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DetailReq) (*dto{{.PackageName}}.{{.StructName}}DetailResp, error)
	PageList(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}PageListReq) (*dto{{.PackageName}}.{{.StructName}}PageListResp, error)
}

type {{.StructNameLowerCamel}}Svc struct {
}

var _ {{.StructName}}Svc = (*{{.StructNameLowerCamel}}Svc)(nil)

func New{{.StructName}}Svc() {{.StructName}}Svc {
	return &{{.StructNameLowerCamel}}Svc{}
}

// Create 创建{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Create(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}CreateReq) (*dto{{.PackageName}}.{{.StructName}}CreateResp, error) {
	insertEntity := &{{.ModelLayerName}}.{{.StructName}}Entity{
{{- range .ModelFields}}
	{{- if isSysField .FieldName}}
		{{- continue}}
	{{- end}}
	{{- if eq .FieldType "time.Time"}}
		{{.FieldName}}: time.Unix(req.{{.FieldName}}, 0),
	{{- else}}
		{{.FieldName}}: req.{{.FieldName}},
	{{- end}}
{{- end}}
	}

	if err := {{.DaoPackageName}}.New{{.StructName}}Dao().Insert(ctx, insertEntity); err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Create] {{.DaoPackageName}} Create fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}CreateError)
	}
	return &dto{{.PackageName}}.{{.StructName}}CreateResp{
		ID: insertEntity.ID,
	}, nil
}

// Delete 删除{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Delete(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DeleteReq) error {
	{{.StructNameLowerCamel}}Entity, err := {{.DaoPackageName}}.New{{.StructName}}Dao().GetByID(ctx, req.ID)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Delete] {{.DaoPackageName}} GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.{{.StructName}}DeleteError)
	}
	if {{.StructNameLowerCamel}}Entity == nil || {{.StructNameLowerCamel}}Entity.ID == 0 {
		return code.GetError(code.{{.StructName}}NotExistError)
	}


	userID := gincontext.GetUserID(ctx)

	if err := {{.DaoPackageName}}.New{{.StructName}}Dao().Delete(ctx, req.ID, userID); err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.Delete] {{.DaoPackageName}} Delete fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.{{.StructName}}DeleteError)
	}
	return nil
}

// Update 更新{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Update(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}UpdateReq) error {
	{{.StructNameLowerCamel}}Entity, err := {{.DaoPackageName}}.New{{.StructName}}Dao().GetByID(ctx, req.ID)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Update] {{.DaoPackageName}} GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.{{.StructName}}UpdateError)
	}
	if {{.StructNameLowerCamel}}Entity == nil || {{.StructNameLowerCamel}}Entity.ID == 0 {
		return code.GetError(code.{{.StructName}}NotExistError)
	}

	updateMap := map[string]any{}
	if err := {{.DaoPackageName}}.New{{.StructName}}Dao().UpdateMap(ctx, req.ID, updateMap); err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Update] {{.DaoPackageName}} UpdateMap fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.{{.StructName}}UpdateError)
	}
	return nil
}

// Detail 根据id获取{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Detail(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DetailReq) (*dto{{.PackageName}}.{{.StructName}}DetailResp, error) {
	{{.StructNameLowerCamel}}Entity, err := {{.DaoPackageName}}.New{{.StructName}}Dao().GetByID(ctx, req.ID)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Detail] {{.DaoPackageName}} GetByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}GetDetailError)
	}
	if {{.StructNameLowerCamel}}Entity == nil || {{.StructNameLowerCamel}}Entity.ID == 0 {
		return nil, code.GetError(code.{{.StructName}}NotExistError)
	}
	resp := &dto{{.PackageName}}.{{.StructName}}DetailResp{
		ID:   {{.StructNameLowerCamel}}Entity.ID,
		{{.StructName}}BaseInfo: obj{{.PackageName}}.{{.StructName}}BaseInfo{
	{{- range .ModelFields}}
		{{- if isSysField .FieldName}}
			{{- continue}}
		{{- end}}
		{{- if eq .FieldType "time.Time"}}
			{{.FieldName}}: {{.StructNameLowerCamel}}Entity.{{.FieldName}}.Unix(),
		{{- else}}
			{{.FieldName}}: {{.StructNameLowerCamel}}Entity.{{.FieldName}},
		{{- end}}
	{{- end}}
		},
		OperatorBaseInfo: gobject.OperatorBaseInfo{
			CreatedAt: {{.StructNameLowerCamel}}Entity.CreatedAt.Unix(),
			UpdatedAt: {{.StructNameLowerCamel}}Entity.UpdatedAt.Unix(),
		},
	}
	return resp, nil
}

// PageList 分页获取{{.Description}}列表
func (svc *{{.StructNameLowerCamel}}Svc) PageList(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}PageListReq) (*dto{{.PackageName}}.{{.StructName}}PageListResp, error) {
	cond := &{{.DaoPackageName}}.{{.StructName}}Cond{
		BaseCond: &genericdao.BaseCond{
			Page:     req.Page,
			PageSize: req.PageSize,
		},
	}
	{{.StructNameLowerCamel}}EntityList, total, err := {{.DaoPackageName}}.New{{.StructName}}Dao().GetPageListByCond(ctx, cond)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}PageList] {{.DaoPackageName}} GetPageListByCond fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}GetPageListError)
	}
	list := make([]dto{{.PackageName}}.{{.StructName}}PageListItem, 0, len({{.StructNameLowerCamel}}EntityList))
	for _, v := range {{.StructNameLowerCamel}}EntityList {
		list = append(list, dto{{.PackageName}}.{{.StructName}}PageListItem{
			ID:   v.ID,
			{{.StructName}}BaseInfo: obj{{.PackageName}}.{{.StructName}}BaseInfo{
		{{- range .ModelFields}}
			{{- if isSysField .FieldName}}
				{{- continue}}
			{{- end}}
			{{- if eq .FieldType "time.Time"}}
				{{.FieldName}}: v.{{.FieldName}}.Unix(),
			{{- else}}
				{{.FieldName}}: v.{{.FieldName}},
			{{- end}}
		{{- end}}
			},
			OperatorBaseInfo: gobject.OperatorBaseInfo{
				UpdatedAt: v.UpdatedAt.Unix(),
			},
		})
	}
	return &dto{{.PackageName}}.{{.StructName}}PageListResp{
		List:  list,
		Total: total,
	}, nil
}


