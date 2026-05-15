package models

import "time"

// SysWorkflowNode 是流程大脑表，用状态节点驱动后台订单操作按钮与财务触发。
type SysWorkflowNode struct {
	ID               uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceID        uint      `json:"service_id" gorm:"index;not null;comment:'关联服务ID'"`
	CurrentStatus    string    `json:"current_status" gorm:"size:64;index;not null;comment:'当前状态码'"`
	ButtonName       string    `json:"button_name" gorm:"size:64;not null;comment:'B端操作按钮名称'"`
	TargetStatus     string    `json:"target_status" gorm:"size:64;not null;comment:'点击按钮后的目标状态码'"`
	RequiredMaterial bool      `json:"required_material" gorm:"default:false;comment:'是否必传资料'"`
	TriggerPayment   bool      `json:"trigger_payment" gorm:"default:false;comment:'是否触发财务流水'"`
	SortOrder        int       `json:"sort_order" gorm:"default:0;comment:'按钮排序值'"`
	Remark           string    `json:"remark" gorm:"size:500;comment:'节点说明'"`
	CreatedAt        time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定动态流程节点表名。
// 1.意图 -> 让不同服务拥有独立状态机配置。
// 2.步骤 -> 通过 service_id 与 current_status 定位当前可执行动作。
// 3.返回 -> 数据库真实表名。
func (SysWorkflowNode) TableName() string { return "sys_workflow_nodes" }
