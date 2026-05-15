package models

import "time"

type OrderStatus string

const (
	// OrderStatusPending 表示订单已提交但尚未受理。
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusQuoted 表示管家已报价，等待客户支付或确认。
	OrderStatusQuoted OrderStatus = "quoted"
	// OrderStatusReviewing 表示资料正在审核。
	OrderStatusReviewing OrderStatus = "reviewing"
	// OrderStatusPaid 表示订单已完成收款确认。
	OrderStatusPaid OrderStatus = "paid"
	// OrderStatusProgress 表示订单正在履约处理中。
	OrderStatusProgress OrderStatus = "in_progress"
	// OrderStatusCompleted 表示订单已经履约完成。
	OrderStatusCompleted OrderStatus = "completed"
	// OrderStatusCancelled 表示订单已由客户或后台取消。
	OrderStatusCancelled OrderStatus = "cancelled"
	// OrderStatusFailed 表示履约失败，需要人工介入或退款。
	OrderStatusFailed OrderStatus = "failed"
	// OrderStatusRefunded 表示订单已完成退款闭环。
	OrderStatusRefunded OrderStatus = "refunded"
	// OrderStatusConfirmed 表示旧版兼容状态，等价于已确认。
	OrderStatusConfirmed OrderStatus = "confirmed"
)

const (
	// DefaultCurrencyCode 是系统默认货币代码，避免金额展示中出现魔术字符串。
	DefaultCurrencyCode = "VND"
)

// Order 是主订单表，围绕“客户 + 服务 + 当前工作流状态 + 动态表单”组织核心业务数据。
type Order struct {
	ID            uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderNo       string    `json:"order_no" gorm:"size:64;uniqueIndex;not null;comment:'订单号'"`
	AppUserID     uint      `json:"app_user_id" gorm:"index;not null;comment:'C端客户ID'"`
	ServiceID     uint      `json:"service_id" gorm:"index;not null;comment:'服务品类ID'"`
	ServiceName   string    `json:"service_name" gorm:"size:128;comment:'下单时的服务名称快照'"`
	ContactName   string    `json:"contact_name" gorm:"size:64;comment:'联系人姓名'"`
	ContactPhone  string    `json:"contact_phone" gorm:"size:32;comment:'联系人手机号'"`
	TotalAmount   int64     `json:"total_amount" gorm:"default:0;comment:'订单总金额，单位为分'"`
	Currency      string    `json:"currency" gorm:"size:16;default:'VND';comment:'币种代码'"`
	CurrentStatus string    `json:"current_status" gorm:"size:64;index;not null;comment:'当前状态码'"`
	PaymentStatus string    `json:"payment_status" gorm:"size:32;default:'unpaid';comment:'支付状态：unpaid未付，pending待确认，paid已付，refunded已退'"`
	FormData      string    `json:"form_data" gorm:"type:json;comment:'动态表单数据，存储各类业务自定义详情'"`
	Remark        string    `json:"remark" gorm:"size:1000;comment:'后台备注'"`
	CreatedAt     time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"comment:'更新时间'"`

	// 旧版兼容字段仅用于历史 handlers 编译过渡，不迁移到新版 orders 表。
	UserID      uint        `json:"user_id,omitempty" gorm:"-"`
	User        User        `json:"-" gorm:"-"`
	Status      OrderStatus `json:"status,omitempty" gorm:"-"`
	Amount      float64     `json:"amount,omitempty" gorm:"-"`
	TGChatID    int64       `json:"tg_chat_id,omitempty" gorm:"-"`
	TGMessageID int64       `json:"tg_message_id,omitempty" gorm:"-"`
	WorkerTGID  int64       `json:"worker_tg_id,omitempty" gorm:"-"`
	Metadata    string      `json:"metadata,omitempty" gorm:"-"`
	Note        string      `json:"note,omitempty" gorm:"-"`
}

// TableName 固定主订单表名。
// 1.意图 -> 让 GORM AutoMigrate 与手写 SQL 都指向 orders 表。
// 2.步骤 -> 返回 orders，并优先使用 form_data 承接垂直业务字段。
// 3.返回 -> 数据库真实表名。
func (Order) TableName() string { return "orders" }
