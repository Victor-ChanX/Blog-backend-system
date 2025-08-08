package middleware

import (
	"blog-server/models"
	"blog-server/utils"
	"bytes"
	"encoding/json"
	"io"
	"runtime"
	"strings"
	"time"

	"github.com/gin-gonic/gin"
)

type responseWriter struct {
	gin.ResponseWriter
	body       *bytes.Buffer
	statusCode int
}

func (w *responseWriter) Write(b []byte) (int, error) {
	w.body.Write(b)
	return w.ResponseWriter.Write(b)
}

func (w *responseWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

// LoggerMiddleware 日志记录中间件
func LoggerMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// 创建响应记录器
		blw := &responseWriter{
			body:           bytes.NewBufferString(""),
			ResponseWriter: c.Writer,
			statusCode:     200,
		}
		c.Writer = blw

		// 读取请求体
		var requestBody string
		if c.Request.Body != nil {
			bodyBytes, _ := io.ReadAll(c.Request.Body)
			c.Request.Body = io.NopCloser(bytes.NewBuffer(bodyBytes))
			requestBody = filterSensitiveData(string(bodyBytes))
		}

		// 获取调用的函数名
		functionName := getFunctionName()

		// 处理请求
		c.Next()

		// 计算响应时间
		responseTime := time.Since(startTime).Milliseconds()

		// 获取用户ID
		var userID *uint
		if uid, exists := c.Get("user_id"); exists {
			if id, ok := uid.(uint); ok {
				userID = &id
			}
		}

		// 确定日志级别和错误信息
		level := "info"
		var errorMessage string
		
		if blw.statusCode >= 400 {
			level = "error"
			// 尝试从响应中提取错误信息
			if response := blw.body.String(); response != "" {
				var respMap map[string]interface{}
				if err := json.Unmarshal(blw.body.Bytes(), &respMap); err == nil {
					if errMsg, exists := respMap["error"]; exists {
						if errStr, ok := errMsg.(string); ok {
							errorMessage = errStr
						}
					}
				}
			}
		} else if blw.statusCode >= 300 {
			level = "warn"
		}

		// 创建日志记录
		apiLog := models.APILog{
			Method:       c.Request.Method,
			Path:         c.Request.URL.Path,
			StatusCode:   blw.statusCode,
			ResponseTime: responseTime,
			UserAgent:    c.Request.UserAgent(),
			IP:           utils.GetRealClientIP(c),
			UserID:       userID,
			FunctionName: functionName,
			Level:        level,
			ErrorMessage: errorMessage,
			RequestBody:  requestBody,
		}

		// 异步保存日志到数据库
		go func() {
			if err := models.DB.Create(&apiLog).Error; err != nil {
				// 如果数据库保存失败，至少输出到控制台
				println("日志保存失败:", err.Error())
			}
		}()
	}
}

// getFunctionName 获取当前处理的函数名
func getFunctionName() string {
	pc := make([]uintptr, 15)
	n := runtime.Callers(1, pc)
	frames := runtime.CallersFrames(pc[:n])
	
	for {
		frame, more := frames.Next()
		// 查找controllers包中的函数
		if strings.Contains(frame.Function, "controllers.") {
			parts := strings.Split(frame.Function, ".")
			if len(parts) > 0 {
				return parts[len(parts)-1]
			}
		}
		if !more {
			break
		}
	}
	return "unknown"
}

// filterSensitiveData 过滤敏感数据
func filterSensitiveData(body string) string {
	if body == "" {
		return ""
	}

	var data map[string]interface{}
	if err := json.Unmarshal([]byte(body), &data); err != nil {
		return body // 如果不是JSON，直接返回
	}

	// 过滤敏感字段
	sensitiveFields := []string{"password", "secret", "token"}
	for _, field := range sensitiveFields {
		if _, exists := data[field]; exists {
			data[field] = "***"
		}
	}

	filtered, err := json.Marshal(data)
	if err != nil {
		return body
	}

	return string(filtered)
}

