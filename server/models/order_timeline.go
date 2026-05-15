package models

import "time"

// OrderTimeline 是订单状态轨迹表，用于记录客户可见履约进度和后台操作日志。
type OrderTimeline struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderID      uint      `json:"order_id" gorm:"index;not null;comment:'关联订单ID'"`
	BeforeStatus string    `json:"before_status" gorm:"size:64;comment:'变更前状态码'"`
	AfterStatus  string    `json:"after_status" gorm:"size:64;not null;comment:'变更后状态码'"`
	Operator     string    `json:"operator" gorm:"size:128;comment:'操作人，可记录员工ID或系统标识'"`
	Remark       string    `json:"remark" gorm:"size:1000;comment:'备注或对客留言'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定订单轨迹表名。
// 1.意图 -> 将主订单当前状态与历史流转记录拆分。
// 2.步骤 -> 每次状态变更追加 order_timelines 记录。
// 3.返回 -> 数据库真实表名。
func (OrderTimeline) TableName() string { return "order_timelines" }
