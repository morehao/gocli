package dao

import (
	"github.com/morehao/example/apps/demoapp/model"
	"github.com/morehao/example/pkg/dbclient"
	"github.com/morehao/golib/biz/genericdao"
	"gorm.io/gorm"
)

type UserCond struct {
	*genericdao.BaseCond
	CompanyID    uint
	CreatedBy    uint
	DeletedBy    uint
	DepartmentID uint
	Name         string
	UpdatedBy    uint
}

func (c *UserCond) BuildCondition(db *gorm.DB, tableName string) {
	if c.BaseCond != nil {
		c.BaseCond.BuildCondition(db, tableName)
	}
	if c.CompanyID != 0 {
		db.Where(tableName+".company_id = ?", c.CompanyID)
	}
	if c.CreatedBy != 0 {
		db.Where(tableName+".created_by = ?", c.CreatedBy)
	}
	if c.DeletedBy != 0 {
		db.Where(tableName+".deleted_by = ?", c.DeletedBy)
	}
	if c.DepartmentID != 0 {
		db.Where(tableName+".department_id = ?", c.DepartmentID)
	}
	if c.Name != "" {
		db.Where(tableName+".name = ?", c.Name)
	}
	if c.UpdatedBy != 0 {
		db.Where(tableName+".updated_by = ?", c.UpdatedBy)
	}
}

type UserDao struct {
	*genericdao.GenericDao[model.UserEntity, model.UserEntityList]
}

func NewUserDao() *UserDao {
	return &UserDao{
		GenericDao: genericdao.NewGenericDao[model.UserEntity, model.UserEntityList](
			model.TableNameUser, "UserDao",
			dbclient.DemoDBGetter,
		),
	}
}
