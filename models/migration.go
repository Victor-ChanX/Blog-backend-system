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