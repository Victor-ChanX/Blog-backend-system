package controllers

import (
	"blog-server/models"
	"net/http"

	"github.com/gin-gonic/gin"
)

type ProfileRequest struct {
	Name     string `json:"name" binding:"required"`
	Email    string `json:"email"`
	Bio      string `json:"bio"`
	Skills   string `json:"skills"`
	Avatar   string `json:"avatar"`
	Website  string `json:"website"`
	GitHub   string `json:"github"`
	LinkedIn string `json:"linkedin"`
	Twitter  string `json:"twitter"`
	Location string `json:"location"`
	Company  string `json:"company"`
	Position string `json:"position"`
}

// GetPublicProfile 获取公共信息（无需认证）
func GetPublicProfile(c *gin.Context) {
	var profile models.Profile
	// 获取第一条记录，如果没有则创建默认值
	if err := models.DB.First(&profile).Error; err != nil {
		c.JSON(http.StatusOK, gin.H{
			"message": "暂无个人信息",
		})
		return
	}

	c.JSON(http.StatusOK, profile)
}

// UpdateProfile 更新公共信息（需要认证）
func UpdateProfile(c *gin.Context) {
	var req ProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 检查认证
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	var profile models.Profile
	// 尝试获取现有记录
	err := models.DB.First(&profile).Error
	if err != nil {
		// 如果不存在，创建新记录
		profile = models.Profile{
			Name:     req.Name,
			Email:    req.Email,
			Bio:      req.Bio,
			Skills:   req.Skills,
			Avatar:   req.Avatar,
			Website:  req.Website,
			GitHub:   req.GitHub,
			LinkedIn: req.LinkedIn,
			Twitter:  req.Twitter,
			Location: req.Location,
			Company:  req.Company,
			Position: req.Position,
		}

		if err := models.DB.Create(&profile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "个人信息创建失败",
			})
			return
		}
	} else {
		// 更新现有记录
		profile.Name = req.Name
		profile.Email = req.Email
		profile.Bio = req.Bio
		profile.Skills = req.Skills
		profile.Avatar = req.Avatar
		profile.Website = req.Website
		profile.GitHub = req.GitHub
		profile.LinkedIn = req.LinkedIn
		profile.Twitter = req.Twitter
		profile.Location = req.Location
		profile.Company = req.Company
		profile.Position = req.Position

		if err := models.DB.Save(&profile).Error; err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "个人信息更新失败",
			})
			return
		}
	}

	c.JSON(http.StatusOK, profile)
}