package main

import (
	"blog-server/config"
	"blog-server/models"
	"blog-server/routes"
	"blog-server/utils"
	"log"

	"github.com/gin-gonic/gin"
)

func main() {
	// 加载配置
	config.LoadConfig()

	// 初始化数据库
	models.InitDB()

	// 自动迁移
	models.Migrate()

	// 初始化存储服务
	if err := utils.InitStorage(); err != nil {
		log.Printf("存储服务初始化失败: %v", err)
		log.Println("图片上传功能将不可用")
	}

	// 初始化Redis
	if err := utils.InitRedis(); err != nil {
		log.Printf("Redis初始化失败: %v", err)
		log.Println("数据分析功能将不可用")
	} else {
		// 启动定时任务
		utils.StartScheduler()
		log.Println("数据转存定时任务已启动")
	}

	// 设置Gin模式
	if config.AppConfig.Mode == "release" {
		gin.SetMode(gin.ReleaseMode)
	}

	// 创建路由
	r := gin.Default()

	// 设置路由
	routes.SetupRoutes(r)

	// 启动服务器
	port := config.AppConfig.Port
	if port == "" {
		port = "8080"
	}

	log.Printf("服务器启动在端口: %s", port)
	log.Fatal(r.Run(":" + port))
}