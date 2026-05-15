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

// AppUser 是 C 端客户表，承载微信小程序、Telegram Mini App 与未来 App 的统一客户身份。
type AppUser struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	WechatOpenID string    `json:"wechat_open_id" gorm:"size:128;uniqueIndex;comment:'微信OpenID，用于微信小程序静默鉴权'"`
	TelegramID   string    `json:"telegram_id" gorm:"size:128;index;comment:'Telegram用户ID，预留Mini App登录'"`
	AppleID      string    `json:"apple_id" gorm:"size:128;index;comment:'Apple用户ID，预留iOS登录'"`
	Phone        string    `json:"phone" gorm:"size:32;index;comment:'手机号'"`
	Nickname     string    `json:"nickname" gorm:"size:64;comment:'用户昵称'"`
	AvatarURL    string    `json:"avatar_url" gorm:"size:512;comment:'用户头像地址'"`
	Balance      int64     `json:"balance" gorm:"default:0;comment:'账户余额，单位为分'"`
	VipLevel     int       `json:"vip_level" gorm:"default:0;comment:'VIP等级，0为普通用户'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定 C 端客户表名，避免 GORM 复数化策略变化影响线上迁移。
// 实现步骤：
// 1. 明确模型与表的一一对应关系。
// 2. 返回数据库真实表名，便于 SQL 脚本和 AutoMigrate 对齐。
func (AppUser) TableName() string { return "app_users" }

// SysUser 是 B 端员工表，后续后台登录将从硬编码账号迁移到该表。
type SysUser struct {
	ID           uint      `json:"id" gorm:"primaryKey;comment:'主键ID'"`
	Username     string    `json:"username" gorm:"size:64;uniqueIndex;not null;comment:'B端员工登录账号'"`
	PasswordHash string    `json:"-" gorm:"size:255;not null;comment:'密码哈希值，禁止明文存储'"`
	RealName     string    `json:"real_name" gorm:"size:64;comment:'员工真实姓名'"`
	Role         string    `json:"role" gorm:"size:32;default:'manager';comment:'员工角色：admin管理员，manager经理'"`
	Status       int       `json:"status" gorm:"default:1;comment:'账号状态：1启用，0禁用'"`
	CreatedAt    time.Time `json:"created_at" gorm:"comment:'创建时间'"`
	UpdatedAt    time.Time `json:"updated_at" gorm:"comment:'更新时间'"`
}

// TableName 固定 B 端员工表名。
// 实现步骤：
// 1. 保持后台账号模型独立于 C 端客户模型。
// 2. 返回 sys_users，方便宝塔部署时直接执行同名 SQL。
func (SysUser) TableName() string { return "sys_users" }

// User 是旧版 Telegram 单表用户模型的兼容层。
// 实现步骤：
// 1. 暂时保留旧路由依赖的字段，保证现有 handlers 可以继续编译。
// 2. 新业务严禁继续扩展该模型，新增 C 端能力统一写入 AppUser。
// 3. 等后台接口完全切换到 AppUser/SysUser 后，可以安全删除该兼容结构。
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

// TableName 固定旧版用户表名，避免影响历史数据读取。
// 实现步骤：
// 1. 对旧版 handlers 保持 users 表映射。
// 2. 与新版 app_users/sys_users 分离，避免身份域混淆。
func (User) TableName() string { return "users" }

// BeforeCreate 为旧版用户自动补齐会话令牌。
// 实现步骤：
// 1. 判断 SessionToken 是否已由调用方传入。
// 2. 若为空，则生成 UUID 字符串。
// 3. 返回 nil，让 GORM 继续执行插入流程。
func (u *User) BeforeCreate(tx *gorm.DB) error {
	if u.SessionToken == "" {
		u.SessionToken = uuid.NewString()
	}
	return nil
}
