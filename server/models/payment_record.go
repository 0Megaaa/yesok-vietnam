package models

import "time"

// PaymentRecord 是支付财务对账表，用于隔离订单履约和资金流水。
type PaymentRecord struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	OrderID      uint      `json:"order_id" gorm:"index;not null;comment:'关联订单ID'"`
	AppUserID    uint      `json:"app_user_id" gorm:"index;not null;comment:'付款客户ID'"`
	PayerName    string    `json:"payer_name" gorm:"size:128;comment:'付款人姓名'"`
	PayAmount    int64     `json:"pay_amount" gorm:"default:0;comment:'支付金额，单位为分'"`
	PayMethod    string    `json:"pay_method" gorm:"size:32;not null;comment:'支付方式：wechat微信，cash现金，bank_transfer银行转账，balance余额'"`
	Status       string    `json:"status" gorm:"size:32;default:'pending';comment:'支付状态：pending待确认，success成功，failed失败，refunded已退款'"`
	ThirdTradeNo string    `json:"third_trade_no" gorm:"size:128;index;comment:'第三方支付流水号'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定支付记录表名。
// 1.意图 -> 将资金流水从订单主表拆出，便于财务对账。
// 2.步骤 -> 通过 order_id 与 third_trade_no 关联订单和第三方平台。
// 3.返回 -> 数据库真实表名。
func (PaymentRecord) TableName() string { return "payment_records" }
