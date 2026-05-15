package models

import "time"

// SysConfig 是系统动态配置表，用于向 C 端输出 Banner、热线、主题色等公共配置。
type SysConfig struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ConfigKey   string    `json:"config_key" gorm:"size:128;uniqueIndex;not null;comment:'配置键名'"`
	ConfigValue string    `json:"config_value" gorm:"type:text;comment:'配置值，可为文本或JSON字符串'"`
	ValueType   string    `json:"value_type" gorm:"size:32;default:'string';comment:'配置类型：string字符串，number数字，json对象'"`
	GroupName   string    `json:"group_name" gorm:"size:64;index;comment:'配置分组'"`
	Remark      string    `json:"remark" gorm:"size:500;comment:'配置说明'"`
	IsPublic    bool      `json:"is_public" gorm:"default:true;comment:'是否允许C端读取'"`
	CreatedAt   time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定系统配置表名。
// 1.意图 -> 将前端全局文案、主题和资源配置动态化。
// 2.步骤 -> 返回 sys_configs，C 端通过公开接口读取 is_public=true 的配置。
// 3.返回 -> 数据库真实表名。
func (SysConfig) TableName() string { return "sys_configs" }
