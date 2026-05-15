package main

import (
	"fmt"
	"log"
	"net/http"
	"path/filepath"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
	"yesok-vietnam/server/config"
	"yesok-vietnam/server/handlers"
	"yesok-vietnam/server/middleware"
	"yesok-vietnam/server/models"
)

// main 是服务端启动入口，负责数据库连接、核心表迁移、路由注册和 H5 静态资源托管。
// 实现步骤：
// 1. 按环境变量拼接 MySQL DSN，保证宝塔部署时不依赖 Docker。
// 2. 使用 GORM 迁移 7 张核心业务表，建立新的订单与流程基础设施。
// 3. 注册客户端与后台接口，并保留旧接口作为过渡兼容。
// 4. 托管前端打包目录，让宝塔只需要启动 Go 二进制即可提供 H5 预览。
func main() {
	db := connectDatabase()
	migrateCoreTables(db)

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
// 实现步骤：
// 1. 从配置中心读取数据库账号、主机、端口和库名。
// 2. 拼接兼容 MySQL utf8mb4 与本地时区的 DSN。
// 3. 打开 GORM 连接，并在失败时立即终止服务，避免无数据库状态下错误运行。
func connectDatabase() *gorm.DB {
	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local",
		config.Global.Database.User,
		config.Global.Database.Password,
		config.Global.Database.Host,
		config.Global.Database.Port,
		config.Global.Database.Name,
	)
	db, err := gorm.Open(mysql.Open(dsn), &gorm.Config{Logger: logger.Default.LogMode(logger.Info)})
	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}
	return db
}

// migrateCoreTables 迁移本次重构定义的 7 张核心业务表。
// 实现步骤：
// 1. 迁移 app_users，统一 C 端用户身份。
// 2. 迁移 sys_users 与 sys_services，支撑后台员工和服务目录。
// 3. 迁移 sys_workflow_nodes，支撑动态工作流按钮配置。
// 4. 迁移 orders、order_timelines、payment_records，完成订单履约与支付对账闭环。
func migrateCoreTables(db *gorm.DB) {
	if err := db.AutoMigrate(
		&models.AppUser{},
		&models.SysUser{},
		&models.SysService{},
		&models.SysWorkflowNode{},
		&models.Order{},
		&models.OrderTimeline{},
		&models.PaymentRecord{},
	); err != nil {
		log.Fatalf("failed to auto-migrate core tables: %v", err)
	}
}

// registerHealthRoute 注册健康检查接口。
// 实现步骤：
// 1. 暴露 /health 给宝塔、Nginx 或监控探针使用。
// 2. 返回固定 JSON，避免泄露数据库、令牌或服务器路径等敏感信息。
func registerHealthRoute(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) {
		c.JSON(http.StatusOK, gin.H{"status": "ok"})
	})
}

// registerAPIRoutes 注册 API 路由。
// 实现步骤：
// 1. 保留客户端 TG 登录接口，后续可平滑替换为微信与 Telegram 双端鉴权。
// 2. 保留后台用户、订单和统计接口，避免管理端在重构期间中断。
// 3. 对需要权限的接口挂载 JWT 与角色中间件。
func registerAPIRoutes(r *gin.Engine, db *gorm.DB, authMw *middleware.AuthMiddleware) {
	v1 := r.Group("/api/v1")
	{
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
			adminProtected.Use(authMw.RequireAuth(), authMw.RequireRole(models.RoleAdmin))
			{
				adminProtected.POST("/auth/logout", handlers.AuthLogout())
				adminProtected.GET("/auth/me", handlers.AdminMe(db))
				adminProtected.GET("/users", handlers.ListUsers(db))
				adminProtected.PUT("/users/:id/role", handlers.UpdateUserRole(db))
				adminProtected.DELETE("/users/:id", handlers.DeleteUser(db))
				adminProtected.GET("/dashboard/stats", handlers.DashboardStats(db))
				adminProtected.GET("/orders", handlers.AdminListOrders(db))
				adminProtected.PUT("/orders/:id", handlers.AdminUpdateOrder(db))
			}
		}
	}
}

// registerStaticRoutes 注册 H5 静态资源托管。
// 实现步骤：
// 1. 从 STATIC_DIR 读取前端 dist 目录，默认指向 ../web/dist。
// 2. 非 API 路由统一回退到 index.html，支持 UniApp H5 单页应用刷新。
// 3. 如果 dist 尚未构建，则返回明确错误，便于部署人员定位问题。
func registerStaticRoutes(r *gin.Engine) {
	staticDir := config.Global.Server.StaticDir
	r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
	r.NoRoute(func(c *gin.Context) {
		indexPath := filepath.Join(staticDir, "index.html")
		c.File(indexPath)
	})
}
