package config

import (
	"os"
	"time"
)

var Global = struct {
	Server   ServerConfig
	Database DatabaseConfig
	TG       TGConfig
}{
	Server: ServerConfig{
		Port:         getEnv("SERVER_PORT", "7625"),
		StaticDir:    getEnv("STATIC_DIR", "../web/dist"),
		HTMLDir:      getEnv("HTML_DIR", "../web/dist"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	},
	Database: DatabaseConfig{
		Host:     getEnv("DB_HOST", "127.0.0.1"),
		Port:     getEnv("DB_PORT", "3306"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", ""),
		Name:     getEnv("DB_NAME", "yesok_vn"),
	},
	TG: TGConfig{
		BotToken: getEnv("TG_BOT_TOKEN", ""),
	},
}

type ServerConfig struct {
	Port         string
	StaticDir    string
	HTMLDir      string
	ReadTimeout  time.Duration
	WriteTimeout time.Duration
}

type DatabaseConfig struct {
	Host     string
	Port     string
	User     string
	Password string
	Name     string
}

type TGConfig struct {
	BotToken string
}

// getEnv 读取环境变量并提供安全默认值。
// 实现步骤：
// 1. 优先读取宝塔面板或系统服务注入的环境变量。
// 2. 如果变量为空，则返回仅适合本地开发的默认值。
// 3. 禁止在代码中写入生产数据库密码、机器人 Token 等敏感信息。
func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
