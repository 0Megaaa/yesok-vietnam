package models

import "time"

type WecomButler struct {
	ID                 uint       `json:"id" gorm:"primaryKey"`
	ButlerType         string     `json:"butler_type" gorm:"column:butler_type;size:32;default:order;comment:'public/order'"`
	Name               string     `json:"name" gorm:"size:64;not null"`
	Phone              string     `json:"phone" gorm:"column:phone;size:32"`
	AvatarURL          string     `json:"avatar_url" gorm:"column:avatar_url;size:512"`
	CorpID             string     `json:"corp_id" gorm:"column:corp_id;size:128"`
	AgentID            string     `json:"agent_id" gorm:"column:agent_id;size:64"`
	AgentSecret        string     `json:"-" gorm:"column:agent_secret;size:256"`
	WecomUserID        string     `json:"wecom_userid" gorm:"column:wecom_userid;size:128;index"`
	WecomName          string     `json:"wecom_name" gorm:"column:wecom_name;size:128"`
	ContactMode        string     `json:"contact_mode" gorm:"column:contact_mode;size:32;default:contact_me;comment:'contact_me/customer_service'"`
	CustomerServiceURL string     `json:"customer_service_url" gorm:"column:customer_service_url;size:1024"`
	ContactWayConfigID string     `json:"contact_way_config_id" gorm:"column:contact_way_config_id;size:128"`
	IsDefaultPublic    int        `json:"is_default_public" gorm:"column:is_default_public;default:0;index"`
	IsDefaultOrder     int        `json:"is_default_order" gorm:"column:is_default_order;default:0;index"`
	IsAssignable       int        `json:"is_assignable" gorm:"column:is_assignable;default:1;index"`
	Status             int        `json:"status" gorm:"column:status;default:1;index"`
	SortOrder          int        `json:"sort_order" gorm:"column:sort_order;default:0"`
	CreatedAt          *time.Time `json:"created_at" gorm:"column:created_at;datetime(3)"`
	UpdatedAt          *time.Time `json:"updated_at" gorm:"column:updated_at;datetime(3)"`
}

func (WecomButler) TableName() string {
	return "yesok_wecom_butlers"
}
