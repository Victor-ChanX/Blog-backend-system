package models

import (
	"time"

	"gorm.io/gorm"
)

// TrackingEvent 追踪事件数据模型
type TrackingEvent struct {
	ID          uint           `json:"id" gorm:"primaryKey"`
	Timestamp   time.Time      `json:"timestamp" gorm:"index;not null"`
	Path        string         `json:"path" gorm:"not null;index"`
	IPAddress   string         `json:"ip_address" gorm:"not null;index"`
	UserAgent   string         `json:"user_agent_hash" gorm:"not null"`
	Referer     string         `json:"referer"`
	EventType   string         `json:"event_type" gorm:"not null;index"` // page_view, article_click, etc.
	ArticleID   *uint          `json:"article_id,omitempty" gorm:"index"`
	SessionID   string         `json:"session_id" gorm:"index"`
	Country     string         `json:"country"`
	City        string         `json:"city"`
	CreatedAt   time.Time      `json:"created_at"`
	UpdatedAt   time.Time      `json:"updated_at"`
	DeletedAt   gorm.DeletedAt `json:"-" gorm:"index"`
}

// DailyStats 每日统计数据
type DailyStats struct {
	ID             uint           `json:"id" gorm:"primaryKey"`
	Date           time.Time      `json:"date" gorm:"uniqueIndex;not null"`
	PageViews      int64          `json:"page_views" gorm:"default:0"`
	UniqueVisitors int64          `json:"unique_visitors" gorm:"default:0"`
	ArticleClicks  int64          `json:"article_clicks" gorm:"default:0"`
	TopPages       string         `json:"top_pages" gorm:"type:jsonb"` // JSON格式存储热门页面
	TopArticles    string         `json:"top_articles" gorm:"type:jsonb"` // JSON格式存储热门文章
	CreatedAt      time.Time      `json:"created_at"`
	UpdatedAt      time.Time      `json:"updated_at"`
	DeletedAt      gorm.DeletedAt `json:"-" gorm:"index"`
}

// RealTimeStats 实时统计数据（Redis缓存结构对应的模型）
type RealTimeStats struct {
	OnlineUsers    int64            `json:"online_users"`
	TodayVisitors  int64            `json:"today_visitors"`
	TodayPageViews int64            `json:"today_page_views"`
	TopPaths       map[string]int64 `json:"top_paths"`
	TopArticles    map[string]int64 `json:"top_articles"`
}

// PageHeatmap 页面热力图数据
type PageHeatmap struct {
	ID        uint           `json:"id" gorm:"primaryKey"`
	Date      time.Time      `json:"date" gorm:"index;not null"`
	Path      string         `json:"path" gorm:"not null;index"`
	Clicks    int64          `json:"clicks" gorm:"default:0"`
	Views     int64          `json:"views" gorm:"default:0"`
	CreatedAt time.Time      `json:"created_at"`
	UpdatedAt time.Time      `json:"updated_at"`
	DeletedAt gorm.DeletedAt `json:"-" gorm:"index"`
}