package models

import "time"

type WecomButler struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	ButlerType         string     `json:"butler_type" gorm:"size:32;default:order;comment:'public/order'"`
	Name               string     `json:"name" gorm:"size:64;not null"`
	Phone              string     `json:"phone" gorm:"size:32"`
	AvatarURL          string     `json:"avatar_url" gorm:"size:512"`
	CorpID             string     `json:"corp_id" gorm:"size:128"`
	AgentID            string     `json:"agent_id" gorm:"size:64"`
	AgentSecret        string     `json:"-" gorm:"size:256"`
	WecomUserID        string     `json:"wecom_userid" gorm:"size:128;index"`
	WecomName          string     `json:"wecom_name" gorm:"size:128"`
	ContactMode        string     `json:"contact_mode" gorm:"size:32;default:contact_me;comment:'contact_me/customer_service'"`
	CustomerServiceURL string     `json:"customer_service_url" gorm:"size:1024"`
	ContactWayConfigID string     `json:"contact_way_config_id" gorm:"size:128"`
	IsDefaultPublic    int        `json:"is_default_public" gorm:"default:0;index"`
	IsDefaultOrder     int        `json:"is_default_order" gorm:"default:0;index"`
	IsAssignable       int        `json:"is_assignable" gorm:"default:1;index"`
	Status             int        `json:"status" gorm:"default:1;index"`
	SortOrder          int        `json:"sort_order" gorm:"default:0"`
	CreatedAt          *time.Time `json:"created_at" gorm:"datetime(3)"`
	UpdatedAt          *time.Time `json:"updated_at" gorm:"datetime(3)"`
}

func (WecomButler) TableName() string {
	return "yesok_wecom_butlers"
}
