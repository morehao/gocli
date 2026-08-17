package dao

import (
	"github.com/morehao/go-ark-template/demo/model"
	"github.com/morehao/go-ark-template/pkg/dbclient"
	"github.com/morehao/golib/dbaccess/gormdao"
	"gorm.io/gorm"
)

type UserCond struct {
	*gormdao.BaseCond
	CompanyID    uint
	DepartmentID uint
	Name         string
	CreatedBy    uint
	UpdatedBy    uint
	DeletedBy    uint
}

func (c *UserCond) BuildCondition(db *gorm.DB, tableName string) {
	if c.BaseCond != nil {
		c.BaseCond.BuildCondition(db, tableName)
	}
	if c.CompanyID != 0 {
		db.Where(tableName+".company_id = ?", c.CompanyID)
	}
	if c.DepartmentID != 0 {
		db.Where(tableName+".department_id = ?", c.DepartmentID)
	}
	if c.Name != "" {
		db.Where(tableName+".name = ?", c.Name)
	}
	if c.CreatedBy != 0 {
		db.Where(tableName+".created_by = ?", c.CreatedBy)
	}
	if c.UpdatedBy != 0 {
		db.Where(tableName+".updated_by = ?", c.UpdatedBy)
	}
	if c.DeletedBy != 0 {
		db.Where(tableName+".deleted_by = ?", c.DeletedBy)
	}
}

type UserDao struct {
	*gormdao.Dao[model.UserEntity, model.UserEntityList, uint]
}

func NewUserDao() *UserDao {
	return &UserDao{
		Dao: gormdao.NewDao[model.UserEntity, model.UserEntityList, uint](
			model.TableNameUser, "UserDao",
			dbclient.DemoDB,
		),
	}
}
