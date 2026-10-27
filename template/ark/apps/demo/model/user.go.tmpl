package model

import (
	"gorm.io/gorm"
)

// UserEntity 用户管理表结构体
//
// 每个字段显式写 PG 的 type，与 backend/scripts/sql/table_init.sql 一一对应（结构体自描述，
// 读模型即知列类型）。type 必须是 PostgreSQL 合法类型：gorm 只把 bool/int/uint/float/string/
// time/bytes 这类逻辑类型名交给方言映射，其它取值原样拼进 DDL，写成其它方言的字面量
// （如 type:bigint unsigned）在 PG 建表会直接 syntax error at or near "unsigned"。
type UserEntity struct {
	gorm.Model
	CompanyID    uint   `gorm:"column:company_id;type:bigint;not null;default:0;comment:公司id"`
	DepartmentID uint   `gorm:"column:department_id;type:bigint;not null;default:0;comment:部门id"`
	Name         string `gorm:"column:name;type:varchar(255);not null;default:'';comment:姓名"`
	CreatedBy    uint   `gorm:"column:created_by;type:bigint;not null;default:0;comment:创建人id"`
	UpdatedBy    uint   `gorm:"column:updated_by;type:bigint;not null;default:0;comment:更新人id"`
	DeletedBy    uint   `gorm:"column:deleted_by;type:bigint;not null;default:0;comment:删除人id"`
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
