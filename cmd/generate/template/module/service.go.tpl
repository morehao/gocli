package svc{{.PackageName}}

import (
	"github.com/gin-gonic/gin"
    {{- if isDefaultDaoLayer .DaoLayerName}}
    "{{.ModulePath}}/{{.AppPathInProject}}/dao/dao{{.PackageName}}"
    {{- else}}
    "{{.ModulePath}}/{{.AppPathInProject}}/{{.DaoLayerName}}/dao{{.PackageName}}"
    {{- end}}
	"{{.ModulePath}}/{{.AppPathInProject}}/internal/dto/dto{{.PackageName}}"
    {{- if isDefaultModelLayer .ModelLayerName}}
    "{{.ModulePath}}/{{.AppPathInProject}}/model"
    {{- else}}
    "{{.ModulePath}}/{{.AppPathInProject}}/{{.ModelLayerName}}"
    {{- end}}
	"{{.ModulePath}}/{{.AppPathInProject}}/object/obj{{.PackageName}}"
	"{{.ModulePath}}/pkg/code"
	"github.com/morehao/golib/gcontext/gincontext"
	"github.com/morehao/golib/glog"
	"github.com/morehao/golib/gobject"
	"github.com/morehao/golib/gutil"
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

	if err := dao{{.PackageName}}.New{{.StructName}}Dao().Insert(ctx, insertEntity); err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Create] dao{{.StructName}} Create fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}CreateError)
	}
	return &dto{{.PackageName}}.{{.StructName}}CreateResp{
		ID: insertEntity.ID,
	}, nil
}

// Delete 删除{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Delete(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DeleteReq) error {
	userID := gincontext.GetUserID(ctx)

	if err := dao{{.PackageName}}.New{{.StructName}}Dao().Delete(ctx, req.ID, userID); err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.Delete] dao{{.StructName}} Delete fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return code.GetError(code.{{.StructName}}DeleteError)
	}
	return nil
}

// Update 更新{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Update(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}UpdateReq) error {

	updateEntity := &{{.ModelLayerName}}.{{.StructName}}Entity{
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
    if err := dao{{.PackageName}}.New{{.StructName}}Dao().UpdateByID(ctx, req.ID, updateEntity); err != nil {
        glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Update] dao{{.StructName}} UpdateByID fail, err:%v, req:%s", err, gutil.ToJsonString(req))
        return code.GetError(code.{{.StructName}}UpdateError)
    }
    return nil
}

// Detail 根据id获取{{.Description}}
func (svc *{{.StructNameLowerCamel}}Svc) Detail(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}DetailReq) (*dto{{.PackageName}}.{{.StructName}}DetailResp, error) {
	detailEntity, err := dao{{.PackageName}}.New{{.StructName}}Dao().GetById(ctx, req.ID)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}Detail] dao{{.StructName}} GetById fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}GetDetailError)
	}
    // 判断是否存在
    if detailEntity == nil || detailEntity.ID == 0 {
        return nil, code.GetError(code.{{.StructName}}NotExistError)
    }
	resp := &dto{{.PackageName}}.{{.StructName}}DetailResp{
		ID:   detailEntity.ID,
		{{.StructName}}BaseInfo: obj{{.PackageName}}.{{.StructName}}BaseInfo{
	{{- range .ModelFields}}
		{{- if isSysField .FieldName}}
			{{- continue}}
		{{- end}}
		{{- if eq .FieldType "time.Time"}}
			{{.FieldName}}: detailEntity.{{.FieldName}}.Unix(),
		{{- else}}
			{{.FieldName}}: detailEntity.{{.FieldName}},
		{{- end}}
	{{- end}}
		},
		OperatorBaseInfo: gobject.OperatorBaseInfo{
			CreatedAt: detailEntity.CreatedAt.Unix(),
			UpdatedAt: detailEntity.UpdatedAt.Unix(),
		},
	}
	return resp, nil
}

// PageList 分页获取{{.Description}}列表
func (svc *{{.StructNameLowerCamel}}Svc) PageList(ctx *gin.Context, req *dto{{.PackageName}}.{{.StructName}}PageListReq) (*dto{{.PackageName}}.{{.StructName}}PageListResp, error) {
	cond := &dao{{.PackageName}}.{{.StructName}}Cond{
		Page:     req.Page,
		PageSize: req.PageSize,
	}
	dataList, total, err := dao{{.PackageName}}.New{{.StructName}}Dao().GetPageListByCond(ctx, cond)
	if err != nil {
		glog.Errorf(ctx, "[svc{{.PackageName}}.{{.StructName}}PageList] dao{{.StructName}} GetPageListByCond fail, err:%v, req:%s", err, gutil.ToJsonString(req))
		return nil, code.GetError(code.{{.StructName}}GetPageListError)
	}
	list := make([]dto{{.PackageName}}.{{.StructName}}PageListItem, 0, len(dataList))
	for _, v := range dataList {
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


