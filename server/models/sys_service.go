package models

import "time"

type SysService struct {
	ID          uint       `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceCode string     `json:"service_code" gorm:"size:64;not null;uniqueIndex;comment:'服务编码'"`
	ServiceName string     `json:"service_name" gorm:"size:128;not null;comment:'服务名称'"`
	DisplayName string     `json:"display_name" gorm:"size:128;comment:'C端展示名称'"`
	Icon        string     `json:"icon" gorm:"size:512;comment:'服务图标'"`
	CoverImage  string     `json:"cover_image" gorm:"size:512;comment:'服务封面图地址'"`
	Description string     `json:"description" gorm:"size:1000;comment:'服务简介'"`
	BasePrice   int64      `json:"base_price" gorm:"default:0;comment:'基础价格，单位为分'"`
	Currency    string     `json:"currency" gorm:"size:16;default:VND;comment:'币种代码'"`
	Unit        string     `json:"unit" gorm:"size:32;comment:'计价单位'"`
	SortOrder   int64      `json:"sort_order" gorm:"default:0;comment:'排序值'"`
	Status      int        `json:"status" gorm:"default:1;index;comment:'服务状态：1启用，0停用'"`
	IsHot       bool       `json:"is_hot" gorm:"default:false;comment:'是否热门服务'"`
	FormSchema  []byte     `json:"form_schema" gorm:"type:json;comment:'C端动态表单结构'"`
	CreatedAt   *time.Time `json:"created_at" gorm:"datetime(3);comment:'创建时间'"`
	UpdatedAt   *time.Time `json:"updated_at" gorm:"datetime(3);comment:'更新时间'"`
}

func (SysService) TableName() string { return "sys_services" }
