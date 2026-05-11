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
	// Extract embedded frontend into an fs.FS sub-tree for SPA serving.
	frontendFS, err := fs.Sub(frontendStatic, "web/dist")
	if err != nil {
		log.Fatalf("failed to extract embedded frontend: %v", err)
	}

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

	// Middleware instances
	authMw := middleware.NewAuthMiddleware()

	// Router
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	// Health check
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{"status": "ok"})
	})

	// ─── API v1 ──────────────────────────────────────────────────────────────
	v1 := r.Group("/api/v1")
	{
		// ── Client routes (TG / future WeChat / APP) ─────────────────────────
		client := v1.Group("/client")
		{
			// Public — no auth required
			client.POST("/auth/tg", handlers.AuthTG(db))
			// client.POST("/auth/wechat", ...) ← placeholder, uncomment when WeChat is ready

			// Protected — requires valid JWT
			clientProtected := client.Group("")
			clientProtected.Use(authMw.RequireAuth())
			{
				clientProtected.GET("/user/me", handlers.GetMe(db))
				clientProtected.GET("/state", handlers.GetState(db))
				clientProtected.PUT("/state", handlers.UpdateState(db))
			}
		}

		// ── Admin routes (PC browser, username + password) ──────────────────
		admin := v1.Group("/admin")
		{
			// Public — no auth required
			admin.POST("/auth/login", handlers.AuthAdmin(db))

			// Protected — requires valid JWT + admin role
			adminProtected := admin.Group("")
			adminProtected.Use(authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin))
			{
				adminProtected.POST("/auth/logout", handlers.AuthLogout())
				adminProtected.GET("/auth/me", handlers.AdminMe(db))

				// User management
				adminProtected.GET("/users", handlers.ListUsers(db))
				adminProtected.PUT("/users/:id/role", handlers.UpdateUserRole(db))
				adminProtected.DELETE("/users/:id", handlers.DeleteUser(db))

				// Dashboard
				adminProtected.GET("/dashboard/stats", handlers.DashboardStats(db))

				// Order management
				adminProtected.GET("/orders", handlers.AdminListOrders(db))
				adminProtected.PUT("/orders/:id", handlers.AdminUpdateOrder(db))
			}
		}
	}

	// ─── SPA fallback ───────────────────────────────────────────────────────
	// Any non-API route returns index.html from the embedded dist.
	// The Vue Router history mode handles client-side routing.
	staticFS := http.FS(frontendFS)
	r.NoRoute(func(c *gin.Context) {
		c.Header("Cache-Control", "no-cache")
		c.FileFromFS(c.Request.URL.Path, staticFS)
	})

	addr := ":" + config.Global.Server.Port
	log.Printf("YesokVietnam starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}
