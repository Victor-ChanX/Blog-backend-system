package models

import (
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Content   string         `json:"content" gorm:"type:text"` // Markdown内容
	Summary   string         `json:"summary" gorm:"type:text"` // 文章摘要
	Status    string         `json:"status" gorm:"default:draft"` // draft, published
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // 软删除
}