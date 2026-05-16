package models

import "time"

// SysDictType 是系统字典类型表，用于统一管理服务标签、资讯分类和业务枚举。
type SysDictType struct {
	ID        uint      `gorm:"primaryKey;comment:'主键ID'" json:"id"`
	DictName  string    `gorm:"size:80;not null;comment:'字典名称'" json:"dict_name"`
	DictCode  string    `gorm:"size:80;uniqueIndex;not null;comment:'字典编码'" json:"dict_code"`
	Remark    string    `gorm:"size:255;comment:'备注说明'" json:"remark"`
	Status    int       `gorm:"comment:'状态：1启用 0禁用'" json:"status"`
	CreatedAt time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

// TableName 指定系统字典类型表名。
// 1.意图 -> 保证 GORM 使用业务指定的 sys_dict_types 表。
// 2.步骤 -> 返回固定表名字符串，避免自动复数化差异。
// 3.返回 -> sys_dict_types。
func (SysDictType) TableName() string { return "sys_dict_types" }
