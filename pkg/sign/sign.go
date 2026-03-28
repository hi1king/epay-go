// pkg/sign/sign.go
package sign

import (
	"crypto/md5"
	"encoding/hex"
	"net/url"
	"sort"
	"strings"
)

// VerifyMD5Sign 验证MD5签名（与原epay兼容）
func VerifyMD5Sign(params url.Values, key, sign string) bool {
	// 按key排序
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" && params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	// 拼接字符串
	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params.Get(k))
	}
	buf.WriteString(key)

	// MD5
	hash := md5.Sum([]byte(buf.String()))
	expected := hex.EncodeToString(hash[:])

	return strings.EqualFold(expected, sign)
}

// GenerateMD5Sign 生成MD5签名
func GenerateMD5Sign(params url.Values, key string) string {
	var keys []string
	for k := range params {
		if k != "sign" && k != "sign_type" && params.Get(k) != "" {
			keys = append(keys, k)
		}
	}
	sort.Strings(keys)

	var buf strings.Builder
	for i, k := range keys {
		if i > 0 {
			buf.WriteString("&")
		}
		buf.WriteString(k)
		buf.WriteString("=")
		buf.WriteString(params.Get(k))
	}
	buf.WriteString(key)

	hash := md5.Sum([]byte(buf.String()))
	return hex.EncodeToString(hash[:])
}
