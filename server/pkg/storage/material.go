package storage

import (
	"crypto/rand"
	"encoding/hex"
	"os"
	"path/filepath"
	"regexp"
	"strings"
	"time"
)

const maxUploadSize = 10 * 1024 * 1024 // 10MB

var allowedExts = map[string]string{
	".jpg":  "image/jpeg",
	".jpeg": "image/jpeg",
	".png":  "image/png",
}

var allowedMimes = map[string]bool{
	"image/jpeg": true,
	"image/png":  true,
}

// MaterialStorageRoot 返回资料文件存储根目录。
// 生产：/www/wwwroot/yesok-vietnam/server/material
// 本地：{项目根目录}/uploads
// 支持 FILE_STORAGE_ROOT 环境变量覆盖。
func MaterialStorageRoot() string {
	if root := os.Getenv("FILE_STORAGE_ROOT"); root != "" {
		return root
	}
	if env := strings.ToLower(strings.TrimSpace(os.Getenv("ENV"))); env == "prod" {
		return "/www/wwwroot/yesok-vietnam/server/material"
	}
	// 本地：server/ 上级目录，即项目根目录
	execPath, _ := os.Executable()
	execDir := filepath.Dir(execPath)
	// 如果在 server/ 目录下运行，execDir 就是 server/
	// parent = 项目根目录
	parent := filepath.Dir(execDir)
	return filepath.Join(parent, "uploads")
}

// BuildMaterialFilename 生成资料文件安全文件名。
// 格式：{order_no}_{yyyyMMdd}_{safe_name}_{random}.{ext}
// fieldKey 优先用于安全名，否则用原始文件名中的英文数字部分。
func BuildMaterialFilename(orderNo string, fieldKey string, originalName string) string {
	ext := strings.ToLower(filepath.Ext(originalName))
	if ext == "" {
		ext = ".png"
	}

	// 优先用 fieldKey 清理后的值
	safeName := cleanFileName(fieldKey)
	if safeName == "" {
		safeName = cleanFileName(originalName)
	}
	if safeName == "" {
		safeName = "file"
	}

	date := time.Now().Format("20060102")
	random := randomString(6)
	orderNo = strings.TrimSpace(orderNo)
	if orderNo == "" {
		orderNo = "ORDER"
	}

	return orderNo + "_" + date + "_" + safeName + "_" + random + ext
}

// IsAllowedFileExt 检查扩展名是否允许。
func IsAllowedFileExt(ext string) bool {
	_, ok := allowedExts[strings.ToLower(ext)]
	return ok
}

// IsAllowedMimeType 检查 MIME 类型是否允许。
func IsAllowedMimeType(mime string) bool {
	return allowedMimes[mime]
}

// AllowedMimes 返回允许的 MIME 类型列表（用于前端提示）。
func AllowedMimes() []string {
	return []string{"image/jpeg", "image/png"}
}

// AllowedExts 返回允许的扩展名列表。
func AllowedExts() []string {
	return []string{".jpg", ".jpeg", ".png"}
}

// MaxUploadSize 返回单文件最大字节数。
func MaxUploadSize() int64 {
	return maxUploadSize
}

// cleanFileName 只保留英文、数字、下划线、横线，其余替换为横线后合并。
func cleanFileName(name string) string {
	// 去掉路径（Windows 反斜杠或 Unix 斜杠）
	name = strings.ReplaceAll(name, "\\", "_")
	name = strings.ReplaceAll(name, "/", "_")

	// 只保留英文、数字、下划线、横线
	reg := regexp.MustCompile(`[^a-zA-Z0-9_\-]`)
	name = reg.ReplaceAllString(name, "_")

	// 合并连续下划线
	for strings.Contains(name, "__") {
		name = strings.ReplaceAll(name, "__", "_")
	}
	name = strings.Trim(name, "_")
	return strings.ToLower(name)
}

// randomString 生成指定长度的随机十六进制字符串。
func randomString(length int) string {
	bytes := make([]byte, (length+1)/2)
	rand.Read(bytes)
	return hex.EncodeToString(bytes)[:length]
}

// EnsureDir 创建目录（如果不存在）。
func EnsureDir(dir string) error {
	return os.MkdirAll(dir, 0755)
}
