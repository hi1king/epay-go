// pkg/utils/trade.go
package utils

import (
	"crypto/rand"
	"fmt"
	"time"
)

// GenerateTradeNo 生成订单号 (时间戳 + 随机数，共24位)
func GenerateTradeNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 5)
	rand.Read(random)
	return fmt.Sprintf("%s%x", timestamp, random)
}

// GenerateWithdrawNo 生成提现单号
func GenerateWithdrawNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 4)
	rand.Read(random)
	return fmt.Sprintf("W%s%x", timestamp, random)
}

// GenerateAPIKey 生成商户API密钥 (32位)
func GenerateAPIKey() string {
	b := make([]byte, 16)
	rand.Read(b)
	return fmt.Sprintf("%x", b)
}

// GenerateRefundNo 生成退款单号
func GenerateRefundNo() string {
	timestamp := time.Now().Format("20060102150405")
	random := make([]byte, 4)
	rand.Read(random)
	return fmt.Sprintf("R%s%x", timestamp, random)
}
