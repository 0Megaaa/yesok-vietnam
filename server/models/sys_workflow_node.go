package models

import "time"

type SysWorkflowNode struct {
	ID              uint       `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceID       uint       `json:"service_id" gorm:"not null;index;comment:'关联服务ID'"`
	StageCode       string     `json:"stage_code" gorm:"size:64;not null;index;comment:'当前节点编码'"`
	StageName       string     `json:"stage_name" gorm:"size:64;not null;comment:'当前节点名称'"`
	MacroStatus     string     `json:"macro_status" gorm:"size:32;not null;comment:'映射主状态'"`
	ActionName      string     `json:"action_name" gorm:"size:64;not null;comment:'B端操作按钮名称'"`
	NextStageCode   string     `json:"next_stage_code" gorm:"size:64;not null;comment:'流转到的目标节点编码'"`
	IsManual        bool       `json:"is_manual" gorm:"default:true;comment:'是否需要人工触发'"`
	RequireMaterial bool       `json:"require_material" gorm:"default:false;comment:'流转到下一步是否必传资料'"`
	NotifyType      string     `json:"notify_type" gorm:"size:32;comment:'触发TG通知类型'"`
	RoleLimit       string     `json:"role_limit" gorm:"size:32;default:admin;comment:'操作权限限制'"`
	SortOrder       int64      `json:"sort_order" gorm:"default:0;comment:'按钮排序'"`
	CreatedAt       *time.Time `json:"created_at" gorm:"datetime(3);"`
	UpdatedAt       *time.Time `json:"updated_at" gorm:"datetime(3);"`
}

func (SysWorkflowNode) TableName() string { return "sys_workflow_nodes" }
