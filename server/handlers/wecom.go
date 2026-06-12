package handlers

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	wecompkg "yesok-vietnam/server/pkg/wecom"
)

func normalizeWecomContactMode(mode string) string {
	mode = strings.TrimSpace(mode)
	if mode == "customer_service" {
		return "customer_service"
	}
	return "contact_me"
}

func buildWecomContactPayload(butler models.WecomButler, contactType string) gin.H {
	return gin.H{
		"corp_id":               strings.TrimSpace(butler.CorpID),
		"contact_mode":          normalizeWecomContactMode(butler.ContactMode),
		"contact_way_config_id": strings.TrimSpace(butler.ContactWayConfigID),
		"service_url":           strings.TrimSpace(butler.CustomerServiceURL),
		"contact_type":          contactType,
		"butler_id":             butler.ID,
		"butler_name":           strings.TrimSpace(butler.Name),
	}
}

func buildOrderButlerContactDescription(order models.Order, butler models.WecomButler) string {
	return strings.Join([]string{
		`<div class="gray">客户已在小程序点击联系专属管家</div>`,
		fmt.Sprintf(`<div class="normal">订单号：%s</div>`, order.OrderNo),
		fmt.Sprintf(`<div class="normal">服务：%s</div>`, order.ServiceName),
		fmt.Sprintf(`<div class="normal">客户：%s</div>`, order.ContactName),
		fmt.Sprintf(`<div class="normal">电话：%s</div>`, order.ContactPhone),
		fmt.Sprintf(`<div class="normal">当前节点：%s</div>`, order.CurrentStage),
		fmt.Sprintf(`<div class="highlight">订单金额：%s</div>`, fmt.Sprintf("%d", order.TotalAmount)),
		`<div class="highlight">请及时联系客户并跟进订单。</div>`,
	}, "")
}

// buildContactButlerTextMessage builds a plain text message for contact butler notification
func buildContactButlerTextMessage(order models.Order) string {
	return strings.Join([]string{
		"【YesOK 客户已联系专属管家】",
		fmt.Sprintf("订单号：%s", order.OrderNo),
		fmt.Sprintf("服务名称：%s", order.ServiceName),
		fmt.Sprintf("客户姓名：%s", order.ContactName),
		fmt.Sprintf("客户电话：%s", order.ContactPhone),
		fmt.Sprintf("当前节点：%s", order.CurrentStage),
		fmt.Sprintf("订单金额：%s", formatMoney(order.TotalAmount)),
		"客户已在小程序点击联系专属管家，请及时跟进。",
	}, "\n")
}

func ClientWecomPublicContact(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var butler models.WecomButler
		if err := db.Where("butler_type = ? AND is_default_public = ? AND status = ?", "public", 1, 1).
			Order("sort_order asc, id asc").
			First(&butler).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "public wecom service not configured"})
			return
		}

		c.JSON(http.StatusOK, buildWecomContactPayload(butler, "public"))
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

		var selectedButler models.WecomButler
		contactType := "order_default"

		if order.ButlerID > 0 {
			if err := db.Where("id = ? AND status = ?", order.ButlerID, 1).First(&selectedButler).Error; err == nil {
				contactType = "order_butler"
			}
		}

		if selectedButler.ID == 0 {
			if err := db.Where("butler_type = ? AND is_default_order = ? AND status = ?", "order", 1, 1).
				Order("sort_order asc, id asc").
				First(&selectedButler).Error; err != nil {
				c.JSON(http.StatusNotFound, gin.H{"error": "order wecom service not configured"})
				return
			}
			contactType = "order_default"
		}

		now := time.Now()

		if err := db.Model(&models.Order{}).
			Where("id = ?", order.ID).
			Updates(map[string]any{
				"butler_contacted_at": now,
				"updated_at":          now,
			}).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to update butler contact time")
			return
		}

		var recentCount int64
		_ = db.Model(&models.OrderTimeline{}).
			Where("order_id = ? AND action_name = ? AND created_at >= ?", order.ID, "contact_order_butler", time.Now().Add(-3*time.Minute)).
			Count(&recentCount).Error

		messageSent := false
		warning := ""
		warningDetail := ""

		if recentCount == 0 {
			payload := gin.H{
				"contact_type":          contactType,
				"butler_id":             selectedButler.ID,
				"butler_name":           strings.TrimSpace(selectedButler.Name),
				"wecom_userid":          strings.TrimSpace(selectedButler.WecomUserID),
				"contact_mode":          normalizeWecomContactMode(selectedButler.ContactMode),
				"contact_way_config_id": strings.TrimSpace(selectedButler.ContactWayConfigID),
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

			client := wecompkg.New(selectedButler.CorpID, selectedButler.AgentID, selectedButler.AgentSecret)
			if client.Enabled() && strings.TrimSpace(selectedButler.WecomUserID) != "" {
				linkURL := buildAdminOrderURL(order.ID)

				ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
				defer cancel()

				var sendErr error

				if linkURL != "" {
					sendErr = client.SendTextCard(
						ctx,
						selectedButler.WecomUserID,
						"YesOK 客户已联系专属管家",
						buildOrderButlerContactDescription(order, selectedButler),
						linkURL,
					)
				} else {
					sendErr = client.SendText(
						ctx,
						selectedButler.WecomUserID,
						buildContactButlerTextMessage(order),
					)
				}

				if sendErr != nil {
					warning = "企业微信通知发送失败"
					warningDetail = sendErr.Error()
					log.Printf("[wecom] contact butler notify failed: order_id=%d butler_id=%d to_user=%s err=%v",
						order.ID,
						selectedButler.ID,
						selectedButler.WecomUserID,
						sendErr,
					)
				} else {
					messageSent = true
				}
			} else {
				warning = "企业微信通知配置不完整"
				warningDetail = fmt.Sprintf(
					"corp_id=%t agent_id=%t agent_secret=%t wecom_userid=%t",
					strings.TrimSpace(selectedButler.CorpID) != "",
					strings.TrimSpace(selectedButler.AgentID) != "",
					strings.TrimSpace(selectedButler.AgentSecret) != "",
					strings.TrimSpace(selectedButler.WecomUserID) != "",
				)
				log.Printf("[wecom] contact butler notify config missing: order_id=%d butler_id=%d detail=%s",
					order.ID,
					selectedButler.ID,
					warningDetail,
				)
			}
		} else {
			warning = "recently notified"
		}

		response := buildWecomContactPayload(selectedButler, contactType)
		response["message_sent"] = messageSent
		if warning != "" {
			response["warning"] = warning
		}
		if warningDetail != "" && gin.Mode() != gin.ReleaseMode {
			response["warning_detail"] = warningDetail
		}
		c.JSON(http.StatusOK, response)
	}
}
