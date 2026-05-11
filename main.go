package main

import (
	"embed"
	"fmt"
	"io/fs"
	"log"
	"net/http"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"yesok-vietnam/config"
	"yesok-vietnam/handlers"
	"yesok-vietnam/middleware"
	"yesok-vietnam/models"
)

//go:embed web/dist/*
var frontendStatic embed.FS

func main() {
	// Extract embedded frontend into an http.FileSystem
	frontendFS, err := fs.Sub(frontendStatic, "web/dist")
	if err != nil {
		log.Fatalf("failed to extract embedded frontend: %v", err)
	}
	staticServer := http.FileServer(http.FS(frontendFS))

	// Database connection
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Global.Database.User,
		config.Global.Database.Password,
		config.Global.Database.Host,
		config.Global.Database.Port,
		config.Global.Database.Name,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	// Auto-migrate models
	if err := db.AutoMigrate(&models.User{}, &models.Order{}); err != nil {
		log.Fatalf("failed to auto-migrate: %v", err)
	}

	// Dependency injection
	authMw := middleware.NewAuthMiddleware(db)

	// Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.Default()
	r.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "GET, POST, OPTIONS, PUT, DELETE")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Authorization, X-Requested-With")
		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}
		c.Next()
	})

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// API routes
	api := r.Group("/api")
	{
		api.POST("/auth/tg", handlers.AuthTG(db))

		protected := api.Group("")
		protected.Use(authMw.RequireAuth())
		{
			protected.GET("/user/me", handlers.GetMe(db))
			protected.GET("/state", handlers.GetState(db))
			protected.PUT("/state", handlers.UpdateState(db))
		}
	}

	// SPA fallback — any non-API route returns index.html
	r.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS(c.Request.URL.Path, staticServer)
	})

	addr := ":" + config.Global.Server.Port
	log.Printf("YesokVietnam starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
