package models

import (
	"time"
)

type Profile struct {
	ID        uint      `json:"id" gorm:"primaryKey"`
	Name      string    `json:"name" gorm:"not null"`
	Email     string    `json:"email"`
	Bio       string    `json:"bio" gorm:"type:text"`    // 个人介绍
	Skills    string    `json:"skills" gorm:"type:text"` // 技能，JSON格式存储
	Avatar    string    `json:"avatar"`                  // 头像URL
	Website   string    `json:"website"`                 // 个人网站
	GitHub    string    `json:"github"`                  // GitHub链接
	LinkedIn  string    `json:"linkedin"`                // LinkedIn链接
	Twitter   string    `json:"twitter"`                 // Twitter链接
	Location  string    `json:"location"`                // 所在地
	Company   string    `json:"company"`                 // 公司
	Position  string    `json:"position"`                // 职位
	CreatedAt time.Time `json:"created_at"`
	UpdatedAt time.Time `json:"updated_at"`
}
