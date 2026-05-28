package models

import "time"

// 审核状态常量
const (
	AuditStatusPending  = "pending"  // 待审核
	AuditStatusApproved = "approved" // 审核通过
	AuditStatusRejected = "rejected" // 审核拒绝
)

type OrderTimeline struct {
	ID           uint       `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderID      uint       `json:"order_id" gorm:"not null;index;comment:'关联订单ID'"`
	BeforeStatus string     `json:"before_status" gorm:"size:64;comment:'变更前状态码'"`
	AfterStatus  string     `json:"after_status" gorm:"size:64;not null;comment:'变更后状态码'"`
	Operator     string     `json:"operator" gorm:"size:128;comment:'操作人，可记录员工ID或系统标识'"`
	Remark       string     `json:"remark" gorm:"size:1000;comment:'备注或对客留言'"`
	CreatedAt    *time.Time `json:"created_at" gorm:"datetime(3);comment:'创建时间'"`
	UpdatedAt    *time.Time `json:"updated_at" gorm:"datetime(3);comment:'更新时间'"`
	ActionName   string     `json:"action_name" gorm:"column:action_code;size:64;comment:'触发该记录的操作动作名称'"`
	Payload      []byte     `json:"payload" gorm:"type:json;comment:'form_input 节点提交的数据 JSON'"`
	AuditStatus  string     `json:"audit_status" gorm:"size:32;default:approved;comment:'审核状态：pending/approved/rejected'"`
}

func (OrderTimeline) TableName() string { return "order_timelines" }
