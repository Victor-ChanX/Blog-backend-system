package routes

import (
	"blog-server/controllers"
	"blog-server/middleware"

	"github.com/gin-gonic/gin"
)

func SetupRoutes(r *gin.Engine) {
	// 添加日志中间件
	r.Use(middleware.LoggerMiddleware())
	
	// 添加CORS中间件
	r.Use(func(c *gin.Context) {
		c.Header("Access-Control-Allow-Origin", "*")
		c.Header("Access-Control-Allow-Methods", "GET, POST, PUT, DELETE, OPTIONS")
		c.Header("Access-Control-Allow-Headers", "Content-Type, Authorization")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// API路由组
	api := r.Group("/api")
	{
		// 认证路由（无需认证）
		auth := api.Group("/auth")
		{
			auth.POST("/register", controllers.Register)
			auth.POST("/login", controllers.Login)
		}

		// 公共信息路由（获取无需认证）
		profile := api.Group("/profile")
		{
			profile.GET("", controllers.GetPublicProfile)
			// 需要认证的路由
			profile.PUT("", middleware.AuthMiddleware(), controllers.UpdateProfile)
		}

		// 文章路由
		articles := api.Group("/articles")
		{
			// 公共访问（无需认证）
			articles.GET("", controllers.GetArticles)
			articles.GET("/:id", controllers.GetArticle)

			// 需要认证的路由
			articles.POST("", middleware.AuthMiddleware(), controllers.CreateArticle)
			articles.PUT("/:id", middleware.AuthMiddleware(), controllers.UpdateArticle)
			articles.DELETE("/:id", middleware.AuthMiddleware(), controllers.DeleteArticle)
		}

		// 需要认证的用户路由
		user := api.Group("/user").Use(middleware.AuthMiddleware())
		{
			user.GET("/profile", controllers.GetProfile)
		}

		// 图片上传路由（需要认证）
		upload := api.Group("/upload").Use(middleware.AuthMiddleware())
		{
			upload.POST("/presigned-url", controllers.GetPresignedURL)
			upload.DELETE("/image", controllers.DeleteImage)
		}

		// 数据分析路由
		analytics := api.Group("/analytics")
		{
			// 公共数据收集接口（无需认证）
			analytics.POST("/track", controllers.Track)
			analytics.GET("/realtime", controllers.GetRealTimeStats)

			// 需要认证的数据查询接口
			analytics.GET("/daily", middleware.AuthMiddleware(), controllers.GetDailyStats)
			analytics.GET("/range", middleware.AuthMiddleware(), controllers.GetStatsRange)
			analytics.GET("/top-pages", middleware.AuthMiddleware(), controllers.GetTopPages)
			
			// 详细数据分析接口（需认证）
			analytics.GET("/events", middleware.AuthMiddleware(), controllers.GetTrackingEvents)
			analytics.GET("/ip-stats", middleware.AuthMiddleware(), controllers.GetIPStats)
			analytics.GET("/user-agent-stats", middleware.AuthMiddleware(), controllers.GetUserAgentStats)
			analytics.GET("/referer-stats", middleware.AuthMiddleware(), controllers.GetRefererStats)
			analytics.GET("/session-stats", middleware.AuthMiddleware(), controllers.GetSessionStats)
			analytics.GET("/event-type-stats", middleware.AuthMiddleware(), controllers.GetEventTypeStats)
			analytics.GET("/hourly-stats", middleware.AuthMiddleware(), controllers.GetHourlyStats)
			analytics.GET("/path-analysis", middleware.AuthMiddleware(), controllers.GetPathAnalysis)
			analytics.GET("/advanced-stats", middleware.AuthMiddleware(), controllers.GetAdvancedStats)
		}
	}

	// 健康检查
	r.GET("/health", func(c *gin.Context) {
		c.JSON(200, gin.H{
			"status":  "ok",
			"message": "Blog server is running",
		})
	})
}