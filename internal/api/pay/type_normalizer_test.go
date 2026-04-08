package payment

import "testing"

func TestResolvePayRoutingStripe(t *testing.T) {
	tests := []struct {
		rawType      string
		rawPayMethod string
		wantType     string
		wantMethod   string
	}{
		{rawType: "stripe", wantType: "stripe", wantMethod: "checkout"},
		{rawType: "stripe_checkout", wantType: "stripe", wantMethod: "checkout"},
		{rawType: "stripe", rawPayMethod: "web", wantType: "stripe", wantMethod: "web"},
		{rawPayMethod: "checkout", wantType: "stripe", wantMethod: "checkout"},
	}

	for _, tt := range tests {
		got, err := resolvePayRouting(tt.rawType, tt.rawPayMethod)
		if err != nil {
			t.Fatalf("resolvePayRouting(%q, %q) returned error: %v", tt.rawType, tt.rawPayMethod, err)
		}
		if got.PayType != tt.wantType || got.PayMethod != tt.wantMethod {
			t.Fatalf("resolvePayRouting(%q, %q) = (%s, %s), want (%s, %s)", tt.rawType, tt.rawPayMethod, got.PayType, got.PayMethod, tt.wantType, tt.wantMethod)
		}
	}
}
