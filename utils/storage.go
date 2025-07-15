package utils

import (
	"blog-server/config"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/aws/aws-sdk-go/aws"
	"github.com/aws/aws-sdk-go/aws/credentials"
	"github.com/aws/aws-sdk-go/aws/session"
	"github.com/aws/aws-sdk-go/service/s3"
)

type StorageService struct {
	s3Client *s3.S3
}

var Storage *StorageService

// InitStorage 初始化存储服务
func InitStorage() error {
	sess, err := session.NewSession(&aws.Config{
		Region:      aws.String("auto"),
		Endpoint:    aws.String(config.AppConfig.R2Endpoint),
		Credentials: credentials.NewStaticCredentials(config.AppConfig.R2AccessKeyID, config.AppConfig.R2SecretAccessKey, ""),
	})
	if err != nil {
		return fmt.Errorf("创建AWS会话失败: %v", err)
	}

	Storage = &StorageService{
		s3Client: s3.New(sess),
	}

	return nil
}


// DeleteImage 删除R2中的图片
func (s *StorageService) DeleteImage(fileName string) error {
	_, err := s.s3Client.DeleteObject(&s3.DeleteObjectInput{
		Bucket: aws.String(config.AppConfig.R2BucketName),
		Key:    aws.String(fileName),
	})

	if err != nil {
		return fmt.Errorf("删除文件失败: %v", err)
	}

	return nil
}

// generateFileName 生成唯一文件名
func generateFileName(originalName string) string {
	ext := filepath.Ext(originalName)
	timestamp := time.Now().Unix()
	
	// 生成随机字符串
	randomBytes := make([]byte, 8)
	rand.Read(randomBytes)
	randomStr := hex.EncodeToString(randomBytes)
	
	return fmt.Sprintf("images/%d_%s%s", timestamp, randomStr, ext)
}

// isValidImageType 验证是否为有效的图片类型
func isValidImageType(filename string) bool {
	ext := strings.ToLower(filepath.Ext(filename))
	validTypes := []string{".jpg", ".jpeg", ".png", ".gif", ".webp"}
	
	for _, validType := range validTypes {
		if ext == validType {
			return true
		}
	}
	return false
}

// getContentType 根据文件扩展名获取Content-Type
func getContentType(filename string) string {
	ext := strings.ToLower(filepath.Ext(filename))
	
	switch ext {
	case ".jpg", ".jpeg":
		return "image/jpeg"
	case ".png":
		return "image/png"
	case ".gif":
		return "image/gif"
	case ".webp":
		return "image/webp"
	default:
		return "application/octet-stream"
	}
}

// GeneratePresignedURL 生成预签名上传URL
func (s *StorageService) GeneratePresignedURL(fileName string, contentType string) (string, error) {
	req, _ := s.s3Client.PutObjectRequest(&s3.PutObjectInput{
		Bucket:      aws.String(config.AppConfig.R2BucketName),
		Key:         aws.String(fileName),
		ContentType: aws.String(contentType),
	})

	// 设置过期时间为15分钟
	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", fmt.Errorf("生成预签名URL失败: %v", err)
	}

	return url, nil
}

// GeneratePresignedDeleteURL 生成预签名删除URL
func (s *StorageService) GeneratePresignedDeleteURL(fileName string) (string, error) {
	req, _ := s.s3Client.DeleteObjectRequest(&s3.DeleteObjectInput{
		Bucket: aws.String(config.AppConfig.R2BucketName),
		Key:    aws.String(fileName),
	})

	// 设置过期时间为15分钟
	url, err := req.Presign(15 * time.Minute)
	if err != nil {
		return "", fmt.Errorf("生成预签名删除URL失败: %v", err)
	}

	return url, nil
}

// GenerateUniqueFileName 生成唯一文件名（给外部调用）
func GenerateUniqueFileName(originalName string) string {
	return generateFileName(originalName)
}

// GeneratePublicURL 生成公共访问URL
func GeneratePublicURL(fileName string) string {
	return fmt.Sprintf("%s/%s", strings.TrimRight(config.AppConfig.R2PublicURL, "/"), fileName)
}

// ExtractFileNameFromURL 从URL中提取文件名
func ExtractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}