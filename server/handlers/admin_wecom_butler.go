package handlers

import (
	"context"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	wecompkg "yesok-vietnam/server/pkg/wecom"
)

type AdminAssignButlerRequest struct {
	ButlerID uint `json:"butler_id" binding:"required"`
}

func AdminListWecomButlers(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&models.WecomButler{})

		statusQuery := strings.TrimSpace(c.Query("status"))
		if statusQuery == "" {
			query = query.Where("status = ?", 1)
		} else if statusValue, err := strconv.Atoi(statusQuery); err == nil {
			query = query.Where("status = ?", statusValue)
		}

		if butlerType := strings.TrimSpace(c.Query("butler_type")); butlerType != "" {
			query = query.Where("butler_type = ?", butlerType)
		}
		if assignableQuery := strings.TrimSpace(c.Query("is_assignable")); assignableQuery != "" {
			if assignable, err := strconv.Atoi(assignableQuery); err == nil {
				query = query.Where("is_assignable = ?", assignable)
			}
		}

		var butlers []models.WecomButler
		if err := query.Order("sort_order asc, id asc").Find(&butlers).Error; err != nil {
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to fetch wecom butlers")
			return
		}

		list := make([]gin.H, 0, len(butlers))
		for _, butler := range butlers {
			list = append(list, gin.H{
				"id":                    butler.ID,
				"name":                  butler.Name,
				"phone":                 butler.Phone,
				"avatar_url":            butler.AvatarURL,
				"corp_id":               butler.CorpID,
				"agent_id":              butler.AgentID,
				"wecom_userid":          butler.WecomUserID,
				"wecom_name":            butler.WecomName,
				"contact_mode":          butler.ContactMode,
				"contact_way_config_id": butler.ContactWayConfigID,
				"customer_service_url":  butler.CustomerServiceURL,
				"butler_type":           butler.ButlerType,
				"is_assignable":         butler.IsAssignable,
				"status":                butler.Status,
				"sort_order":            butler.SortOrder,
			})
		}

		c.JSON(http.StatusOK, gin.H{"list": list, "total": len(list)})
	}
}

func AdminAssignOrderButler(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, err := strconv.ParseUint(c.Param("id"), 10, 64)
		if err != nil || id == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "invalid order id")
			return
		}

		var req AdminAssignButlerRequest
		if err := c.ShouldBindJSON(&req); err != nil || req.ButlerID == 0 {
			httpError(c, http.StatusBadRequest, ErrCodeInvalidRequest, "butler_id is required")
			return
		}

		var updatedOrder models.Order
		var assignedButler models.WecomButler
		now := time.Now()

		err = db.Transaction(func(tx *gorm.DB) error {
			var order models.Order
			if err := tx.First(&order, id).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return newUserError(ErrCodeNotFound, "order not found")
				}
				return err
			}

			if err := tx.Where("id = ? AND butler_type = ? AND is_assignable = ? AND status = ?", req.ButlerID, "order", 1, 1).
				First(&assignedButler).Error; err != nil {
				if err == gorm.ErrRecordNotFound {
					return newUserError(ErrCodeNotFound, "assignable order butler not found")
				}
				return err
			}

			beforeStage := strings.TrimSpace(order.CurrentStage)
			afterStage := beforeStage
			nextMacroStatus := strings.TrimSpace(order.MacroStatus)

			if beforeStage == "wait_butler_assign" {
				var assignNode models.SysWorkflowNode
				if err := tx.Where(
					"service_id = ? AND stage_code = ? AND action_name = ? AND (executor_role = ? OR executor_role = 'both')",
					order.ServiceID,
					beforeStage,
					"assign_butler",
					"admin",
				).Order("sort_order ASC, id ASC").First(&assignNode).Error; err == nil {
					if strings.TrimSpace(assignNode.TargetStatus) != "" {
						afterStage = strings.TrimSpace(assignNode.TargetStatus)
					}
					if strings.TrimSpace(assignNode.MacroStatus) != "" {
						nextMacroStatus = strings.TrimSpace(assignNode.MacroStatus)
					}
				}
			}

			updates := map[string]any{
				"butler_id":           assignedButler.ID,
				"butler_name":         strings.TrimSpace(assignedButler.Name),
				"butler_wecom_userid": strings.TrimSpace(assignedButler.WecomUserID),
				"butler_contact_url":  strings.TrimSpace(assignedButler.CustomerServiceURL),
				"butler_assigned_at":  now,
				"updated_at":          now,
			}
			if beforeStage == "wait_butler_assign" {
				updates["current_stage"] = afterStage
				updates["macro_status"] = nextMacroStatus
			}
			if err := tx.Model(&order).Updates(updates).Error; err != nil {
				return err
			}

			payload := gin.H{
				"butler_id":             assignedButler.ID,
				"butler_name":           strings.TrimSpace(assignedButler.Name),
				"wecom_userid":          strings.TrimSpace(assignedButler.WecomUserID),
				"contact_mode":          assignedButler.ContactMode,
				"contact_way_config_id": assignedButler.ContactWayConfigID,
				"customer_service_url":  strings.TrimSpace(assignedButler.CustomerServiceURL),
			}
			if err := tx.Create(&models.OrderTimeline{
				OrderID:      order.ID,
				BeforeStatus: beforeStage,
				AfterStatus:  afterStage,
				Operator:     "admin",
				Remark:       fmt.Sprintf("已分配管家：%s", strings.TrimSpace(assignedButler.Name)),
				ActionName:   "assign_butler",
				Payload:      marshalJSON(payload),
				AuditStatus:  models.AuditStatusApproved,
				CreatedAt:    &now,
				UpdatedAt:    &now,
			}).Error; err != nil {
				return err
			}

			if err := tx.First(&updatedOrder, order.ID).Error; err != nil {
				return err
			}
			return nil
		})
		if err != nil {
			if userErr, ok := err.(*userError); ok {
				httpError(c, http.StatusBadRequest, userErr.code, userErr.message)
				return
			}
			httpError(c, http.StatusInternalServerError, ErrCodeInternalError, "failed to assign butler")
			return
		}

		messageSent := false
		warning := ""
		client := wecompkg.New(assignedButler.CorpID, assignedButler.AgentID, assignedButler.AgentSecret)
		if client.Enabled() && strings.TrimSpace(assignedButler.WecomUserID) != "" {
			adminOrigin := strings.TrimSpace(os.Getenv("ADMIN_ORIGIN_URL"))
			linkURL := adminOrigin
			if adminOrigin != "" {
				linkURL = strings.TrimRight(adminOrigin, "/") + fmt.Sprintf("/admin/orders/detail?id=%d", updatedOrder.ID)
			}
			description := strings.Join([]string{
				fmt.Sprintf("<div class=\"gray\">订单号：%s</div>", updatedOrder.OrderNo),
				fmt.Sprintf("<div class=\"normal\">服务名称：%s</div>", updatedOrder.ServiceName),
				fmt.Sprintf("<div class=\"normal\">客户姓名：%s</div>", updatedOrder.ContactName),
				fmt.Sprintf("<div class=\"normal\">客户电话：%s</div>", updatedOrder.ContactPhone),
				fmt.Sprintf("<div class=\"normal\">当前节点：%s</div>", updatedOrder.CurrentStage),
				fmt.Sprintf("<div class=\"highlight\">订单金额：%s</div>", formatMoney(updatedOrder.TotalAmount)),
			}, "")
			ctx, cancel := context.WithTimeout(c.Request.Context(), 8*time.Second)
			defer cancel()
			if err := client.SendTextCard(ctx, assignedButler.WecomUserID, "YesOK 待管家沟通", description, linkURL); err != nil {
				warning = "企业微信通知发送失败"
			} else {
				messageSent = true
			}
		} else {
			warning = "企业微信通知发送失败"
		}

		response := gin.H{
			"ok":           true,
			"message_sent": messageSent,
			"order":        buildOrderPayloadForRole(db, updatedOrder, "admin"),
		}
		if warning != "" {
			response["warning"] = warning
		}
		c.JSON(http.StatusOK, response)
	}
}
