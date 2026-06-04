package handlers

import (
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
	"yesok-vietnam/server/models"
)

type SaveDictTypeRequest struct {
	DictName string `json:"dict_name"`
	DictCode string `json:"dict_code"`
	Remark   string `json:"remark"`
	Status   int    `json:"status"`
}

type SaveDictDataRequest struct {
	DictCode  string          `json:"dict_code"`
	DictLabel string          `json:"dict_label"`
	DictValue string          `json:"dict_value"`
	SortOrder int             `json:"sort_order"`
	Status    int             `json:"status"`
	Remark    string          `json:"remark"`
	MetaJSON  models.JSONText `json:"meta_json"`
}

type SaveArticleRequest struct {
	Title     string `json:"title"`
	CoverImg  string `json:"cover_img"`
	Summary   string `json:"summary"`
	Content   string `json:"content"`
	Category  string `json:"category"`
	Author    string `json:"author"`
	Status    int    `json:"status"`
	SortOrder int    `json:"sort_order"`
	ViewCount int    `json:"view_count"`
}

// UploadFile 保存后台上传图片到本地 uploads 目录。
// 1.意图 -> 支撑资讯封面和服务图片从 B 端真实上传，而不是前端硬编码地址。
// 2.步骤 -> 校验 multipart 文件、创建 uploads 日期目录、保存文件并拼接静态访问 URL。
// 3.返回 -> url/name/size，供前端表单写入 cover_img 或 cover_image 字段。
func UploadFile(db *gorm.DB) gin.HandlerFunc {
	_ = db
	return func(c *gin.Context) {
		file, err := c.FormFile("file")
		if err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "file is required"})
			return
		}
		ext := strings.ToLower(filepath.Ext(file.Filename))
		if ext == "" {
			ext = ".png"
		}
		safeBase := sanitizeUploadName(strings.TrimSuffix(file.Filename, filepath.Ext(file.Filename)))
		if safeBase == "" {
			safeBase = "image"
		}
		day := time.Now().Format("20060102")
		dir := filepath.Join("uploads", day)
		if err := os.MkdirAll(dir, 0755); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to prepare upload dir"})
			return
		}
		name := fmt.Sprintf("%s_%d%s", safeBase, time.Now().UnixNano(), ext)
		savePath := filepath.Join(dir, name)
		if err := c.SaveUploadedFile(file, savePath); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to save file"})
			return
		}
		url := "/" + filepath.ToSlash(savePath)
		c.JSON(http.StatusOK, gin.H{"url": url, "name": name, "size": file.Size})
	}
}

// AdminListDictTypes 返回后台字典类型列表。
// 1.意图 -> 让后台可维护服务分类、资讯分类等业务枚举类型。
// 2.步骤 -> 查询 sys_dict_types 并按 ID 倒序输出。
// 3.返回 -> list 格式字典类型数组。
func AdminListDictTypes(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var list []models.SysDictType
		db.Order("id desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminSaveDictType 新增字典类型。
// 1.意图 -> 支撑后台扩展新的枚举分组。
// 2.步骤 -> 绑定 JSON、保留显式状态并写入 sys_dict_types。
// 3.返回 -> 新创建的字典类型记录。
func AdminSaveDictType(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveDictTypeRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.DictCode) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dict type"})
			return
		}
		item := models.SysDictType{DictName: strings.TrimSpace(req.DictName), DictCode: strings.TrimSpace(req.DictCode), Remark: req.Remark, Status: req.Status}
		if err := db.Select("*").Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dict type"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// AdminUpdateDictType 更新字典类型。
// 1.意图 -> 允许后台修改枚举分组名称、编码、备注和启停状态。
// 2.步骤 -> 按 ID 查找记录，绑定 JSON 并保存字段。
// 3.返回 -> 更新后的字典类型记录。
func AdminUpdateDictType(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var item models.SysDictType
		if err := db.First(&item, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "dict type not found"})
			return
		}
		var req SaveDictTypeRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		updates := map[string]interface{}{"dict_name": strings.TrimSpace(req.DictName), "dict_code": strings.TrimSpace(req.DictCode), "remark": req.Remark, "status": req.Status}
		if err := db.Model(&item).Updates(updates).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update dict type"})
			return
		}
		db.First(&item, id)
		c.JSON(http.StatusOK, item)
	}
}

// AdminDeleteDictType 删除字典类型。
// 1.意图 -> 允许后台清理无效枚举分组。
// 2.步骤 -> 按 ID 删除 sys_dict_types 记录。
// 3.返回 -> 删除成功标识。
func AdminDeleteDictType(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := db.Delete(&models.SysDictType{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete dict type"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// AdminListDictData 返回后台字典数据列表。
// 1.意图 -> 支持前后端按 dict_code 查询；左侧类型联动时通过 type_id 定位字典数据。
// 2.步骤 -> dict_type / dict_code / dictType 三参兼容，取第一个非空值；为空时返回空列表（不过滤全部）。
// 3.返回 -> {"code":200,"data":[...],"msg":"ok"}。
func AdminListDictData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		// ── 参数获取：dict_type → dict_code → dictType ──
		dictType := c.Query("dict_type")
		if dictType == "" {
			dictType = c.Query("dict_code")
		}
		if dictType == "" {
			dictType = c.Query("dictType")
		}

		// 初始化查询
		tx := db.Model(&models.SysDictData{})

		// ── 核心过滤：dict_type 不为空时必须加条件 ──
		if dictType != "" {
			tx = tx.Where("dict_code = ?", strings.TrimSpace(dictType))
		}

		// 状态过滤（默认只看启用状态）
		status := strings.TrimSpace(c.Query("status"))
		if status == "" {
			tx = tx.Where("status = ?", 1)
		} else {
			tx = tx.Where("status = ?", status)
		}

		// 排序（sort_order 升序）
		tx = tx.Order("sort_order ASC, id ASC")

		// 分页（默认 pageSize=200，确保工作流字典完整返回）
		pageSize := 200
		if ps := c.Query("pageSize"); ps != "" {
			if v, err := strconv.Atoi(ps); err == nil && v > 0 {
				pageSize = v
			}
		}
		pageNum := 1
		if pn := c.Query("pageNum"); pn != "" {
			if v, err := strconv.Atoi(pn); err == nil && v > 0 {
				pageNum = v
			}
		}
		tx = tx.Limit(pageSize).Offset((pageNum - 1) * pageSize)

		var list []models.SysDictData
		tx.Find(&list)
		c.JSON(http.StatusOK, gin.H{"code": 200, "data": list, "msg": "ok"})
	}
}

// AdminSaveDictData 新增字典数据。
// 1.意图 -> 支撑后台扩展具体枚举项。
// 2.步骤 -> 绑定 JSON、保留显式状态并写入 sys_dict_data。
// 3.返回 -> 新创建的字典数据记录。
func AdminSaveDictData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveDictDataRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.DictCode) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid dict data"})
			return
		}
		item := dictDataFromRequest(req)
		if err := db.Select("*").Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create dict data"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// AdminUpdateDictData 更新字典数据。
// 1.意图 -> 允许后台调整枚举项标签、值、排序和状态。
// 2.步骤 -> 按 ID 查找记录，绑定 JSON 并保存字段。
// 3.返回 -> 更新后的字典数据记录。
func AdminUpdateDictData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var item models.SysDictData
		if err := db.First(&item, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "dict data not found"})
			return
		}
		var req SaveDictDataRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		updates := map[string]interface{}{"dict_code": strings.TrimSpace(req.DictCode), "dict_label": strings.TrimSpace(req.DictLabel), "dict_value": strings.TrimSpace(req.DictValue), "sort_order": req.SortOrder, "status": req.Status, "remark": req.Remark, "meta_json": req.MetaJSON}
		if err := db.Model(&item).Updates(updates).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update dict data"})
			return
		}
		db.First(&item, id)
		c.JSON(http.StatusOK, item)
	}
}

// AdminDeleteDictData 删除字典数据。
// 1.意图 -> 允许后台清理无效枚举项。
// 2.步骤 -> 按 ID 删除 sys_dict_data 记录。
// 3.返回 -> 删除成功标识。
func AdminDeleteDictData(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := db.Delete(&models.SysDictData{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete dict data"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// AdminListArticles 返回后台资讯列表。
// 1.意图 -> 支撑 B 端资讯配置页进行文章运营。
// 2.步骤 -> 支持按 category/status 筛选，并按排序与创建时间输出。
// 3.返回 -> list 格式资讯文章数组。
func AdminListArticles(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		query := db.Model(&models.SysArticle{})
		if category := strings.TrimSpace(c.Query("category")); category != "" {
			query = query.Where("category = ?", category)
		}
		if status := strings.TrimSpace(c.Query("status")); status != "" && status != "all" {
			query = query.Where("status = ?", status)
		}
		var list []models.SysArticle
		query.Order("sort_order asc, created_at desc").Find(&list)
		c.JSON(http.StatusOK, gin.H{"list": list})
	}
}

// AdminSaveArticle 新增资讯文章。
// 1.意图 -> 让后台可发布 C 端首页与资讯 Tab 所需内容。
// 2.步骤 -> 绑定 JSON、补齐默认作者并保留显式状态写入 sys_articles。
// 3.返回 -> 新创建的资讯文章记录。
func AdminSaveArticle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		var req SaveArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil || strings.TrimSpace(req.Title) == "" {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid article"})
			return
		}
		item := articleFromRequest(req)
		if err := db.Select("*").Create(&item).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to create article"})
			return
		}
		c.JSON(http.StatusOK, item)
	}
}

// AdminUpdateArticle 更新资讯文章。
// 1.意图 -> 允许后台编辑标题、封面、摘要、正文、分类和发布状态。
// 2.步骤 -> 按 ID 查找记录，绑定 JSON 并保存字段。
// 3.返回 -> 更新后的资讯文章记录。
func AdminUpdateArticle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		var item models.SysArticle
		if err := db.First(&item, id).Error; err != nil {
			c.JSON(http.StatusNotFound, gin.H{"error": "article not found"})
			return
		}
		var req SaveArticleRequest
		if err := c.ShouldBindJSON(&req); err != nil {
			c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
			return
		}
		updates := map[string]interface{}{"title": strings.TrimSpace(req.Title), "cover_img": strings.TrimSpace(req.CoverImg), "summary": req.Summary, "content": req.Content, "category": strings.TrimSpace(req.Category), "author": strings.TrimSpace(req.Author), "status": req.Status, "sort_order": req.SortOrder, "view_count": req.ViewCount}
		if updates["cover_img"] == "" {
			updates["cover_img"] = "/static/img.png"
		}
		if updates["category"] == "" {
			updates["category"] = "guide"
		}
		if updates["author"] == "" {
			updates["author"] = "Yesok Vietnam"
		}
		if err := db.Model(&item).Updates(updates).Error; err != nil {

			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to update article"})
			return
		}
		db.First(&item, id)
		c.JSON(http.StatusOK, item)
	}
}

// AdminDeleteArticle 删除资讯文章。
// 1.意图 -> 允许后台清理无效资讯内容。
// 2.步骤 -> 按 ID 删除 sys_articles 记录。
// 3.返回 -> 删除成功标识。
func AdminDeleteArticle(db *gorm.DB) gin.HandlerFunc {
	return func(c *gin.Context) {
		id, _ := strconv.ParseUint(c.Param("id"), 10, 64)
		if err := db.Delete(&models.SysArticle{}, id).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to delete article"})
			return
		}
		c.JSON(http.StatusOK, gin.H{"ok": true})
	}
}

// dictDataFromRequest 将字典请求转换为模型。
// 1.意图 -> 统一新增与更新的字段映射，降低表单处理重复代码。
// 2.步骤 -> 裁剪关键字符串并保留显式启停状态。
// 3.返回 -> 可直接 Create 或 Updates 的 SysDictData。
func dictDataFromRequest(req SaveDictDataRequest) models.SysDictData {
	return models.SysDictData{DictCode: strings.TrimSpace(req.DictCode), DictLabel: strings.TrimSpace(req.DictLabel), DictValue: strings.TrimSpace(req.DictValue), SortOrder: req.SortOrder, Status: req.Status, Remark: req.Remark, MetaJSON: req.MetaJSON}
}

// articleFromRequest 将资讯请求转换为模型。
// 1.意图 -> 统一新增与更新的字段映射，保持资讯接口返回结构一致。
// 2.步骤 -> 裁剪标题分类，补齐默认封面和作者并保留显式发布状态。
// 3.返回 -> 可直接 Create 或 Updates 的 SysArticle。
func articleFromRequest(req SaveArticleRequest) models.SysArticle {
	if strings.TrimSpace(req.CoverImg) == "" {
		req.CoverImg = "/static/img.png"
	}
	if strings.TrimSpace(req.Author) == "" {
		req.Author = "Yesok Vietnam"
	}
	if strings.TrimSpace(req.Category) == "" {
		req.Category = "guide"
	}
	return models.SysArticle{Title: strings.TrimSpace(req.Title), CoverImg: strings.TrimSpace(req.CoverImg), Summary: req.Summary, Content: req.Content, Category: strings.TrimSpace(req.Category), Author: strings.TrimSpace(req.Author), Status: req.Status, SortOrder: req.SortOrder, ViewCount: req.ViewCount}
}

// sanitizeUploadName 清洗上传文件基础名。
// 1.意图 -> 避免文件名包含路径、空格或特殊字符导致保存异常。
// 2.步骤 -> 仅保留数字、字母、下划线、连字符和中文字符，其他字符替换为下划线。
// 3.返回 -> 安全的文件名基础字符串。
func sanitizeUploadName(name string) string {
	var builder strings.Builder
	for _, r := range name {
		if (r >= '0' && r <= '9') || (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || r == '_' || r == '-' || (r >= '\u4e00' && r <= '\u9fa5') {
			builder.WriteRune(r)
		} else {
			builder.WriteRune('_')
		}
	}
	return strings.Trim(builder.String(), "_")
}
