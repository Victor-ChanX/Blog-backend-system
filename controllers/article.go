package controllers

import (
	"blog-server/models"
	"net/http"
	"strconv"
	"strings"

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

	// 使用事务确保数据一致性
	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 创建文章基本信息
	article := models.Article{
		Title:   req.Title,
		Summary: req.Summary,
		Status:  req.Status,
		UserID:  userID.(uint),
	}

	if err := tx.Create(&article).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章创建失败",
		})
		return
	}

	// 创建文章内容
	articleContent := models.ArticleContent{
		ArticleID: article.ID,
		Content:   req.Content,
	}

	if err := tx.Create(&articleContent).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章内容创建失败",
		})
		return
	}

	tx.Commit()

	// 预加载用户信息和内容
	models.DB.Preload("User").Preload("Content").First(&article, article.ID)

	c.JSON(http.StatusCreated, article)
}

// GetArticles 获取文章列表
func GetArticles(c *gin.Context) {
	var articles []models.Article
	query := models.DB.Model(&models.Article{})

	// 字段选择
	fields := c.Query("fields")
	if fields != "" {
		selectedFields := parseFields(fields)
		query = query.Select(selectedFields)
		
		// 如果需要用户信息，则预加载
		if needsUserInfo(selectedFields) {
			query = query.Preload("User")
		}
		
		// 如果需要内容信息，则预加载
		if needsContentInfo(selectedFields) {
			query = query.Preload("Content")
		}
	} else {
		// 默认预加载用户信息，但不加载内容
		query = query.Preload("User")
	}

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
	
	query := models.DB.Model(&models.Article{})

	// 字段选择
	fields := c.Query("fields")
	if fields != "" {
		selectedFields := parseFields(fields)
		query = query.Select(selectedFields)
		
		// 如果需要用户信息，则预加载
		if needsUserInfo(selectedFields) {
			query = query.Preload("User")
		}
		
		// 如果需要内容信息，则预加载
		if needsContentInfo(selectedFields) {
			query = query.Preload("Content")
		}
	} else {
		// 默认预加载用户信息和内容（获取单篇文章通常需要完整内容）
		query = query.Preload("User").Preload("Content")
	}

	if err := query.First(&article, id).Error; err != nil {
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

	// 使用事务确保数据一致性
	tx := models.DB.Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 更新文章基本信息
	article.Title = req.Title
	article.Summary = req.Summary
	if req.Status != "" {
		article.Status = req.Status
	}

	if err := tx.Save(&article).Error; err != nil {
		tx.Rollback()
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "文章更新失败",
		})
		return
	}

	// 更新文章内容
	var articleContent models.ArticleContent
	if err := tx.Where("article_id = ?", article.ID).First(&articleContent).Error; err != nil {
		// 如果内容不存在，创建新的
		articleContent = models.ArticleContent{
			ArticleID: article.ID,
			Content:   req.Content,
		}
		if err := tx.Create(&articleContent).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "文章内容创建失败",
			})
			return
		}
	} else {
		// 更新现有内容
		articleContent.Content = req.Content
		if err := tx.Save(&articleContent).Error; err != nil {
			tx.Rollback()
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "文章内容更新失败",
			})
			return
		}
	}

	tx.Commit()

	// 预加载用户信息和内容
	models.DB.Preload("User").Preload("Content").First(&article, article.ID)

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

// parseFields 解析fields参数
func parseFields(fields string) []string {
	if fields == "" {
		return nil
	}
	
	fieldList := strings.Split(fields, ",")
	var selectedFields []string
	
	// 定义允许的字段
	allowedFields := map[string]string{
		"id":         "id",
		"title":      "title",
		"content":    "content",  // 这个字段会触发内容表的预加载
		"summary":    "summary",
		"status":     "status",
		"user_id":    "user_id",
		"created_at": "created_at",
		"updated_at": "updated_at",
	}
	
	for _, field := range fieldList {
		field = strings.TrimSpace(field)
		if dbField, exists := allowedFields[field]; exists {
			selectedFields = append(selectedFields, dbField)
		}
	}
	
	// 如果没有有效字段，返回nil使用默认查询
	if len(selectedFields) == 0 {
		return nil
	}
	
	return selectedFields
}

// needsUserInfo 检查是否需要用户信息
func needsUserInfo(selectedFields []string) bool {
	if selectedFields == nil {
		return true // 默认情况下需要用户信息
	}
	
	// 如果选择了user_id字段，通常需要用户信息
	for _, field := range selectedFields {
		if field == "user_id" {
			return true
		}
	}
	
	return false
}

// needsContentInfo 检查是否需要内容信息
func needsContentInfo(selectedFields []string) bool {
	if selectedFields == nil {
		return false // 默认情况下不需要内容信息（除非是获取单篇文章）
	}
	
	// 如果选择了content字段，需要内容信息
	for _, field := range selectedFields {
		if field == "content" {
			return true
		}
	}
	
	return false
}