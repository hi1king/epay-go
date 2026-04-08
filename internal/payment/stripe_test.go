package payment

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
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
