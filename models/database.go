package models

import (
	"blog-server/config"
	"log"

	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

var DB *gorm.DB

func InitDB() {
	var err error
	DB, err = gorm.Open(postgres.Open(config.AppConfig.DatabaseURL), &gorm.Config{})
	if err != nil {
		log.Fatal("数据库连接失败:", err)
	}

	log.Println("数据库连接成功")
}

func Migrate() {
	// 使用新的迁移系统
	if err := RunMigrations(); err != nil {
		log.Fatal("数据库迁移失败:", err)
	}
	log.Println("数据库迁移完成")
}