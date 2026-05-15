package models

import "time"

// SysService 是服务配置表，由 B 端维护并实时驱动 C 端首页与下单入口。
type SysService struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceCode string    `json:"service_code" gorm:"size:64;uniqueIndex;not null;comment:'服务编码，用于前端路由与业务识别'"`
	ServiceName string    `json:"service_name" gorm:"size:128;not null;comment:'服务名称'"`
	DisplayName string    `json:"display_name" gorm:"size:128;comment:'C端展示名称'"`
	Icon        string    `json:"icon" gorm:"size:512;comment:'服务图标地址或图标标识'"`
	CoverImage  string    `json:"cover_image" gorm:"size:512;comment:'服务封面图地址'"`
	Description string    `json:"description" gorm:"size:1000;comment:'服务简介'"`
	BasePrice   int64     `json:"base_price" gorm:"default:0;comment:'基础价格，单位为分'"`
	Currency    string    `json:"currency" gorm:"size:16;default:'VND';comment:'币种代码'"`
	Unit        string    `json:"unit" gorm:"size:32;comment:'计价单位，例如次、人、单'"`
	SortOrder   int       `json:"sort_order" gorm:"default:0;comment:'排序值，越小越靠前'"`
	Status      int       `json:"status" gorm:"default:1;index;comment:'服务状态：1启用，0停用'"`
	IsHot       bool      `json:"is_hot" gorm:"default:false;comment:'是否热门服务'"`
	FormSchema  string    `json:"form_schema" gorm:"type:json;comment:'C端动态表单结构 JSON'"`
	CreatedAt   time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定服务配置表名。
// 1.意图 -> 将 C 端服务卡片、价格和表单配置从前端硬编码中解耦。
// 2.步骤 -> 统一返回 sys_services，后台修改后 C 端可通过接口实时读取。
// 3.返回 -> 数据库真实表名。
func (SysService) TableName() string { return "sys_services" }
