package controllers

import (
	"blog-server/utils"
	"net/http"

	"github.com/gin-gonic/gin"
)

type UploadResponse struct {
	URL      string `json:"url"`
	FileName string `json:"file_name"`
	Size     int64  `json:"size"`
}

// UploadImage 上传图片
func UploadImage(c *gin.Context) {
	// 检查认证
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	// 获取上传的文件
	file, header, err := c.Request.FormFile("image")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "获取上传文件失败: " + err.Error(),
		})
		return
	}
	defer file.Close()

	// 初始化存储服务（如果还未初始化）
	if utils.Storage == nil {
		if err := utils.InitStorage(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "存储服务初始化失败: " + err.Error(),
			})
			return
		}
	}

	// 上传图片
	url, err := utils.Storage.UploadImage(file, header)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "图片上传失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, UploadResponse{
		URL:      url,
		FileName: header.Filename,
		Size:     header.Size,
	})
}

// DeleteImage 删除图片
func DeleteImage(c *gin.Context) {
	// 检查认证
	_, exists := c.Get("user_id")
	if !exists {
		c.JSON(http.StatusUnauthorized, gin.H{
			"error": "未授权",
		})
		return
	}

	type DeleteRequest struct {
		URL string `json:"url" binding:"required"`
	}

	var req DeleteRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "请求参数错误: " + err.Error(),
		})
		return
	}

	// 从URL中提取文件名
	fileName := utils.ExtractFileNameFromURL(req.URL)
	if fileName == "" {
		c.JSON(http.StatusBadRequest, gin.H{
			"error": "无效的图片URL",
		})
		return
	}

	// 初始化存储服务（如果还未初始化）
	if utils.Storage == nil {
		if err := utils.InitStorage(); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{
				"error": "存储服务初始化失败: " + err.Error(),
			})
			return
		}
	}

	// 删除图片
	if err := utils.Storage.DeleteImage(fileName); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{
			"error": "图片删除失败: " + err.Error(),
		})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"message": "图片删除成功",
	})
}