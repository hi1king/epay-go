package payment

import (
	"fmt"
	"strings"
)

type resolvedPayRouting struct {
	PayType   string
	PayMethod string
}

func resolvePayRouting(rawType, rawPayMethod string) (*resolvedPayRouting, error) {
	normalizedType := normalizePayToken(rawType)
	normalizedMethod := normalizePayMethod(rawPayMethod)

	switch normalizedType {
	case "wx_native", "wechat_native", "native_wx", "native_wechat":
		return buildResolvedPayRouting("wxpay", "native", normalizedMethod)
	case "wx_h5", "wechat_h5":
		return buildResolvedPayRouting("wxpay", "h5", normalizedMethod)
	case "wx_jsapi", "wechat_jsapi":
		return buildResolvedPayRouting("wxpay", "jsapi", normalizedMethod)
	case "alipay_scan", "ali_scan", "alipay_qrcode", "ali_qrcode", "alipay_native", "ali_native":
		return buildResolvedPayRouting("alipay", "scan", normalizedMethod)
	case "alipay_h5", "ali_h5", "alipay_wap", "ali_wap":
		return buildResolvedPayRouting("alipay", "h5", normalizedMethod)
	case "alipay_web", "ali_web", "alipay_pc", "ali_pc":
		return buildResolvedPayRouting("alipay", "web", normalizedMethod)
	case "native", "scan", "qrcode", "h5", "wap", "jsapi", "web", "pc", "precreate":
		if normalizedMethod == "" {
			return nil, fmt.Errorf("type=%s 无法唯一确定支付渠道，请改为传 wxpay/alipay，或使用 wx_native、alipay_scan 这类明确值", rawType)
		}
		return nil, fmt.Errorf("type=%s 仅表示支付场景，不能单独作为支付渠道；请改为传 wxpay/alipay，或使用 wx_native、alipay_scan 这类明确值", rawType)
	case "wxpay", "wechat":
		return buildResolvedPayRouting("wxpay", defaultPayMethodForProvider("wxpay", normalizedMethod), normalizedMethod)
	case "alipay", "ali":
		return buildResolvedPayRouting("alipay", defaultPayMethodForProvider("alipay", normalizedMethod), normalizedMethod)
	case "":
		if normalizedMethod == "" {
			return buildResolvedPayRouting("wxpay", "scan", normalizedMethod)
		}
		if isWechatMethod(normalizedMethod) {
			return buildResolvedPayRouting("wxpay", normalizedMethod, normalizedMethod)
		}
		if isAlipayMethod(normalizedMethod) {
			return buildResolvedPayRouting("alipay", normalizedMethod, normalizedMethod)
		}
		return nil, fmt.Errorf("无法根据 pay_method=%s 确定支付渠道，请显式传入 type", rawPayMethod)
	default:
		return buildResolvedPayRouting(normalizedType, normalizedMethod, normalizedMethod)
	}
}

func buildResolvedPayRouting(payType, fallbackMethod, explicitMethod string) (*resolvedPayRouting, error) {
	method := explicitMethod
	if method == "" {
		method = fallbackMethod
	}
	if method == "" {
		return nil, fmt.Errorf("未识别到支付方式")
	}
	if payType == "alipay" && method == "native" {
		method = "scan"
	}
	return &resolvedPayRouting{PayType: payType, PayMethod: method}, nil
}

func normalizePayToken(value string) string {
	replacer := strings.NewReplacer("-", "_", " ", "", ".", "_")
	return strings.ToLower(strings.TrimSpace(replacer.Replace(value)))
}

func normalizePayMethod(value string) string {
	switch normalizePayToken(value) {
	case "", "default":
		return ""
	case "native", "scan", "qrcode", "precreate":
		return "native"
	case "wap", "h5":
		return "h5"
	case "web", "pc", "page":
		return "web"
	case "jsapi":
		return "jsapi"
	default:
		return strings.ToLower(strings.TrimSpace(value))
	}
}

func defaultPayMethodForProvider(payType, payMethod string) string {
	if payMethod != "" {
		return payMethod
	}
	if payType == "alipay" {
		return "scan"
	}
	return "native"
}

func isWechatMethod(payMethod string) bool {
	switch payMethod {
	case "native", "jsapi":
		return true
	default:
		return false
	}
}

func isAlipayMethod(payMethod string) bool {
	switch payMethod {
	case "web":
		return true
	default:
		return false
	}
}
