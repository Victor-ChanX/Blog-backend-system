package utils

import (
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetRealClientIP 获取真实的客户端IP地址
func GetRealClientIP(c *gin.Context) string {
	// 优先级顺序检查常见的代理头
	headers := []string{
		"X-Forwarded-For",
		"X-Real-IP",
		"CF-Connecting-IP", // Cloudflare
		"X-Forwarded",
		"Forwarded-For",
		"Forwarded",
	}

	for _, header := range headers {
		if ip := c.GetHeader(header); ip != "" {
			// X-Forwarded-For 可能包含多个IP，取第一个
			if header == "X-Forwarded-For" {
				ips := strings.Split(ip, ",")
				if len(ips) > 0 {
					ip = strings.TrimSpace(ips[0])
				}
			}
			
			// 验证IP格式并排除内网IP
			if IsValidPublicIP(ip) {
				return ip
			}
		}
	}

	// 如果没有找到有效的代理头，使用Gin的ClientIP
	clientIP := c.ClientIP()
	
	// 如果是内网IP，尝试从RemoteAddr获取
	if IsPrivateIP(clientIP) {
		if remoteAddr := c.Request.RemoteAddr; remoteAddr != "" {
			if host, _, err := net.SplitHostPort(remoteAddr); err == nil {
				if IsValidPublicIP(host) {
					return host
				}
			}
		}
	}

	return clientIP
}

// IsValidPublicIP 检查是否是有效的公网IP
func IsValidPublicIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return false
	}
	
	// 排除内网IP
	return !IsPrivateIP(ip)
}

// IsPrivateIP 检查是否是内网IP
func IsPrivateIP(ip string) bool {
	parsedIP := net.ParseIP(ip)
	if parsedIP == nil {
		return true // 无效IP当作内网处理
	}

	// 检查是否是内网地址
	privateRanges := []string{
		"127.0.0.0/8",    // 回环地址
		"10.0.0.0/8",     // 私有网络
		"172.16.0.0/12",  // 私有网络
		"192.168.0.0/16", // 私有网络
		"::1/128",        // IPv6回环
		"fc00::/7",       // IPv6私有网络
	}

	for _, cidr := range privateRanges {
		_, network, err := net.ParseCIDR(cidr)
		if err != nil {
			continue
		}
		if network.Contains(parsedIP) {
			return true
		}
	}

	return false
}