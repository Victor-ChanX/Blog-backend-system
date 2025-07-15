package models

import (
	"encoding/json"
	"time"

	"gorm.io/gorm"
)

type Article struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Title     string         `json:"title" gorm:"not null"`
	Summary   string         `json:"summary" gorm:"type:text"` // 文章摘要
	Status    string         `json:"status" gorm:"default:draft"` // draft, published
	UserID    uint           `json:"user_id" gorm:"not null"`
	User      User           `json:"user" gorm:"foreignKey:UserID"`
	Content   *ArticleContent `json:"-" gorm:"foreignKey:ArticleID"` // 关联文章内容，JSON中隐藏
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // 软删除
}

// MarshalJSON 自定义JSON序列化
func (a Article) MarshalJSON() ([]byte, error) {
	type Alias Article
	aux := &struct {
		*Alias
		Content string `json:"content,omitempty"`
	}{
		Alias: (*Alias)(&a),
	}
	
	// 如果有内容，则设置content字段
	if a.Content != nil {
		aux.Content = a.Content.Content
	}
	
	return json.Marshal(aux)
}

type ArticleContent struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	ArticleID uint           `json:"article_id" gorm:"not null;index"`
	Content   string         `json:"content" gorm:"type:text"` // Markdown内容
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"` // 软删除
}