package handlers

import (
	"io"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
	"yesok-vietnam/server/pkg/storage"
)

// ClientUploadOrderMaterial 处理 C 端用户上传订单资料文件。
// POST /api/v1/client/orders/:id/materials/upload
// Content-Type: multipart/form-data
// 必填参数：file（图片文件）
// 可选参数：field_key（表单字段标识）
//
// 安全保证：
// 1. 必须在登录状态下传
// 2. 只能上传自己订单的文件
// 3. 仅允许 jpg/jpeg/png
// 4. MIME 内容校验
// 5. 单文件 10MB 限制
// 6. 文件名完全随机化，防覆盖
// 7. 返回相对路径，不暴露服务器绝对路径
func ClientUploadOrderMaterial(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 鉴权：当前登录的 app_user id
		appUserVal, ok := c.Get("uid")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}
		appUserID, ok := appUserVal.(uint)
		if !ok || appUserID == 0 {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		orderIDStr := c.Param("id")
		orderID, err := parseUint(orderIDStr)
		if err != nil || orderID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		// 只能操作自己的订单
		var order models.Order
		if err := db.Where("id = ? AND app_user_id = ?", orderID, appUserID).First(&order).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在或无权访问"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询订单失败"})
			return
		}

		// 查服务获取 service_code
		var service models.SysService
		if err := db.First(&service, order.ServiceID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "服务不存在"})
			return
		}

		// 接收文件
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
			return
		}

		if fileHeader.Size > storage.MaxUploadSize() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片不能超过10MB"})
			return
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !storage.IsAllowedFileExt(ext) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg、jpeg、png 格式"})
			return
		}

		src, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败"})
			return
		}
		defer src.Close()

		// MIME 内容校验（防止伪造扩展名）
		buffer := make([]byte, 512)
		n, _ := src.Read(buffer)
		mimeType := http.DetectContentType(buffer[:n])
		if !storage.IsAllowedMimeType(mimeType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型不合法，仅支持 jpg、jpeg、png"})
			return
		}

		if _, err := src.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}

		fieldKey := strings.TrimSpace(c.PostForm("field_key"))
		serviceCode := safePathPart(service.ServiceCode)
		filename := storage.BuildMaterialFilename(order.OrderNo, fieldKey, fileHeader.Filename)

		storageRoot := storage.MaterialStorageRoot()
		dir := filepath.Join(storageRoot, serviceCode)
		if err := storage.EnsureDir(dir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
			return
		}

		dstPath := filepath.Join(dir, filename)
		out, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
			return
		}

		relativePath := filepath.Join(serviceCode, filename)
		url := "/material/" + strings.ReplaceAll(relativePath, string(filepath.Separator), "/")

		c.JSON(http.StatusOK, gin.H{
			"url":           url,
			"path":          relativePath,
			"filename":      filename,
			"original_name": fileHeader.Filename,
			"size":          fileHeader.Size,
			"mime_type":     mimeType,
		})
	}
}

// safePathPart 清理路径组成部分，防止目录遍历攻击。
func safePathPart(s string) string {
	s = strings.ReplaceAll(s, "..", "_")
	s = strings.ReplaceAll(s, "/", "_")
	s = strings.ReplaceAll(s, "\\", "_")
	s = strings.ReplaceAll(s, "\x00", "")
	return strings.TrimSpace(s)
}

// AdminUploadOrderMaterial 处理 B 端管理员上传订单资料文件。
// POST /api/v1/admin/orders/:id/materials/upload
// Content-Type: multipart/form-data
// 必填参数：file（图片文件）
// 可选参数：field_key（表单字段标识）
//
// 安全保证：
// 1. 必须在 B 端管理员登录状态下传
// 2. 仅允许 jpg/jpeg/png
// 3. MIME 内容校验
// 4. 单文件 10MB 限制
// 5. 文件名随机化，防覆盖
// 6. 返回 /material/... 相对路径，不暴露服务器绝对路径
func AdminUploadOrderMaterial(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// 鉴权：必须是 B 端管理员
		_, ok := c.Get("uid")
		if !ok {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "未登录"})
			return
		}

		orderIDStr := c.Param("id")
		orderID, err := parseUint(orderIDStr)
		if err != nil || orderID == 0 {
			c.JSON(http.StatusBadRequest, gin.H{"error": "无效的订单ID"})
			return
		}

		// 查询订单
		var order models.Order
		if err := db.First(&order, orderID).Error; err != nil {
			if err == gorm.ErrRecordNotFound {
				c.JSON(http.StatusNotFound, gin.H{"error": "订单不存在"})
				return
			}
			c.JSON(http.StatusInternalServerError, gin.H{"error": "查询订单失败"})
			return
		}

		// 查询服务获取 service_code
		var service models.SysService
		if err := db.First(&service, order.ServiceID).Error; err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "服务不存在"})
			return
		}

		// 接收文件
		fileHeader, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "请选择要上传的图片"})
			return
		}

		if fileHeader.Size > storage.MaxUploadSize() {
			c.JSON(http.StatusBadRequest, gin.H{"error": "图片不能超过10MB"})
			return
		}

		ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
		if !storage.IsAllowedFileExt(ext) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "仅支持 jpg、jpeg、png 格式"})
			return
		}

		src, err := fileHeader.Open()
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "读取文件失败"})
			return
		}
		defer src.Close()

		// MIME 内容校验
		buffer := make([]byte, 512)
		n, _ := src.Read(buffer)
		mimeType := http.DetectContentType(buffer[:n])
		if !storage.IsAllowedMimeType(mimeType) {
			c.JSON(http.StatusBadRequest, gin.H{"error": "文件类型不合法，仅支持 jpg、jpeg、png"})
			return
		}

		if _, err := src.Seek(0, io.SeekStart); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "读取文件失败"})
			return
		}

		fieldKey := strings.TrimSpace(c.PostForm("field_key"))
		serviceCode := safePathPart(service.ServiceCode)
		filename := storage.BuildMaterialFilename(order.OrderNo, fieldKey, fileHeader.Filename)

		storageRoot := storage.MaterialStorageRoot()
		dir := filepath.Join(storageRoot, serviceCode)
		if err := storage.EnsureDir(dir); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "创建上传目录失败"})
			return
		}

		dstPath := filepath.Join(dir, filename)
		out, err := os.Create(dstPath)
		if err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "保存文件失败"})
			return
		}
		defer out.Close()

		if _, err := io.Copy(out, src); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "写入文件失败"})
			return
		}

		relativePath := filepath.Join(serviceCode, filename)
		url := "/material/" + strings.ReplaceAll(relativePath, string(filepath.Separator), "/")

		c.JSON(http.StatusOK, gin.H{
			"url":           url,
			"path":          relativePath,
			"filename":      filename,
			"original_name": fileHeader.Filename,
			"size":          fileHeader.Size,
			"mime_type":     mimeType,
		})
	}
}
