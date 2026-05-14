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
		StaticDir:    getEnv("STATIC_DIR", "./static"),
		HTMLDir:      getEnv("HTML_DIR", "./html"),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	},
	Database: DatabaseConfig{
		Host:     getEnv("DB_HOST", "72.61.213.87"),
		Port:     getEnv("DB_PORT", "3610"),
		User:     getEnv("DB_USER", "root"),
		Password: getEnv("DB_PASSWORD", "ijh6789$"),
		Name:     getEnv("DB_NAME", "yesok_vn"),
	},
	TG: TGConfig{
		BotToken: getEnv("TG_BOT_TOKEN", "8740938291:AAHu1z_yUt8jEaXdd2D4O-AR72sUBH34Tp8"),
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

func getEnv(key, fallback string) string {
	if v := os.Getenv(key); v != "" {
		return v
	}
	return fallback
}
