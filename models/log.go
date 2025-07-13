package models

import (
	"time"
)

type APILog struct {
	ID           uint      `json:"id" gorm:"primaryKey"`
	Method       string    `json:"method" gorm:"not null"`        // HTTP方法
	Path         string    `json:"path" gorm:"not null"`          // 请求路径
	StatusCode   int       `json:"status_code" gorm:"not null"`   // 响应状态码
	ResponseTime int64     `json:"response_time"`                 // 响应时间(毫秒)
	UserAgent    string    `json:"user_agent"`                    // 用户代理
	IP           string    `json:"ip"`                            // 客户端IP
	UserID       *uint     `json:"user_id" gorm:"index"`          // 用户ID(可为空)
	FunctionName string    `json:"function_name"`                 // 调用的函数名
	Level        string    `json:"level" gorm:"default:info"`     // 日志级别: info, warn, error
	ErrorMessage string    `json:"error_message" gorm:"type:text"` // 错误信息(如果有)
	RequestBody  string    `json:"request_body" gorm:"type:text"` // 请求体(敏感信息已过滤)
	CreatedAt    time.Time `json:"created_at"`
}