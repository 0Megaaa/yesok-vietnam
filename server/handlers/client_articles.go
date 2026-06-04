package handlers

import (
	"net/http"
	"strconv"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

// ClientListArticles 输出已发布资讯列表，驱动 C 端首页和资讯 Tab。
// 1.意图 -> 消灭 C 端资讯标题、封面、摘要和分类硬编码。
// 2.步骤 -> 按 status=1 查询 sys_articles，支持 category 和 limit 参数并按排序输出。
// 3.返回 -> 资讯数组，包含标题、封面、摘要、分类、作者、浏览量与发布时间。
func ClientListArticles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&models.SysArticle{}).Where("status = ?", 1)
		if category := strings.TrimSpace(c.Query("category")); category != "" && category != "all" {
			query = query.Where("category = ?", category)
		}
		limit := 20
		if rawLimit := strings.TrimSpace(c.Query("limit")); rawLimit != "" {
			if parsed, err := strconv.Atoi(rawLimit); err == nil && parsed > 0 && parsed <= 50 {
				limit = parsed
			}
		}

		var articles []models.SysArticle
		if err := query.Order("sort_order asc, created_at desc").Limit(limit).Find(&articles).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to fetch articles"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"list": articles})
	}
}

// ClientGetArticle 返回 C 端资讯详情，并增加浏览次数。
func ClientGetArticle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article id"})
			return
		}

		var article models.SysArticle
		if err := db.Where("id = ? AND status = ?", id, 1).First(&article).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}

		_ = db.Model(&article).UpdateColumn("view_count", gorm.Expr("view_count + ?", 1)).Error
		article.ViewCount++

		c.JSON(http.StatusOK, gin.H{"article": article})
	}
}
