// internal/plugin/stripe_test.go
package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"strconv"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func TestStripeAmountConversion(t *testing.T) {
	amount, err := stripeAmountToMinor(decimal.RequireFromString("10.25"), "usd")
	if err != nil {
		t.Fatalf("stripeAmountToMinor returned error: %v", err)
	}
	if amount != 1025 {
		t.Fatalf("expected 1025, got %d", amount)
	}

	if got := stripeMinorToAmount(1025, "usd"); !got.Equal(decimal.RequireFromString("10.25")) {
		t.Fatalf("unexpected amount: %s", got.String())
	}
}

func TestStripeZeroDecimalCurrency(t *testing.T) {
	amount, err := stripeAmountToMinor(decimal.RequireFromString("100"), "jpy")
	if err != nil {
		t.Fatalf("stripeAmountToMinor returned error: %v", err)
	}
	if amount != 100 {
		t.Fatalf("expected 100, got %d", amount)
	}

	if _, err := stripeAmountToMinor(decimal.RequireFromString("100.50"), "jpy"); err == nil {
		t.Fatal("expected zero-decimal currency validation error")
	}
}

func TestVerifyStripeWebhookSignature(t *testing.T) {
	secret := "whsec_test"
	payload := []byte(`{"id":"evt_test","type":"payment_intent.succeeded"}`)
	timestamp := time.Now().Unix()

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	signature := hex.EncodeToString(mac.Sum(nil))

	header := "t=" + strconv.FormatInt(timestamp, 10) + ",v1=" + signature
	if err := verifyStripeWebhookSignature(secret, header, payload, 5*time.Minute); err != nil {
		t.Fatalf("verifyStripeWebhookSignature returned error: %v", err)
	}
}

func TestStripeConfigLegacyAliases(t *testing.T) {
	adapter, err := NewStripeAdapter(json.RawMessage(`{
		"appid": "sk_test_legacy",
		"appkey": "whsec_legacy",
		"currency_code": "USD",
		"currency_rate": "0.137",
		"paymentMethods": "card,paypal"
	}`))
	if err != nil {
		t.Fatalf("NewStripeAdapter returned error: %v", err)
	}

	stripeAdapter, ok := adapter.(*StripeAdapter)
	if !ok {
		t.Fatalf("adapter type = %T, want *StripeAdapter", adapter)
	}
	if stripeAdapter.config.SecretKey != "sk_test_legacy" {
		t.Fatalf("unexpected secret key: %s", stripeAdapter.config.SecretKey)
	}
	if stripeAdapter.config.WebhookSecret != "whsec_legacy" {
		t.Fatalf("unexpected webhook secret: %s", stripeAdapter.config.WebhookSecret)
	}
	if stripeAdapter.config.Currency != "usd" {
		t.Fatalf("unexpected currency: %s", stripeAdapter.config.Currency)
	}
	if !stripeAdapter.currencyRate.Equal(decimal.RequireFromString("0.137")) {
		t.Fatalf("unexpected currency rate: %s", stripeAdapter.currencyRate.String())
	}
	if stripeAdapter.config.PaymentMethodTypes != "card,paypal" {
		t.Fatalf("unexpected payment methods: %s", stripeAdapter.config.PaymentMethodTypes)
	}
}

func TestResolveStripeRoute(t *testing.T) {
	tests := []struct {
		payType    string
		payMethod  string
		wantKind   stripeRouteKind
		wantMethod string
	}{
		{payType: "alipay", wantKind: stripeRouteKindDirect, wantMethod: stripeDirectMethodAlipay},
		{payType: "wxpay", wantKind: stripeRouteKindDirect, wantMethod: stripeDirectMethodWechatPay},
		{payType: "paypal", wantKind: stripeRouteKindCheckout, wantMethod: stripeCheckoutMethodPaypal},
		{payType: "bank", wantKind: stripeRouteKindCheckout, wantMethod: stripeCheckoutMethodCard},
		{payType: "stripe", payMethod: "wechat_pay", wantKind: stripeRouteKindDirect, wantMethod: stripeDirectMethodWechatPay},
	}

	for _, tt := range tests {
		got := resolveStripeRoute(tt.payType, tt.payMethod)
		if got.Kind != tt.wantKind {
			t.Fatalf("resolveStripeRoute(%q, %q) kind = %s, want %s", tt.payType, tt.payMethod, got.Kind, tt.wantKind)
		}
		if tt.wantKind == stripeRouteKindDirect {
			if got.PaymentMethod != tt.wantMethod {
				t.Fatalf("resolveStripeRoute(%q, %q) payment method = %s, want %s", tt.payType, tt.payMethod, got.PaymentMethod, tt.wantMethod)
			}
		} else if len(got.CheckoutMethodTypes) == 0 || got.CheckoutMethodTypes[0] != tt.wantMethod {
			t.Fatalf("resolveStripeRoute(%q, %q) checkout methods = %v, want first %s", tt.payType, tt.payMethod, got.CheckoutMethodTypes, tt.wantMethod)
		}
	}
}
