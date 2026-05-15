package models

import "time"

// SysUser 是 B 端员工表，支撑管家后台登录、权限与员工矩阵管理。
type SysUser struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	Username     string    `json:"username" gorm:"size:64;uniqueIndex;not null;comment:'B端员工登录账号'"`
	PasswordHash string    `json:"-" gorm:"size:255;not null;comment:'密码哈希值，禁止明文存储'"`
	RealName     string    `json:"real_name" gorm:"size:64;comment:'员工真实姓名'"`
	Role         string    `json:"role" gorm:"size:32;default:'manager';comment:'员工角色：admin管理员，manager经理'"`
	Status       int       `json:"status" gorm:"default:1;comment:'账号状态：1启用，0禁用'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定 B 端员工表名。
// 1.意图 -> 保持后台账号模型独立于 C 端客户模型。
// 2.步骤 -> 返回 sys_users，方便 SQL 脚本、GORM 与后台接口统一。
// 3.返回 -> 数据库真实表名。
func (SysUser) TableName() string { return "sys_users" }
