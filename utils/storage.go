package utils

import (
	"blog-server/config"
	"bytes"
	"crypto/rand"
	"encoding/hex"
	"fmt"
	"io"
	"mime/multipart"
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

// UploadImage 上传图片到R2
func (s *StorageService) UploadImage(file multipart.File, header *multipart.FileHeader) (string, error) {
	// 验证文件类型
	if !isValidImageType(header.Filename) {
		return "", fmt.Errorf("不支持的文件类型")
	}

	// 验证文件大小 (限制为5MB)
	if header.Size > 5*1024*1024 {
		return "", fmt.Errorf("文件大小不能超过5MB")
	}

	// 生成唯一文件名
	fileName := generateFileName(header.Filename)

	// 读取文件内容
	fileContent, err := io.ReadAll(file)
	if err != nil {
		return "", fmt.Errorf("读取文件失败: %v", err)
	}

	// 上传到R2
	_, err = s.s3Client.PutObject(&s3.PutObjectInput{
		Bucket:        aws.String(config.AppConfig.R2BucketName),
		Key:           aws.String(fileName),
		Body:          bytes.NewReader(fileContent),
		ContentType:   aws.String(getContentType(header.Filename)),
		ContentLength: aws.Int64(header.Size),
	})

	if err != nil {
		return "", fmt.Errorf("上传文件失败: %v", err)
	}

	// 返回公共URL
	publicURL := fmt.Sprintf("%s/%s", strings.TrimRight(config.AppConfig.R2PublicURL, "/"), fileName)
	return publicURL, nil
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

// ExtractFileNameFromURL 从URL中提取文件名
func ExtractFileNameFromURL(url string) string {
	parts := strings.Split(url, "/")
	if len(parts) > 0 {
		return parts[len(parts)-1]
	}
	return ""
}