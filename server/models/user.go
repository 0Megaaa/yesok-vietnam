package models

import (
	"time"

	"github.com/google/uuid"
	"gorm.io/gorm"
)

const (
	// RoleAdmin 表示后台超级管理员角色，用于后台全部权限校验。
	RoleAdmin = "admin"
	// RoleUser 表示普通 C 端客户角色，保留给旧版 JWT 中间件兼容使用。
	RoleUser = "user"
	// RoleWorker 表示服务执行人员角色，保留给旧版后台接口兼容使用。
	RoleWorker = "worker"
	// RoleManager 表示新版 B 端业务经理角色，用于动态工作流处理。
	RoleManager = "manager"
)

// User 是旧版 Telegram 单表用户模型的兼容层。
type User struct {
	ID           uint           `gorm:"primarykey" json:"id"`
	CreatedAt    time.Time      `json:"created_at"`
	UpdatedAt    time.Time      `json:"updated_at"`
	DeletedAt    gorm.DeletedAt `gorm:"index" json:"-"`
	TGID         int64          `gorm:"uniqueIndex;not null" json:"tg_id"`
	Username     string         `gorm:"size:64" json:"username"`
	Role         string         `gorm:"size:16;default:user;index" json:"role"`
	Balance      float64        `gorm:"type:decimal(16,2);default:0" json:"balance"`
	FirstName    string         `gorm:"size:128" json:"first_name"`
	LastName     string         `gorm:"size:128" json:"last_name"`
	Language     string         `gorm:"size:16" json:"language"`
	AvatarURL    string         `gorm:"size:512" json:"avatar_url"`
	SessionToken string         `gorm:"size:128;index" json:"-"`
}

// TableName 固定旧版用户表名。
// 1.意图 -> 保留历史 Telegram 登录接口的最低兼容能力。
// 2.步骤 -> 返回 users，避免与新版 app_users/sys_users 身份域混淆。
// 3.返回 -> 数据库真实表名。
func (User) TableName() string { return "users" }

// BeforeCreate 为旧版用户自动补齐会话令牌。
// 1.意图 -> 保证旧版用户记录存在可追踪会话标识。
// 2.步骤 -> 当 SessionToken 为空时生成 UUID 字符串。
// 3.返回 -> nil 表示允许 GORM 继续创建。
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SessionToken == "" {
		u.SessionToken = uuid.NewString()
	}
	return nil
}
