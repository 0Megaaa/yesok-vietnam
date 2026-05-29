package models

import "time"

// AppUser 是 C 端客户表，承载微信小程序、Telegram Mini App 与未来 App 的统一客户身份。
type AppUser struct {
	ID             uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	WechatOpenID   string    `json:"wechat_open_id" gorm:"size:128;index;comment:'微信OpenID，用于微信小程序静默鉴权'"`
	TelegramID     string    `json:"telegram_id" gorm:"size:128;index;comment:'Telegram用户ID，预留Mini App登录'"`
	AppleID        string    `json:"apple_id" gorm:"size:128;index;comment:'Apple用户ID，预留iOS登录'"`
	Phone          string    `json:"phone" gorm:"size:32;index;comment:'手机号'"`
	Nickname       string    `json:"nickname" gorm:"size:64;comment:'用户昵称'"`
	AvatarURL      string    `json:"avatar_url" gorm:"size:512;comment:'用户头像地址'"`
	LoginProvider  string    `json:"login_provider" gorm:"size:32;index;default:wechat;comment:'登录提供方：wechat/telegram/apple/phone'"`
	ClientPlatform string    `json:"client_platform" gorm:"size:32;index;default:mp_weixin;comment:'客户端平台：mp_weixin/h5/app/admin'"`
	Balance        int64     `json:"balance" gorm:"default:0;comment:'账户余额，单位为分'"`
	VipLevel       int       `json:"vip_level" gorm:"default:0;comment:'VIP等级，0为普通用户'"`
	CreatedAt      time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt      time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定 C 端客户表名。
// 1.意图 -> 明确模型与数据库表的一一对应关系。
// 2.步骤 -> 返回 app_users，避免 GORM 复数化策略影响线上迁移。
// 3.返回 -> 数据库真实表名。
func (AppUser) TableName() string { return "app_users" }
