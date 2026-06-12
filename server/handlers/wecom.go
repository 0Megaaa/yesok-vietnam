package handlers

import (
	"fmt"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

func ClientWecomPublicContact(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var butler models.WecomButler
		if err := db.Where("butler_type = ? AND is_default_public = ? AND status = ?", "public", 1, 1).
			Order("sort_order asc, id asc").
			First(&butler).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "public wecom service not configured"})
			return
		}

		c.JSON(http.StatusOK, gin.H{
			"corp_id":      strings.TrimSpace(butler.CorpID),
			"service_url":  strings.TrimSpace(butler.CustomerServiceURL),
			"contact_type": "public",
		})
	}
}

func ClientOrderWecomContact(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		uidVal, ok := c.Get("uid")
		if !ok {
			httpError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "unauthorized")
			return
		}
		uid, _ := uidVal.(uint)
		if uid == 0 {
			httpError(c, http.StatusUnauthorized, ErrCodeUnauthorized, "invalid user")
			return
		}

		id, err := parseUint(c.Param("id"))
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var order models.Order
		if err := db.Where("id = ? AND app_user_id = ?", id, uid).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				httpError(c, http.StatusNotFound, ErrCodeNotFound, "order not found")
				return
			}
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to query order")
			return
		}

		response := gin.H{}
		payload := gin.H{}
		now := time.Now()

		if order.ButlerID > 0 && strings.TrimSpace(order.ButlerContactURL) != "" {
			contactType := "order_butler"
			corpID := ""
			butlerName := strings.TrimSpace(order.ButlerName)

			var butler models.WecomButler
			if err := db.Where("id = ?", order.ButlerID).First(&butler).Error; err == nil {
				corpID = strings.TrimSpace(butler.CorpID)
				if butlerName == "" {
					butlerName = strings.TrimSpace(butler.Name)
				}
			}
			if corpID == "" {
				var defaultOrderButler models.WecomButler
				if err := db.Where("butler_type = ? AND is_default_order = ? AND status = ?", "order", 1, 1).
					Order("sort_order asc, id asc").
					First(&defaultOrderButler).Error; err == nil {
					corpID = strings.TrimSpace(defaultOrderButler.CorpID)
				}
			}

			response = gin.H{
				"corp_id":      corpID,
				"service_url":  strings.TrimSpace(order.ButlerContactURL),
				"contact_type": contactType,
				"butler_name":  butlerName,
			}
			payload = gin.H{
				"contact_type": contactType,
				"butler_id":    order.ButlerID,
				"butler_name":  butlerName,
			}
		} else {
			var butler models.WecomButler
			if err := db.Where("butler_type = ? AND is_default_order = ? AND status = ?", "order", 1, 1).
				Order("sort_order asc, id asc").
				First(&butler).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "order wecom service not configured"})
				return
			}

			response = gin.H{
				"corp_id":      strings.TrimSpace(butler.CorpID),
				"service_url":  strings.TrimSpace(butler.CustomerServiceURL),
				"contact_type": "order_default",
				"butler_name":  strings.TrimSpace(butler.Name),
			}
			payload = gin.H{
				"contact_type": "order_default",
				"butler_id":    butler.ID,
				"butler_name":  strings.TrimSpace(butler.Name),
			}
		}

		if err := db.Create(&models.OrderTimeline{
			OrderID:       order.ID,
			BeforeStatus:  order.CurrentStage,
			AfterStatus:   order.CurrentStage,
			Operator:      fmt.Sprintf("client:%d", uid),
			Remark:        "用户点击联系专属管家",
			ActionName:    "contact_order_butler",
			Payload:       marshalJSON(payload),
			AuditStatus:   models.AuditStatusApproved,
			CreatedAt:     &now,
			UpdatedAt:     &now,
			AuditOperator: "",
		}).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to write timeline")
			return
		}

		c.JSON(http.StatusOK, response)
	}
}
