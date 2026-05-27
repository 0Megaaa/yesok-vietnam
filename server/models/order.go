package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusQuoted    OrderStatus = "quoted"
	OrderStatusReviewing OrderStatus = "reviewing"
	OrderStatusPaid      OrderStatus = "paid"
	OrderStatusProgress  OrderStatus = "in_progress"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusRefunded  OrderStatus = "refunded"
	OrderStatusConfirmed OrderStatus = "confirmed"
)

const DefaultCurrencyCode = "VND"

type Order struct {
	ID            uint           `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	CreatedAt     *time.Time     `json:"created_at" gorm:"datetime(3);comment:'创建时间'"`
	UpdatedAt     *time.Time     `json:"updated_at" gorm:"datetime(3);comment:'更新时间'"`
	DeletedAt     gorm.DeletedAt `json:"deleted_at" gorm:"datetime(3);index;comment:'软删除时间'"`
	UserID        uint           `json:"user_id" gorm:"not null;index;comment:'B端用户ID'"`
	OrderNo       string         `json:"order_no" gorm:"size:64;not null;uniqueIndex;comment:'订单号'"`
	MacroStatus   string         `json:"macro_status" gorm:"size:50;default:pending;comment:'当前主状态'"`
	CurrentStage  string         `json:"current_stage" gorm:"size:100;default:start;comment:'当前微观节点'"`
	FormSnapshot  []byte         `json:"form_snapshot" gorm:"type:json;comment:'下单时动态表单数据快照'"`
	Amount        float64        `json:"amount" gorm:"type:decimal(16,2);not null;comment:'订单金额'"`
	Currency      string         `json:"currency" gorm:"size:16;default:VND;comment:'币种代码'"`
	TGChatID      int64          `json:"tg_chat_id" gorm:"index;comment:'TG会话ID'"`
	TGMessageID   int64          `json:"tg_message_id" gorm:"comment:'TG消息ID'"`
	WorkerTGID    int64          `json:"worker_tg_id" gorm:"index;comment:'接单员工TG ID'"`
	Metadata      string         `json:"metadata" gorm:"type:text;comment:'元数据'"`
	Note          string         `json:"note" gorm:"size:1024;comment:'内部备注'"`
	AppUserID     uint           `json:"app_user_id" gorm:"not null;index;comment:'C端客户ID'"`
	ServiceID     uint           `json:"service_id" gorm:"not null;index;comment:'服务品类ID'"`
	ServiceName   string         `json:"service_name" gorm:"size:128;comment:'下单时的服务名称快照'"`
	ContactName   string         `json:"contact_name" gorm:"size:64;comment:'联系人姓名'"`
	ContactPhone  string         `json:"contact_phone" gorm:"size:32;comment:'联系人手机号'"`
	TotalAmount   int64          `json:"total_amount" gorm:"not null;default:0;comment:'订单总金额(分)'"`
	PaymentStatus string         `json:"payment_status" gorm:"size:32;default:unpaid;comment:'支付状态'"`
	FormData      []byte         `json:"form_data" gorm:"type:json;comment:'动态表单数据'"`
	Remark        string         `json:"remark" gorm:"size:1000;comment:'后台备注'"`
}

func (Order) TableName() string { return "orders" }
