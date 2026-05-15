package models

import "time"

type OrderStatus string

const (
	// OrderStatusPending 表示订单已提交但尚未受理。
	OrderStatusPending OrderStatus = "pending"
	// OrderStatusConfirmed 表示后台已确认并进入资料核验。
	OrderStatusConfirmed OrderStatus = "confirmed"
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
)

const (
	// DefaultCurrencyCode 是系统默认货币代码，避免金额展示中出现魔术字符串。
	DefaultCurrencyCode = "VND"
)

// Order 是新版主订单表，围绕“客户 + 服务 + 当前工作流状态 + 动态表单”组织核心业务数据。
type Order struct {
	ID            uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderNo       string    `json:"order_no" gorm:"size:64;uniqueIndex;not null;comment:'订单号'"`
	AppUserID     uint      `json:"app_user_id" gorm:"index;not null;comment:'C端客户ID'"`
	ServiceID     uint      `json:"service_id" gorm:"index;not null;comment:'服务品类ID'"`
	TotalAmount   int64     `json:"total_amount" gorm:"default:0;comment:'订单总金额，单位为分'"`
	CurrentStatus string    `json:"current_status" gorm:"size:64;index;not null;comment:'当前状态码'"`
	FormData      string    `json:"form_data" gorm:"type:json;comment:'动态表单数据，存储各类业务自定义详情'"`
	CreatedAt     time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt     time.Time `json:"updated_at" gorm:"comment:'更新时间'"`

	// 以下字段仅用于旧版 handlers 编译过渡，不会写入新版 orders 表。
	// 实现步骤：
	// 1. 暂时接住旧接口中的 user_id、status、amount 等字段访问。
	// 2. 通过 gorm:"-" 明确禁止迁移到数据库，保证 7 表结构干净。
	// 3. 待旧后台接口迁移完成后一次性删除该兼容区块。
	UserID      uint        `json:"user_id,omitempty" gorm:"-"`
	User        User        `json:"-" gorm:"-"`
	Status      OrderStatus `json:"status,omitempty" gorm:"-"`
	Amount      float64     `json:"amount,omitempty" gorm:"-"`
	Currency    string      `json:"currency,omitempty" gorm:"-"`
	TGChatID    int64       `json:"tg_chat_id,omitempty" gorm:"-"`
	TGMessageID int64       `json:"tg_message_id,omitempty" gorm:"-"`
	WorkerTGID  int64       `json:"worker_tg_id,omitempty" gorm:"-"`
	Metadata    string      `json:"metadata,omitempty" gorm:"-"`
	Note        string      `json:"note,omitempty" gorm:"-"`
}

// TableName 固定主订单表名。
// 实现步骤：
// 1. 让 GORM AutoMigrate 与手写 SQL 都指向 orders 表。
// 2. 后续新增订单字段必须优先评估是否应进入 form_data 动态 JSON。
func (Order) TableName() string { return "orders" }
