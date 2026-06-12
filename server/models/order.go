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

// ActionType 工作流动作类型常量。
const (
	ActionTypeButtonClick = "button_click"
	ActionTypeFormInput   = "form_input"
	ActionTypeWxPay       = "wx_pay"
)

type Order struct {
	ID                uint           `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	CreatedAt         *time.Time     `json:"created_at" gorm:"column:created_at;datetime(3);comment:'创建时间'"`
	UpdatedAt         *time.Time     `json:"updated_at" gorm:"column:updated_at;datetime(3);comment:'更新时间'"`
	DeletedAt         gorm.DeletedAt `json:"deleted_at" gorm:"column:deleted_at;datetime(3);index;comment:'软删除时间'"`
	UserID            uint           `json:"user_id" gorm:"column:user_id;not null;index;comment:'B端用户ID'"`
	OrderNo           string         `json:"order_no" gorm:"column:order_no;size:64;not null;uniqueIndex;comment:'订单号'"`
	MacroStatus       string         `json:"macro_status" gorm:"column:macro_status;size:50;default:pending;comment:'当前主状态'"`
	CurrentStage      string         `json:"current_stage" gorm:"column:current_stage;size:100;default:start;comment:'当前微观节点'"`
	FormSnapshot      []byte         `json:"form_snapshot" gorm:"column:form_snapshot;type:json;comment:'下单时动态表单数据快照'"`
	Amount            float64        `json:"amount" gorm:"column:amount;type:decimal(16,2);not null;comment:'订单金额'"`
	Currency          string         `json:"currency" gorm:"column:currency;size:16;default:VND;comment:'币种代码'"`
	Note              string         `json:"note" gorm:"column:note;size:1024;comment:'内部备注'"`
	AppUserID         uint           `json:"app_user_id" gorm:"column:app_user_id;not null;index;comment:'C端客户ID'"`
	ServiceID         uint           `json:"service_id" gorm:"column:service_id;not null;index;comment:'服务品类ID'"`
	ServiceName       string         `json:"service_name" gorm:"column:service_name;size:128;comment:'下单时的服务名称快照'"`
	ContactName       string         `json:"contact_name" gorm:"column:contact_name;size:64;comment:'联系人姓名'"`
	ContactPhone      string         `json:"contact_phone" gorm:"column:contact_phone;size:32;comment:'联系人手机号'"`
	TotalAmount       int64          `json:"total_amount" gorm:"column:total_amount;not null;default:0;comment:'订单总金额(分)'"`
	PaymentStatus     string         `json:"payment_status" gorm:"column:payment_status;size:32;default:unpaid;comment:'支付状态'"`
	FormData          []byte         `json:"form_data" gorm:"column:form_data;type:json;comment:'动态表单数据'"`
	Remark            string         `json:"remark" gorm:"column:remark;size:1000;comment:'后台备注'"`
	ButlerID          uint           `json:"butler_id" gorm:"column:butler_id;comment:'分配管家ID'"`
	ButlerName        string         `json:"butler_name" gorm:"column:butler_name;size:64;comment:'分配管家姓名'"`
	ButlerWecomUserID string         `json:"butler_wecom_userid" gorm:"column:butler_wecom_user_id;size:128;comment:'分配管家企业微信UserID'"`
	ButlerContactURL  string         `json:"butler_contact_url" gorm:"column:butler_contact_url;size:1024;comment:'订单专属管家微信客服/联系我链接'"`
	ButlerAssignedAt  *time.Time     `json:"butler_assigned_at" gorm:"column:butler_assigned_at;datetime(3);comment:'管家分配时间'"`
	ButlerContactedAt *time.Time     `json:"butler_contacted_at" gorm:"column:butler_contacted_at;datetime(3);comment:'客户最近一次点击联系管家时间'"`
}

func (Order) TableName() string { return "orders" }
