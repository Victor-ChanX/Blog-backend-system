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
			upload.POST("/image", controllers.UploadImage)
			upload.DELETE("/image", controllers.DeleteImage)
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