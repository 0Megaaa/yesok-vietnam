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

	// ⚠️ 注意：如果以下四个包在本地仍报错 "missing dot in first path element"
	// 说明你 server 目录下的 go.mod 里的 module 名字不是 yesok-vietnam/server
	// 如果 go.mod 里写的是 module server，请把下面四行的前缀改为 "server/xxx"
	"yesok-vietnam/server/config"
	"yesok-vietnam/server/handlers"
	"yesok-vietnam/server/middleware"
	"yesok-vietnam/server/models"
)

// main 是服务端启动入口，负责数据库连接、核心表迁移、种子数据、路由注册和 H5 静态资源托管。
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

// connectDatabase 创建数据库连接。强制定死连接远程 MySQL。
func connectDatabase() *gorm.DB {
	dbType := "mysql"

	username := "root"          // 数据库账号
	password := "YOUR_PASSWORD" // ⚠️ 请在替换后手动把这里改成真实的数据库密码！
	dbName := "yesok"           // 数据库名称

	// 自动拼接成 Go 标准的 MySQL DSN 格式 (指向 72.61.213.87:3610)
	dsn := fmt.Sprintf("%s:%s@tcp(72.61.213.87:3610)/%s?charset=utf8mb4&parseTime=True&loc=Local", username, password, dbName)

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

// registerStaticRoutes 注册 H5 静态资源和上传资源托管。
func registerStaticRoutes(r *gin.Engine) {
	staticDir := config.Global.Server.StaticDir
	r.Static("/uploads", "./uploads")
	r.StaticFS("/assets", http.Dir(filepath.Join(staticDir, "assets")))
	r.StaticFS("/static", http.Dir(filepath.Join(staticDir, "static")))
	r.NoRoute(func(c *gin.Context) { c.File(filepath.Join(staticDir, "index.html")) })
}
