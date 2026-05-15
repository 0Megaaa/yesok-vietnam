package models

import "time"

// SysService 是服务品类表，统一管理首页服务卡片、下单入口和后台流程配置。
type SysService struct {
	ID          uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceName string    `json:"service_name" gorm:"size:128;not null;comment:'服务名称'"`
	Icon        string    `json:"icon" gorm:"size:512;comment:'服务图标地址或图标标识'"`
	BasePrice   int64     `json:"base_price" gorm:"default:0;comment:'基础价格，单位为分'"`
	CreatedAt   time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt   time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定服务品类表名。
// 实现步骤：
// 1. 将服务配置与订单数据解耦。
// 2. 返回 sys_services，便于后台未来维护服务目录。
func (SysService) TableName() string { return "sys_services" }

// SysWorkflowNode 是动态流程配置表，用按钮和状态流转驱动后台履约动作。
type SysWorkflowNode struct {
	ID               uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	ServiceID        uint      `json:"service_id" gorm:"index;not null;comment:'关联服务ID'"`
	CurrentStatus    string    `json:"current_status" gorm:"size:64;not null;comment:'当前状态码'"`
	ButtonName       string    `json:"button_name" gorm:"size:64;not null;comment:'B端操作按钮名称'"`
	TargetStatus     string    `json:"target_status" gorm:"size:64;not null;comment:'点击按钮后的目标状态码'"`
	RequiredMaterial bool      `json:"required_material" gorm:"default:false;comment:'是否必传资料'"`
	TriggerPayment   bool      `json:"trigger_payment" gorm:"default:false;comment:'是否触发支付'"`
	CreatedAt        time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt        time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定动态流程节点表名。
// 实现步骤：
// 1. 每个服务可以拥有多套状态按钮。
// 2. 通过 service_id + current_status 查找当前可执行动作。
func (SysWorkflowNode) TableName() string { return "sys_workflow_nodes" }

// OrderTimeline 是订单状态轨迹表，用于记录客户可见的履约进度和后台操作日志。
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
// 实现步骤：
// 1. 主订单只保存当前状态。
// 2. 每次状态变更追加一条轨迹，保证履约历史可追溯。
func (OrderTimeline) TableName() string { return "order_timelines" }

// PaymentRecord 是支付财务对账表，用于隔离订单履约和资金流水。
type PaymentRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderID      uint      `json:"order_id" gorm:"index;not null;comment:'关联订单ID'"`
	AppUserID    uint      `json:"app_user_id" gorm:"index;not null;comment:'付款客户ID'"`
	PayAmount    int64     `json:"pay_amount" gorm:"default:0;comment:'支付金额，单位为分'"`
	PayMethod    string    `json:"pay_method" gorm:"size:32;not null;comment:'支付方式：wechat微信，tg_pay Telegram支付，balance余额'"`
	Status       string    `json:"status" gorm:"size:32;default:'pending';comment:'支付状态：pending待支付，success成功，failed失败，refunded已退款'"`
	ThirdTradeNo string    `json:"third_trade_no" gorm:"size:128;index;comment:'第三方支付流水号'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定支付记录表名。
// 实现步骤：
// 1. 将资金流水从订单主表拆出。
// 2. 通过 third_trade_no 与第三方支付平台对账。
func (PaymentRecord) TableName() string { return "payment_records" }
