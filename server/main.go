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

	"yesok-vietnam/server/config"
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

	addr := ":" + config.Global.Server.Port
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
func registerAPIRoutes(r *gin.Engine, db *gorm.DB, authMw *middleware.AuthMiddleware) {
	v1 := r.Group("/api/v1")
	{
		v1.GET("/services", handlers.ClientListServices(db))
		v1.POST("/orders", handlers.ClientCreateOrder(db))
		v1.GET("/configs", handlers.ClientGetConfigs(db))
		v1.GET("/articles", handlers.ClientListArticles(db))

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
				adminProtected.POST("/upload", handlers.UploadFile(db))
				adminProtected.GET("/dict-types", handlers.AdminListDictTypes(db))
				adminProtected.POST("/dict-types", handlers.AdminSaveDictType(db))
				adminProtected.PUT("/dict-types/:id", handlers.AdminUpdateDictType(db))
				adminProtected.DELETE("/dict-types/:id", handlers.AdminDeleteDictType(db))
				adminProtected.GET("/dict-data", handlers.AdminListDictData(db))
				adminProtected.POST("/dict-data", handlers.AdminSaveDictData(db))
				adminProtected.PUT("/dict-data/:id", handlers.AdminUpdateDictData(db))
				adminProtected.DELETE("/dict-data/:id", handlers.AdminDeleteDictData(db))
				adminProtected.GET("/articles", handlers.AdminListArticles(db))
				adminProtected.POST("/articles", handlers.AdminSaveArticle(db))
				adminProtected.PUT("/articles/:id", handlers.AdminUpdateArticle(db))
				adminProtected.DELETE("/articles/:id", handlers.AdminDeleteArticle(db))
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

// registerStaticRoutes 注册前端静态资源和上传资源托管。
// 路由优先级：/uploads > /admin/* > /client/* > NoRoute(SPA fallback)
func registerStaticRoutes(r *gin.Engine) {
	isDev := getEnv("ENV", "prod") == "dev"

	var adminDir, clientDir string
	if isDev {
		// 开发环境：指向各自源码 dist 目录
		adminDir = "../web-admin/dist"
		clientDir = "../web-client/dist"
	} else {
		// 生产环境：统一放在 server/dist/ 下
		adminDir = "./dist/admin"
		clientDir = "./dist/client"
	}

	// 1. 上传目录（最高优先级）
	r.Static("/uploads", "./uploads")

	// 2. 管理后台静态资源（/admin/* → dist/admin/）
	r.Static("/admin", adminDir)

	// 3. 用户端静态资源（/client/* → dist/client/）
	r.Static("/client", clientDir)

	// 4. 资源路径别名（修正 Cursor 错误：r.Static 的第二个参数必须是普通的 string 路径）
	r.Static("/assets", filepath.Join(clientDir, "assets"))
	r.Static("/static", filepath.Join(clientDir, "static"))

	// 5. 兜底处理 (SPA 路由 fallback)：解决刷新页面 404 问题
	r.NoRoute(func(c *gin.Context) {
		path := c.Request.URL.Path

		// 如果是请求后台管理端路由，返回 admin 的 index.html
		if strings.HasPrefix(path, "/admin") {
			c.File(filepath.Join(adminDir, "index.html"))
			return
		}

		// 其余所有非 API 请求，全部返回用户端的 index.html
		if !strings.HasPrefix(path, "/api") {
			c.File(filepath.Join(clientDir, "index.html"))
			return
		}
	})
}

// getEnv 返回环境变量值，若为空则返回默认值。
func getEnv(key, fallback string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return fallback
}
