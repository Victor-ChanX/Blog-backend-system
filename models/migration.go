package models

import (
	"fmt"
	"log"
	"time"

	"gorm.io/gorm"
)

type Migration struct {
	ID        uint      `gorm:"primaryKey"`
	Version   string    `gorm:"unique;not null"`
	Name      string    `gorm:"not null"`
	AppliedAt time.Time `gorm:"default:CURRENT_TIMESTAMP"`
}

type MigrationFunc func(*gorm.DB) error

type MigrationItem struct {
	Version string
	Name    string
	Up      MigrationFunc
	Down    MigrationFunc
}

var migrations = []MigrationItem{
	{
		Version: "001",
		Name:    "create_initial_tables",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&User{}, &Article{}, &Profile{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&User{}, &Article{}, &Profile{})
		},
	},
	{
		Version: "002",
		Name:    "create_api_log_table",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&APILog{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&APILog{})
		},
	},
	{
		Version: "003",
		Name:    "separate_article_content",
		Up: func(db *gorm.DB) error {
			// 创建文章内容表
			if err := db.AutoMigrate(&ArticleContent{}); err != nil {
				return err
			}

			// 检查articles表是否存在content字段
			if db.Migrator().HasColumn(&Article{}, "content") {
				// 迁移现有文章内容到新表
				var articles []struct {
					ID        uint
					Content   string
					CreatedAt time.Time
					UpdatedAt time.Time
				}
				
				if err := db.Table("articles").Select("id, content, created_at, updated_at").Find(&articles).Error; err != nil {
					return err
				}

				for _, article := range articles {
					// 检查是否已有内容记录
					var existingContent ArticleContent
					if err := db.Where("article_id = ?", article.ID).First(&existingContent).Error; err == gorm.ErrRecordNotFound {
						// 创建内容记录
						articleContent := ArticleContent{
							ArticleID: article.ID,
							Content:   article.Content,
							CreatedAt: article.CreatedAt,
							UpdatedAt: article.UpdatedAt,
						}
						if err := db.Create(&articleContent).Error; err != nil {
							return err
						}
					}
				}

				// 删除原文章表的content列
				if err := db.Migrator().DropColumn(&Article{}, "content"); err != nil {
					return err
				}
			}

			return nil
		},
		Down: func(db *gorm.DB) error {
			// 添加回content列
			if !db.Migrator().HasColumn(&Article{}, "content") {
				if err := db.Migrator().AddColumn(&Article{}, "content"); err != nil {
					return err
				}
			}

			// 恢复内容到文章表
			var articles []Article
			if err := db.Find(&articles).Error; err != nil {
				return err
			}

			for _, article := range articles {
				var content ArticleContent
				if err := db.Where("article_id = ?", article.ID).First(&content).Error; err == nil {
					// 更新文章内容
					if err := db.Model(&article).Update("content", content.Content).Error; err != nil {
						return err
					}
				}
			}

			// 删除文章内容表
			return db.Migrator().DropTable(&ArticleContent{})
		},
	},
	{
		Version: "004",
		Name:    "create_analytics_tables",
		Up: func(db *gorm.DB) error {
			return db.AutoMigrate(&TrackingEvent{}, &DailyStats{}, &PageHeatmap{})
		},
		Down: func(db *gorm.DB) error {
			return db.Migrator().DropTable(&TrackingEvent{}, &DailyStats{}, &PageHeatmap{})
		},
	},
}

// RunMigrations 执行所有未应用的迁移
func RunMigrations() error {
	// 确保migration表存在
	if err := DB.AutoMigrate(&Migration{}); err != nil {
		return fmt.Errorf("创建迁移表失败: %v", err)
	}

	for _, migration := range migrations {
		var existingMigration Migration
		result := DB.Where("version = ?", migration.Version).First(&existingMigration)
		
		if result.Error == gorm.ErrRecordNotFound {
			log.Printf("执行迁移: %s - %s", migration.Version, migration.Name)
			
			// 在事务中执行迁移
			err := DB.Transaction(func(tx *gorm.DB) error {
				// 执行迁移
				if err := migration.Up(tx); err != nil {
					return err
				}
				
				// 记录迁移
				migrationRecord := Migration{
					Version:   migration.Version,
					Name:      migration.Name,
					AppliedAt: time.Now(),
				}
				return tx.Create(&migrationRecord).Error
			})
			
			if err != nil {
				return fmt.Errorf("迁移 %s 失败: %v", migration.Version, err)
			}
			
			log.Printf("迁移 %s 完成", migration.Version)
		} else if result.Error != nil {
			return fmt.Errorf("检查迁移状态失败: %v", result.Error)
		} else {
			log.Printf("迁移 %s 已应用，跳过", migration.Version)
		}
	}
	
	return nil
}

// RollbackMigration 回滚指定版本的迁移
func RollbackMigration(version string) error {
	var migration *MigrationItem
	for _, m := range migrations {
		if m.Version == version {
			migration = &m
			break
		}
	}
	
	if migration == nil {
		return fmt.Errorf("未找到版本 %s 的迁移", version)
	}
	
	// 检查迁移是否已应用
	var existingMigration Migration
	result := DB.Where("version = ?", version).First(&existingMigration)
	if result.Error == gorm.ErrRecordNotFound {
		return fmt.Errorf("迁移 %s 未应用，无法回滚", version)
	}
	
	log.Printf("回滚迁移: %s - %s", migration.Version, migration.Name)
	
	// 在事务中执行回滚
	err := DB.Transaction(func(tx *gorm.DB) error {
		// 执行回滚
		if err := migration.Down(tx); err != nil {
			return err
		}
		
		// 删除迁移记录
		return tx.Delete(&existingMigration).Error
	})
	
	if err != nil {
		return fmt.Errorf("回滚迁移 %s 失败: %v", version, err)
	}
	
	log.Printf("迁移 %s 回滚完成", version)
	return nil
}

// GetMigrationStatus 获取迁移状态
func GetMigrationStatus() ([]map[string]interface{}, error) {
	var appliedMigrations []Migration
	if err := DB.Order("version").Find(&appliedMigrations).Error; err != nil {
		return nil, err
	}
	
	appliedMap := make(map[string]Migration)
	for _, m := range appliedMigrations {
		appliedMap[m.Version] = m
	}
	
	var status []map[string]interface{}
	for _, migration := range migrations {
		applied, exists := appliedMap[migration.Version]
		item := map[string]interface{}{
			"version": migration.Version,
			"name":    migration.Name,
			"applied": exists,
		}
		if exists {
			item["applied_at"] = applied.AppliedAt
		}
		status = append(status, item)
	}
	
	return status, nil
}