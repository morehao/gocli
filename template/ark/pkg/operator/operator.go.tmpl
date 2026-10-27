// Package operator 定义请求级操作者身份：controller 负责填充，service 只读。
//
// 本包不依赖 gin，也不承担数据隔离：身份由 controller 从登录态取出后写入请求 DTO。
package operator

// Operator 当前请求的操作者身份。
//
// 字段带 json:"-" / form:"-"，客户端无法通过请求体或查询参数伪造身份；
// 归属信息（TenantID / DeptID 等）进 DTO 只用于业务判断与落库，
// **不承担数据隔离**——隔离由 DAO 侧的租户作用域强制。
type Operator struct {
	UserID   uint   `json:"-" form:"-"` // 操作人用户ID（写 created_by / updated_by / deleted_by）
	PersonID uint   `json:"-" form:"-"` // 自然人ID
	TenantID uint   `json:"-" form:"-"` // 租户ID
	OrgID    uint   `json:"-" form:"-"` // 组织ID
	DeptID   uint   `json:"-" form:"-"` // 部门ID
	UserType string `json:"-" form:"-"` // 用户类型
}
