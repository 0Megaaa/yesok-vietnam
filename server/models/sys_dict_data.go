package models

import "time"

// SysDictData 是系统字典数据表，用于承载某一字典编码下的可配置枚举项。
type SysDictData struct {
	ID        uint      `gorm:"primaryKey;comment:'主键ID'" json:"id"`
	DictCode  string    `gorm:"size:80;index;not null;comment:'字典编码，关联 sys_dict_types.dict_code'" json:"dict_code"`
	DictLabel string    `gorm:"size:120;not null;comment:'字典标签'" json:"dict_label"`
	DictValue string    `gorm:"size:120;not null;comment:'字典值'" json:"dict_value"`
	SortOrder int       `gorm:"default:0;comment:'排序值'" json:"sort_order"`
	Status    int       `gorm:"comment:'状态：1启用 0禁用'" json:"status"`
	Remark    string    `gorm:"size:255;comment:'备注说明'" json:"remark"`
	MetaJSON  JSONText  `gorm:"column:meta_json;type:json;comment:'字典扩展元数据'" json:"meta_json"`
	CreatedAt time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

// TableName 指定系统字典数据表名。
// 1.意图 -> 保证 GORM 使用业务指定的 sys_dict_data 表。
// 2.步骤 -> 返回固定表名字符串，避免自动复数化差异。
// 3.返回 -> sys_dict_data。
func (SysDictData) TableName() string { return "sys_dict_data" }
