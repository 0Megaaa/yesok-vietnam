package models

import "time"

// SysArticle 是资讯文章表，用于驱动 C 端首页资讯模块和资讯 Tab 列表。
type SysArticle struct {
	ID        uint      `gorm:"primaryKey;comment:'主键ID'" json:"id"`
	Title     string    `gorm:"size:160;not null;comment:'文章标题'" json:"title"`
	CoverImg  string    `gorm:"size:255;comment:'封面图片地址'" json:"cover_img"`
	Summary   string    `gorm:"size:300;comment:'摘要'" json:"summary"`
	Content   string    `gorm:"type:text;comment:'正文内容'" json:"content"`
	Category  string    `gorm:"size:80;index;comment:'资讯分类'" json:"category"`
	Author    string    `gorm:"size:80;comment:'作者'" json:"author"`
	Status    int       `gorm:"comment:'状态：1发布 0草稿'" json:"status"`
	SortOrder int       `gorm:"default:0;comment:'排序值'" json:"sort_order"`
	ViewCount int       `gorm:"default:0;comment:'浏览次数'" json:"view_count"`
	CreatedAt time.Time `gorm:"comment:'创建时间'" json:"created_at"`
	UpdatedAt time.Time `gorm:"comment:'更新时间'" json:"updated_at"`
}

// TableName 指定资讯文章表名。
// 1.意图 -> 保证 GORM 使用业务指定的 sys_articles 表。
// 2.步骤 -> 返回固定表名字符串，避免自动复数化差异。
// 3.返回 -> sys_articles。
func (SysArticle) TableName() string { return "sys_articles" }
