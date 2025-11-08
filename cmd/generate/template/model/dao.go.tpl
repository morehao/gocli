package dao{{.PackageName}}

import (
	"fmt"
	"time"
	
    "{{.ModulePath}}/pkg/code"

    {{- if isDefaultDaoLayer .DaoLayerName}}
    "{{.ModulePath}}/{{.AppPathInProject}}/dao"
    {{- else}}
    "{{.ModulePath}}/{{.AppPathInProject}}/dao/{{.DaoLayerName}}"
    {{- end}}
    {{- if isDefaultModelLayer .ModelLayerName}}
    "{{.ModulePath}}/{{.AppPathInProject}}/model"
    {{- else}}
    "{{.ModulePath}}/{{.AppPathInProject}}/model/{{.ModelLayerName}}"
    {{- end}}

    "github.com/gin-gonic/gin"
    "github.com/morehao/golib/gutils"
    "gorm.io/gorm"
)

type {{.StructName}}Cond struct {
	ID             uint
	IDs            []uint
	IsDelete       bool
	Page           int
	PageSize       int
	CreatedAtStart int64
	CreatedAtEnd   int64
	OrderField     string
}

type {{.StructName}}Dao struct {
	{{.DaoLayerName}}.Base
}

func New{{.StructName}}Dao() *{{.StructName}}Dao {
	return &{{.StructName}}Dao{}
}

func (d *{{.StructName}}Dao) TableName() string {
	return {{.ModelLayerName}}.TableName{{.StructName}}
}

func (d *{{.StructName}}Dao) WithTx(db *gorm.DB) *{{.StructName}}Dao {
	return &{{.StructName}}Dao{
		Base: {{.DaoLayerName}}.Base{Tx: db},
	}
}

func (d *{{.StructName}}Dao) Insert(ctx *gin.Context, entity *{{.ModelLayerName}}.{{.StructName}}Entity) error {
	db := d.DB(ctx).Table(d.TableName())
	if err := db.Create(entity).Error; err != nil {
		return code.GetError(code.DBInsertErr).Wrapf(err, "[{{.StructName}}Dao] Insert fail, entity:%s", gutils.ToJsonString(entity))
	}
	return nil
}

func (d *{{.StructName}}Dao) BatchInsert(ctx *gin.Context, entityList {{.ModelLayerName}}.{{.StructName}}EntityList) error {
	if len(entityList) == 0 {
		return code.GetError(code.DBInsertErr).Wrapf(nil, "[{{.StructName}}Dao] BatchInsert fail, entityList is empty")
	}

	db := d.DB(ctx).Table(d.TableName())
	if err := db.Create(entityList).Error; err != nil {
		return code.GetError(code.DBInsertErr).Wrapf(err, "[{{.StructName}}Dao] BatchInsert fail, entityList:%s", gutils.ToJsonString(entityList))
	}
	return nil
}

func (d *{{.StructName}}Dao) UpdateByID(ctx *gin.Context, id uint, entity *{{.ModelLayerName}}.{{.StructName}}Entity) error {
	db := d.DB(ctx).Table(d.TableName())
	if err := db.Where("id = ?", id).Updates(entity).Error; err != nil {
		return code.GetError(code.DBUpdateErr).Wrapf(err, "[{{.StructName}}Dao] UpdateByID fail, id:%d entity:%s", id, gutils.ToJsonString(entity))
	}
	return nil
}

func (d *{{.StructName}}Dao) UpdateMap(ctx *gin.Context, id uint, updateMap map[string]interface{}) error {
	db := d.DB(ctx).Table(d.TableName())
	if err := db.Where("id = ?", id).Updates(updateMap).Error; err != nil {
		return code.GetError(code.DBUpdateErr).Wrapf(err, "[{{.StructName}}Dao] UpdateMap fail, id:%d, updateMap:%s", id, gutils.ToJsonString(updateMap))
	}
	return nil
}

func (d *{{.StructName}}Dao) Delete(ctx *gin.Context, id, deletedBy uint) error {
	db := d.DB(ctx).Table(d.TableName())
	updatedField := map[string]interface{}{
		"deleted_time": time.Now(),
		"deleted_by":   deletedBy,
	}
	if err := db.Where("id = ?", id).Updates(updatedField).Error; err != nil {
		return code.GetError(code.DBUpdateErr).Wrapf(err, "[{{.StructName}}Dao] Delete fail, id:%d, deletedBy:%d", id, deletedBy)
	}
	return nil
}

func (d *{{.StructName}}Dao) GetById(ctx *gin.Context, id uint) (*{{.ModelLayerName}}.{{.StructName}}Entity, error) {
	var entity {{.ModelLayerName}}.{{.StructName}}Entity
	db := d.DB(ctx).Table(d.TableName())
	if err := db.Where("id = ?", id).Find(&entity).Error; err != nil {
		return nil, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] GetById fail, id:%d", id)
	}
	return &entity, nil
}

func (d *{{.StructName}}Dao) GetByCond(ctx *gin.Context, cond *{{.StructName}}Cond) (*{{.ModelLayerName}}.{{.StructName}}Entity, error) {
	var entity {{.ModelLayerName}}.{{.StructName}}Entity
	db := d.DB(ctx).Table(d.TableName())

	d.BuildCondition(db, cond)

	if err := db.Find(&entity).Error; err != nil {
		return nil, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] GetById fail, cond:%s", gutils.ToJsonString(cond))
	}
	return &entity, nil
}

func (d *{{.StructName}}Dao) GetListByCond(ctx *gin.Context, cond *{{.StructName}}Cond) ({{.ModelLayerName}}.{{.StructName}}EntityList, error) {
	var entityList {{.ModelLayerName}}.{{.StructName}}EntityList
	db := d.DB(ctx).Table(d.TableName())

	d.BuildCondition(db, cond)

	if err := db.Find(&entityList).Error; err != nil {
		return nil, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] GetListByCond fail, cond:%s", gutils.ToJsonString(cond))
	}
	return entityList, nil
}

func (d *{{.StructName}}Dao) GetPageListByCond(ctx *gin.Context, cond *{{.StructName}}Cond) ({{.ModelLayerName}}.{{.StructName}}EntityList, int64, error) {
	db := d.DB(ctx).Table(d.TableName())

	d.BuildCondition(db, cond)

	var count int64
	if err := db.Count(&count).Error; err != nil {
		return nil, 0, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] GetPageListByCond count fail, cond:%s", gutils.ToJsonString(cond))
	}
	if cond.PageSize > 0 && cond.Page > 0 {
		db.Offset((cond.Page - 1) * cond.PageSize).Limit(cond.PageSize)
	}
	var entityList {{.ModelLayerName}}.{{.StructName}}EntityList
	if err := db.Find(&entityList).Error; err != nil {
		return nil, 0, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] GetPageListByCond find fail, cond:%s", gutils.ToJsonString(cond))
	}
	return entityList, count, nil
}

func (d *{{.StructName}}Dao) CountByCond(ctx *gin.Context, cond *{{.StructName}}Cond) (int64, error) {
	db := d.DB(ctx).Table(d.TableName())

	d.BuildCondition(db, cond)
	var count int64
	if err := db.Count(&count).Error; err != nil {
		return 0, code.GetError(code.DBFindErr).Wrapf(err, "[{{.StructName}}Dao] CountByCond fail, cond:%s", gutils.ToJsonString(cond))
	}
	return count, nil
}

func (d *{{.StructName}}Dao) BuildCondition(db *gorm.DB, cond *{{.StructName}}Cond) {
	if cond.ID > 0 {
		query := fmt.Sprintf("%s.id = ?", d.TableName())
		db.Where(query, cond.ID)
	}
	if len(cond.IDs) > 0 {
		query := fmt.Sprintf("%s.id in (?)", d.TableName())
		db.Where(query, cond.IDs)
	}
	if cond.CreatedAtStart > 0 {
		query := fmt.Sprintf("%s.created_at >= ?", d.TableName())
		db.Where(query, time.Unix(cond.CreatedAtStart, 0))
	}
	if cond.CreatedAtEnd > 0 {
		query := fmt.Sprintf("%s.created_at <= ?", d.TableName())
		db.Where(query, time.Unix(cond.CreatedAtEnd, 0))
	}
	if cond.IsDelete {
		db.Unscoped()
	}

	if cond.OrderField != "" {
		db.Order(cond.OrderField)
	}
} 