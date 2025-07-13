package controllers

import (
	"blog-server/models"
	"net/http"
	"strconv"

	"github.com/gin-gonic/gin"
)

type ArticleRequest struct {
	Title   string `json:"title" binding:"required"`
	Content string `json:"content" binding:"required"`
	Summary string `json:"summary"`
	Status  string `json:"status"` // draft, published
}

// CreateArticle 创建文章
func CreateArticle(c *gin.Context) {
	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	// 设置默认状态
	if req.Status == "" {
		req.Status = "draft"
	}

	article := models.Article{
		Title:   req.Title,
		Content: req.Content,
		Summary: req.Summary,
		Status:  req.Status,
		UserID:  userID.(uint),
	}

	if err := models.DB.Create(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章创建失败",
		})
		return
	}

	// 预加载用户信息
	models.DB.Preload("User").First(&article, article.ID)

	c.JSON(http.StatusCreated, article)
}

// GetArticles 获取文章列表
func GetArticles(c *gin.Context) {
	var articles []models.Article
	query := models.DB.Preload("User")

	// 状态过滤
	if status := c.Query("status"); status != "" {
		query = query.Where("status = ?", status)
	}

	// 分页
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	limit, _ := strconv.Atoi(c.DefaultQuery("limit", "10"))
	offset := (page - 1) * limit

	var total int64
	query.Model(&models.Article{}).Count(&total)

	if err := query.Offset(offset).Limit(limit).Order("created_at DESC").Find(&articles).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "获取文章列表失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"articles": articles,
		"total":    total,
		"page":     page,
		"limit":    limit,
	})
}

// GetArticle 获取单篇文章
func GetArticle(c *gin.Context) {
	id := c.Param("id")
	var article models.Article

	if err := models.DB.Preload("User").First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文章不存在",
		})
		return
	}

	c.JSON(http.StatusOK, article)
}

// UpdateArticle 更新文章
func UpdateArticle(c *gin.Context) {
	id := c.Param("id")
	var req ArticleRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文章不存在",
		})
		return
	}

	// 检查文章所有权
	if article.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权限修改此文章",
		})
		return
	}

	// 更新文章
	article.Title = req.Title
	article.Content = req.Content
	article.Summary = req.Summary
	if req.Status != "" {
		article.Status = req.Status
	}

	if err := models.DB.Save(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章更新失败",
		})
		return
	}

	// 预加载用户信息
	models.DB.Preload("User").First(&article, article.ID)

	c.JSON(http.StatusOK, article)
}

// DeleteArticle 删除文章（软删除）
func DeleteArticle(c *gin.Context) {
	id := c.Param("id")
	userID, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var article models.Article
	if err := models.DB.First(&article, id).Error; err != nil {
		c.JSON(http.StatusNotFound, gin.H{
			"error": "文章不存在",
		})
		return
	}

	// 检查文章所有权
	if article.UserID != userID.(uint) {
		c.JSON(http.StatusForbidden, gin.H{
			"error": "无权限删除此文章",
		})
		return
	}

	// 软删除
	if err := models.DB.Delete(&article).Error; err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章删除失败",
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "文章删除成功",
	})
}