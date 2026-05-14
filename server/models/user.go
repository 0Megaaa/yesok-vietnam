package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	RoleAdmin  = "admin"
	RoleUser   = "user"
	RoleWorker = "worker"
)

type User struct {
	ID        uint           `gorm:"primarykey" json:"id"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `gorm:"index" json:"-"`

	TGID         int64   `gorm:"uniqueIndex;not null" json:"tg_id"`
	Username     string  `gorm:"size:128" json:"username"`
	Role         string  `gorm:"size:16;default:user;not null" json:"role"`
	Balance      float64 `gorm:"type:decimal(16,2);default:0" json:"balance"`
	FirstName    string  `gorm:"size:256" json:"first_name"`
	LastName     string  `gorm:"size:256" json:"last_name"`
	Language     string  `gorm:"size:8;default:en" json:"language"`
	AvatarURL    string  `gorm:"size:512" json:"avatar_url"`
	SessionToken string  `gorm:"size:256;index" json:"-"`
}

func (User) TableName() string { return "users" }

func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SessionToken == "" {
		u.SessionToken = uuid.NewString()
	}
	return nil
}
