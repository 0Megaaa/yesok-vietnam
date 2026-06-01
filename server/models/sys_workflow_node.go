package models

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// FormFieldDef 描述动态表单中的单个字段。
// 支持的 type：text|textarea|number|date|datetime|select|image|file|phone
type FormFieldDef struct {
	Key      string `json:"key"`      // 字段标识
	Label    string `json:"label"`    // 前端显示标签
	Type     string `json:"type"`     // text|textarea|number|date|datetime|select|image|file|phone
	Required bool   `json:"required"` // 是否必填
	Options  []struct {
		Label string `json:"label"`
		Value string `json:"value"`
	} `json:"options,omitempty"` // select 类型时的选项
}

// FormFields 是 form_fields 列的 JSON 类型别名，实现 GORM 的 Scanner/Valuer 接口。
type FormFields []FormFieldDef

// Value 实现 driver.Valuer，将 FormFields 序列化为 JSON 字符串存入数据库。
func (f FormFields) Value() (driver.Value, error) {
	if f == nil {
		return nil, nil
	}
	b, err := json.Marshal(f)
	if err != nil {
		return nil, err
	}
	return string(b), nil
}

// Scan 实现 sql.Scanner，从数据库读取 JSON 字符串并反序列化为 FormFields。
func (f *FormFields) Scan(value interface{}) error {
	if value == nil {
		*f = nil
		return nil
	}

	var b []byte
	switch v := value.(type) {
	case []byte:
		b = v
	case string:
		b = []byte(v)
	default:
		return nil
	}

	if len(b) == 0 {
		*f = nil
		return nil
	}

	return json.Unmarshal(b, f)
}

// SysWorkflowNode 工作流动作面板配置，对应 sys_workflow_nodes 表。
// 动作行为类型（action_type）：
//   - button_click : 普通按钮，点击即推进状态
//   - form_input   : 需弹窗收集 inputData 后推进
//   - wx_pay       : 唤起微信支付
type SysWorkflowNode struct {
	ID                uint       `json:"id" gorm:"primaryKey"`
	ServiceID         uint       `json:"service_id" gorm:"not null;index:idx_service_id;comment:'关联服务ID'"`
	StageCode         string     `json:"stage_code" gorm:"size:64;not null;index:idx_stage_code;comment:'当前节点编码'"`
	StageName         string     `json:"stage_name" gorm:"size:64;not null;comment:'当前节点名称'"`
	MacroStatus       string     `json:"macro_status" gorm:"size:32;not null;comment:'映射主状态 (对应sys_dict_data)'"`
	ActionName        string     `json:"action_name" gorm:"size:64;not null;comment:'动作标识 (内部使用，唯一键)'"`
	ButtonLabel       string     `json:"button_label" gorm:"size:64;not null;comment:'按钮名称 (UI显示)'"`
	ExecutorRole      string     `json:"executor_role" gorm:"size:32;not null;default:admin;comment:'admin/client/both'"`
	ActionType        string     `json:"action_type" gorm:"size:32;not null;default:button_click;comment:'button_click/form_input/wx_pay'"`
	FormFields        FormFields `json:"form_fields" gorm:"type:json;comment:'JSON: [{key,label,type,required,options}]'"`
	NeedAudit         bool       `json:"need_audit" gorm:"default:false;comment:'提交后是否需人工审核确认才推进'"`
	NotifyType        string     `json:"notify_type" gorm:"size:32;comment:'触发通知类型，对应 notify_type 字典'"`
	TargetStatus      string     `json:"target_status" gorm:"size:64;not null;comment:'流转目标状态'"`
	AuditRejectStatus string     `json:"audit_reject_status" gorm:"size:64;comment:'审核拒绝后的回退节点编码'"`
	SortOrder         int64      `json:"sort_order" gorm:"default:0;comment:'排序'"`
	CreatedAt         *time.Time `json:"created_at" gorm:"datetime(3);"`
	UpdatedAt         *time.Time `json:"updated_at" gorm:"datetime(3);"`
}

func (SysWorkflowNode) TableName() string { return "sys_workflow_nodes" }
