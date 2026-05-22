package main

import (
	"flag"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/gin-gonic/gin"
	"gorm.io/driver/mysql"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"

	"yesok-vietnam/server/handlers"
	"yesok-vietnam/server/middleware"
	"yesok-vietnam/server/models"
)

// main 是服务端启动入口，负责数据库连接、核心表迁移、种子数据、路由注册和 H5 静态资源托管。
func main() {
	// 1. 定义命令行参数
	runMigrate := flag.Bool("migrate", false, "是否在启动时执行数据库迁移和种子数据注入")
	flag.Parse()

	db := connectDatabase()

	// 2. 根据参数决定是否执行耗时的迁移操作
	if *runMigrate {
		log.Println("检测到 -migrate 参数，开始执行数据库表结构迁移与种子数据检查...")
		migrateCoreTables(db)
		seedCoreData(db)
		log.Println("迁移与种子数据注入完成！")
	} else {
		log.Println("跳过数据库表结构迁移与种子数据注入 (如需执行请加上 -migrate 参数)")
	}

	authMw := middleware.NewAuthMiddleware()
	gin.SetMode(gin.ReleaseMode)
	r := gin.New()
	r.Use(gin.Logger(), gin.Recovery())
	r.Use(middleware.CORS())

	registerHealthRoute(r)
	registerAPIRoutes(r, db, authMw)
	registerStaticRoutes(r)

	addr := ":" + getEnv("SERVER_PORT", "7625")
	log.Printf("YesokVietnam starting on %s", addr)
	if err := r.Run(addr); err != nil {
		log.Fatalf("server failed: %v", err)
	}
}

// connectDatabase 创建数据库连接。强制定死连接远程 MySQL。
func connectDatabase() *gorm.DB {
	dbType := "mysql"

	username := getEnv("DB_USER", "root")
	password := getEnv("DB_PASS", "fangchenye520")
	dbName := getEnv("DB_NAME", "yesok_vn")
	dbHost := getEnv("DB_HOST", "127.0.0.1")
	dbPort := getEnv("DB_PORT", "3306")

	dsn := fmt.Sprintf("%s:%s@tcp(%s:%s)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, dbHost, dbPort, dbName)

	var db *gorm.DB
	var err error

	log.Printf("正在连接远程 %s 数据库...", dbType)

	db, err = gorm.Open(mysql.Open(dsn), &gorm.Config{
		Logger: logger.Default.LogMode(logger.Info),
	})

	if err != nil {
		log.Fatalf("failed to connect to database: %v", err)
	}

	return db
}

// migrateCoreTables 迁移核心业务表。
func migrateCoreTables(db *gorm.DB) {
	if err := db.AutoMigrate(&models.AppUser{}, &models.SysUser{}, &models.SysService{}, &models.SysWorkflowNode{}, &models.Order{}, &models.OrderTimeline{}, &models.PaymentRecord{}, &models.SysConfig{}, &models.SysDictType{}, &models.SysDictData{}, &models.SysArticle{}); err != nil {
		log.Fatalf("failed to auto-migrate core tables: %v", err)
	}
}

// registerHealthRoute 注册健康检查接口。
func registerHealthRoute(r *gin.Engine) {
	r.GET("/health", func(c *gin.Context) { c.JSON(http.StatusOK, gin.H{"status": "ok"}) })
}

// registerAPIRoutes 注册 API 路由。
// 彻底拆解父子嵌套关系：公开组与鉴权组完全并列，互不继承，终结 401 死锁。
func registerAPIRoutes(r *gin.Engine, db *gorm.DB, authMw *middleware.AuthMiddleware) {
	// ==========================================================
	// 1. 全裸独立公开组 (免 Token，畅通小程序冷启动与健康检查)
	// ==========================================================
	apiV1 := r.Group("/api/v1")
	{
		// 系统健康检查 /api/v1/state
		apiV1.GET("/state", handlers.GetState(db))
		apiV1.GET("/configs", handlers.ClientGetConfigs(db))

		// 客户端免登录公共数据
		apiV1.GET("/client/services", handlers.ClientListServices(db))
		apiV1.GET("/client/articles", handlers.ClientListArticles(db))

		// 凭证换取登录接口 (公开)
		apiV1.POST("/auth/login", handlers.AuthAdmin(db))  // 后台管家登录
		apiV1.POST("/client/auth/tg", handlers.AuthTG(db)) // Telegram Mini App 登录
	}

	// ==========================================================
	// 2. 高级鉴权受保护组 (必须带 Token 才可访问)
	// ==========================================================
	authGroup := r.Group("/api/v1")
	authGroup.Use(authMw.RequireAuth())
	{
		// B 端管理后台路由
		authGroup.GET("/admin/auth/me", handlers.AdminMe(db))
		authGroup.GET("/admin/dashboard/stats", handlers.DashboardStats(db))
		authGroup.GET("/admin/orders", handlers.AdminListOrders(db))
		authGroup.GET("/admin/orders/:id", handlers.AdminGetOrder(db))
		authGroup.PUT("/admin/orders/:id", handlers.AdminUpdateOrder(db))
		authGroup.GET("/admin/services", handlers.AdminListServices(db))
		authGroup.POST("/admin/services", handlers.AdminSaveService(db))
		authGroup.PUT("/admin/services/:id", handlers.AdminUpdateService(db))
		authGroup.POST("/admin/upload", handlers.UploadFile(db))
		authGroup.GET("/admin/dict-types", handlers.AdminListDictTypes(db))
		authGroup.POST("/admin/dict-types", handlers.AdminSaveDictType(db))
		authGroup.PUT("/admin/dict-types/:id", handlers.AdminUpdateDictType(db))
		authGroup.DELETE("/admin/dict-types/:id", handlers.AdminDeleteDictType(db))
		authGroup.GET("/admin/dict-data", handlers.AdminListDictData(db))
		authGroup.POST("/admin/dict-data", handlers.AdminSaveDictData(db))
		authGroup.PUT("/admin/dict-data/:id", handlers.AdminUpdateDictData(db))
		authGroup.DELETE("/admin/dict-data/:id", handlers.AdminDeleteDictData(db))
		authGroup.GET("/admin/articles", handlers.AdminListArticles(db))
		authGroup.POST("/admin/articles", handlers.AdminSaveArticle(db))
		authGroup.PUT("/admin/articles/:id", handlers.AdminUpdateArticle(db))
		authGroup.DELETE("/admin/articles/:id", handlers.AdminDeleteArticle(db))
		authGroup.GET("/admin/payments", handlers.AdminListPayments(db))
		authGroup.GET("/admin/app-users", handlers.AdminListAppUsers(db))
		authGroup.GET("/admin/sys-users", handlers.AdminListSysUsers(db))
		authGroup.POST("/admin/sys-users", handlers.AdminCreateSysUser(db))
		authGroup.PUT("/admin/sys-users/:id", handlers.AdminUpdateSysUser(db))
		authGroup.DELETE("/admin/sys-users/:id", handlers.AdminDeleteSysUser(db))
		authGroup.GET("/admin/users", handlers.ListUsers(db))
		authGroup.PUT("/admin/users/:id/role", handlers.UpdateUserRole(db))
		authGroup.DELETE("/admin/users/:id", handlers.DeleteUser(db))

		// C 端用户私有路由
		authGroup.GET("/client/user/me", handlers.GetMe(db))
		authGroup.GET("/client/state", handlers.GetState(db))
		authGroup.PUT("/client/state", handlers.UpdateState(db))
		authGroup.POST("/client/auth/logout", handlers.AuthLogout())
	}
}

// registerStaticRoutes 注册前端静态资源和上传资源托管。
// 路由优先级：/uploads > /admin/* > /client/* > NoRoute(SPA fallback)
func registerStaticRoutes(r *gin.Engine) {
	isDev := getEnv("ENV", "prod") == "dev"

	var adminDir, clientDir string
	if isDev {
		adminDir = "../web-admin/dist"
		clientDir = "../web-client/dist"
	} else {
		adminDir = "./dist/admin"
		clientDir = "./dist/client"
	}

	// 1. 最高优先级：上传目录托管
	r.Static("/uploads", "./uploads")

	// 2. 挂载 C 端（用户端 Mini App）的静态产物根目录
	r.Static("/client", clientDir)

	// 3. 挂载 B 端（管理后台）的静态产物根目录
	r.Static("/admin", adminDir)

	// 4. 根路径别名防呆（优先指向 C 端）
	r.Static("/assets", filepath.Join(clientDir, "assets"))
	r.Static("/static", filepath.Join(clientDir, "static"))

	// 5. 单页应用 (SPA) 路由兜底转发中心
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		if strings.HasPrefix(path, "/api") {
			c.JSON(http.StatusNotFound, gin.H{"code": 404, "message": "API route not found"})
			return
		}

		if strings.HasPrefix(path, "/admin") {
			if strings.Contains(path, "/assets/") {
				filename := path[strings.Index(path, "/assets/")+8:]
				c.File(filepath.Join(adminDir, "assets", filename))
				return
			}
			c.File(filepath.Join(adminDir, "index.html"))
			return
		}

		if strings.Contains(path, "/assets/") {
			filename := path[strings.Index(path, "/assets/")+8:]
			c.File(filepath.Join(clientDir, "assets", filename))
			return
		}

		c.File(filepath.Join(clientDir, "index.html"))
	})
}

// getEnv 返回环境变量值，若为空则返回默认值。
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
