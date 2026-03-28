// pkg/utils/utils.go
package utils

import (
	"crypto/md5"
	"encoding/hex"
	"net"
	"strings"

	"github.com/gin-gonic/gin"
)

// GetClientIP 获取客户端真实IP
func GetClientIP(c *gin.Context) string {
	// 优先从 X-Forwarded-For 获取
	xff := c.GetHeader("X-Forwarded-For")
	if xff != "" {
		ips := strings.Split(xff, ",")
		if len(ips) > 0 {
			ip := strings.TrimSpace(ips[0])
			if net.ParseIP(ip) != nil {
				return ip
			}
		}
	}

	// 其次从 X-Real-IP 获取
	xri := c.GetHeader("X-Real-IP")
	if xri != "" && net.ParseIP(xri) != nil {
		return xri
	}

	// 最后使用 RemoteAddr
	ip, _, _ := net.SplitHostPort(c.Request.RemoteAddr)
	return ip
}

// MD5 计算MD5哈希
func MD5(str string) string {
	h := md5.New()
	h.Write([]byte(str))
	return hex.EncodeToString(h.Sum(nil))
}

// ContainsString 检查字符串切片是否包含指定字符串
func ContainsString(slice []string, str string) bool {
	for _, s := range slice {
		if s == str {
			return true
		}
	}
	return false
}
