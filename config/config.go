package config

import (
	"log"
	"os"

	"github.com/joho/godotenv"
)

type Config struct {
	DatabaseURL      string
	JWTSecret        string
	RegisterPassword string
	Port             string
	Mode             string
	// Cloudflare R2配置
	R2AccessKeyID     string
	R2SecretAccessKey string
	R2BucketName      string
	R2Endpoint        string
	R2PublicURL       string
}

var AppConfig *Config

func LoadConfig() {
	// 加载.env文件
	if err := godotenv.Load(".env"); err != nil {
		log.Printf("加载.env文件失败: %v，使用环境变量", err)
	} else {
		log.Println("成功加载.env文件")
	}

	AppConfig = &Config{
		DatabaseURL:       getEnv("DATABASE_URL", "postgres://username:password@localhost:5432/blog_db?sslmode=disable"),
		JWTSecret:         getEnv("JWT_SECRET", "your-secret-key"),
		RegisterPassword:  getEnv("REGISTER_PASSWORD", "admin123"),
		Port:              getEnv("PORT", "8080"),
		Mode:              getEnv("GIN_MODE", "debug"),
		R2AccessKeyID:     getEnv("R2_ACCESS_KEY_ID", ""),
		R2SecretAccessKey: getEnv("R2_SECRET_ACCESS_KEY", ""),
		R2BucketName:      getEnv("R2_BUCKET_NAME", ""),
		R2Endpoint:        getEnv("R2_ENDPOINT", ""),
		R2PublicURL:       getEnv("R2_PUBLIC_URL", ""),
	}

	log.Println("配置加载完成")
}

func getEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}