package model

import (
	"gorm.io/gorm"
)

// UserEntity 用户管理表结构体
type UserEntity struct {
	gorm.Model
	CompanyID    uint   `gorm:"column:company_id;type:bigint unsigned;not null;default 0;comment:公司id"`
	DepartmentID uint   `gorm:"column:department_id;type:bigint unsigned;not null;default 0;comment:部门id"`
	Name         string `gorm:"column:name;type:varchar(255);not null;default '';comment:姓名"`
	CreatedBy    uint   `gorm:"column:created_by;type:bigint unsigned;not null;default 0;comment:创建人id"`
	UpdatedBy    uint   `gorm:"column:updated_by;type:bigint unsigned;not null;default 0;comment:更新人id"`
	DeletedBy    uint   `gorm:"column:deleted_by;type:bigint unsigned;not null;default 0;comment:删除人id"`
}

type UserEntityList []UserEntity

func (UserEntity) TableName() string {
	return TableNameUser
}

func (l UserEntityList) ToMap() map[uint]UserEntity {
	m := make(map[uint]UserEntity)
	for _, v := range l {
		m[v.ID] = v
	}
	return m
}
