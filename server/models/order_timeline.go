package models

import "time"

type OrderTimeline struct {
	ID           uint       `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderID      uint       `json:"order_id" gorm:"not null;index;comment:'关联订单ID'"`
	BeforeStatus string     `json:"before_status" gorm:"size:64;comment:'变更前状态码'"`
	AfterStatus  string     `json:"after_status" gorm:"size:64;not null;comment:'变更后状态码'"`
	Operator     string     `json:"operator" gorm:"size:128;comment:'操作人，可记录员工ID或系统标识'"`
	Remark       string     `json:"remark" gorm:"size:1000;comment:'备注或对客留言'"`
	CreatedAt    *time.Time `json:"created_at" gorm:"datetime(3);comment:'创建时间'"`
	UpdatedAt    *time.Time `json:"updated_at" gorm:"datetime(3);comment:'更新时间'"`
	IsNotifyUser bool       `json:"is_notify_user" gorm:"default:false;comment:'是否已推送通知给用户'"`
	ActionCode   string     `json:"action_code" gorm:"size:64;comment:'触发该记录的操作动作代码'"`
}

func (OrderTimeline) TableName() string { return "order_timelines" }
