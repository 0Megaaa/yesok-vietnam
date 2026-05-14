package models

import (
	"time"

	"gorm.io/gorm"
)

type OrderStatus string

const (
	OrderStatusPending   OrderStatus = "pending"
	OrderStatusConfirmed OrderStatus = "confirmed"
	OrderStatusProgress  OrderStatus = "in_progress"
	OrderStatusCompleted OrderStatus = "completed"
	OrderStatusCancelled OrderStatus = "cancelled"
	OrderStatusFailed    OrderStatus = "failed"
	OrderStatusRefunded  OrderStatus = "refunded"
)

type Order struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	UserID  uint   `gorm:"index;not null" json:"user_id"`
	User    User   `gorm:"foreignKey:UserID" json:"-"`
	OrderNo string `gorm:"uniqueIndex;size:64;not null" json:"order_no"`

	Status   OrderStatus `gorm:"size:16;default:pending;not null" json:"status"`
	Amount   float64     `gorm:"type:decimal(16,2);not null" json:"amount"`
	Currency string      `gorm:"size:8;default:VND" json:"currency"`

	TGChatID    int64 `gorm:"index" json:"tg_chat_id"`
	TGMessageID int64 `json:"tg_message_id"`
	WorkerTGID  int64 `gorm:"index" json:"worker_tg_id"`

	Metadata string `gorm:"type:text" json:"metadata"`
	Note     string `gorm:"size:1024" json:"note"`
}

func (Order) TableName() string { return "orders" }
