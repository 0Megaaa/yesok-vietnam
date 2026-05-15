package main

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"yesok-vietnam/server/config"
	"yesok-vietnam/server/handlers"
	"yesok-vietnam/server/middleware"
	"yesok-vietnam/server/models"
)

// main 是服务端启动入口，负责数据库连接、8 表迁移、种子数据、路由注册和 H5 静态资源托管。
// 1.意图 -> 启动 Yesok 2.0 全栈真实数据链路。
// 2.步骤 -> 连接数据库、迁移 8 张核心表、注入种子数据、挂载 C/B 端 API 与静态资源。
// 3.返回 -> 无返回；服务启动失败时记录 fatal 日志。
func main() {
	db := connectDatabase()
	migrateCoreTables(db)
	seedCoreData(db)

	authMw := middleware.NewAuthMiddleware()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	registerHealthRoute(r)
	registerAPIRoutes(r, db, authMw)
	registerStaticRoutes(r)

	addr := ":" + config.Global.Server.Port
	log.Printf("YesokVietnam starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// connectDatabase 创建数据库连接。
// 1.意图 -> 同时支持生产 MySQL 与沙盒 SQLite 联调，保护线上部署方式。
// 2.步骤 -> DB_DRIVER=sqlite 时打开本地文件，否则按环境变量拼接 MySQL DSN。
// 3.返回 -> 已建立的 GORM 数据库连接。
func connectDatabase() *gorm.DB {
	if os.Getenv("DB_DRIVER") == "sqlite" {
		path := os.Getenv("DB_SQLITE_PATH")
		if path == "" {
			path = "yesok_vn.db"
		}
		db, err := gorm.Open(sqlite.Open(path), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
		if err != nil {
			log.Fatalf("failed to connect sqlite database: %v", err)
		}
		return db
	}
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", config.Global.Database.User, config.Global.Database.Password, config.Global.Database.Host, config.Global.Database.Port, config.Global.Database.Name)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

// migrateCoreTables 迁移 8 张核心业务表。
// 1.意图 -> 建立 app_users、sys_users、sys_services、sys_workflow_nodes、orders、order_timelines、payment_records、sys_configs。
// 2.步骤 -> 使用 GORM AutoMigrate 幂等建表，字段 comment 与模型保持一致。
// 3.返回 -> 无返回；失败时终止服务。
func migrateCoreTables(db *gorm.DB) {
	if err := db.AutoMigrate(&models.AppUser{}, &models.SysUser{}, &models.SysService{}, &models.SysWorkflowNode{}, &models.Order{}, &models.OrderTimeline{}, &models.PaymentRecord{}, &models.SysConfig{}); err != nil {
		log.Fatalf("failed to auto-migrate core tables: %v", err)
	}
}

// registerHealthRoute 注册健康检查接口。
// 1.意图 -> 给部署平台和联调脚本提供稳定探针。
// 2.步骤 -> 暴露 /health 并返回固定 JSON。
// 3.返回 -> HTTP 200 状态。
func registerHealthRoute(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}

// registerAPIRoutes 注册 API 路由。
// 1.意图 -> 同时服务 C 端公开业务链路与 B 端管家后台。
// 2.步骤 -> 注册 /api/v1/services、/orders、/configs 与 /api/v1/admin 全业务矩阵接口。
// 3.返回 -> 无返回，路由写入 Gin 引擎。
func registerAPIRoutes(r *gin.Engine, db *gorm.DB, authMw *middleware.AuthMiddleware) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/services", handlers.ClientListServices(db))
		v1.POST("/orders", handlers.ClientCreateOrder(db))
		v1.GET("/configs", handlers.ClientGetConfigs(db))

		client := v1.Group("/client")
		{
			client.POST("/auth/tg", handlers.AuthTG(db))
			clientProtected := client.Group("")
			clientProtected.Use(authMw.RequireAuth())
			{
				clientProtected.GET("/user/me", handlers.GetMe(db))
				clientProtected.GET("/state", handlers.GetState(db))
				clientProtected.PUT("/state", handlers.UpdateState(db))
			}
		}

		admin := v1.Group("/admin")
		{
			admin.POST("/auth/login", handlers.AuthAdmin(db))
			adminProtected := admin.Group("")
			adminProtected.Use(authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin, models.RoleManager))
			{
				adminProtected.POST("/auth/logout", handlers.AuthLogout())
				adminProtected.GET("/auth/me", handlers.AdminMe(db))
				adminProtected.GET("/dashboard/stats", handlers.DashboardStats(db))
				adminProtected.GET("/orders", handlers.AdminListOrders(db))
				adminProtected.GET("/orders/:id", handlers.AdminGetOrder(db))
				adminProtected.PUT("/orders/:id", handlers.AdminUpdateOrder(db))
				adminProtected.GET("/services", handlers.AdminListServices(db))
				adminProtected.POST("/services", handlers.AdminSaveService(db))
				adminProtected.PUT("/services/:id", handlers.AdminUpdateService(db))
				adminProtected.GET("/payments", handlers.AdminListPayments(db))
				adminProtected.GET("/app-users", handlers.AdminListAppUsers(db))
				adminProtected.GET("/sys-users", handlers.AdminListSysUsers(db))
				adminProtected.POST("/sys-users", handlers.AdminCreateSysUser(db))
				adminProtected.PUT("/sys-users/:id", handlers.AdminUpdateSysUser(db))
				adminProtected.DELETE("/sys-users/:id", handlers.AdminDeleteSysUser(db))
				adminProtected.GET("/users", handlers.ListUsers(db))
				adminProtected.PUT("/users/:id/role", handlers.UpdateUserRole(db))
				adminProtected.DELETE("/users/:id", handlers.DeleteUser(db))
			}
		}
	}
}

// registerStaticRoutes 注册 H5 静态资源托管。
// 1.意图 -> 让 Go 服务可直接托管前端 dist。
// 2.步骤 -> 托管 assets 与 static 目录，非 API 路由回退 index.html。
// 3.返回 -> 无返回。
func registerStaticRoutes(r *gin.Engine) {
	staticDir := config.Global.Server.StaticDir
	r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
	r.StaticFS("/static", http.Dir(filepath.Join(staticDir, "static")))
	r.NoRoute(func(c *gin.Context) { c.File(filepath.Join(staticDir, "index.html")) })
}
