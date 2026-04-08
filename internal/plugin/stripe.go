// internal/plugin/stripe.go
package payment

import (
	"context"
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"time"

	"github.com/shopspring/decimal"
)

const (
	stripeAPIBaseURL            = "https://api.stripe.com/v1"
	stripeWebhookTolerance      = 5 * time.Minute
	defaultStripeCurrency       = "usd"
	defaultStripePayMethods     = "card"
	stripeDirectMethodAlipay    = "alipay"
	stripeDirectMethodWechatPay = "wechat_pay"
	stripeCheckoutMethodCard    = "card"
	stripeCheckoutMethodPaypal  = "paypal"
	stripeCheckoutMethodLink    = "link"
	stripeMetadataTradeNo       = "trade_no"
	stripeMetadataSubject       = "subject"
	stripeMetadataNotifyURL     = "notify_url"
)

var stripeZeroDecimalCurrencies = map[string]struct{}{
	"bif": {}, "clp": {}, "djf": {}, "gnf": {}, "jpy": {}, "kmf": {},
	"krw": {}, "mga": {}, "pyg": {}, "rwf": {}, "ugx": {}, "vnd": {},
	"vuv": {}, "xaf": {}, "xof": {}, "xpf": {},
}

// StripeConfig Stripe 支付配置
type StripeConfig struct {
	SecretKey          string `json:"secret_key"`
	PublishableKey     string `json:"publishable_key"`
	WebhookSecret      string `json:"webhook_secret"`
	SuccessURL         string `json:"success_url"`
	CancelURL          string `json:"cancel_url"`
	Currency           string `json:"currency"`
	CurrencyRate       string `json:"currency_rate"`
	PaymentMethodTypes string `json:"payment_method_types"`
}

// StripeAdapter Stripe 支付适配器
type StripeAdapter struct {
	client       *http.Client
	config       *StripeConfig
	currencyRate decimal.Decimal
}

type stripeCheckoutSession struct {
	ID                string            `json:"id"`
	URL               string            `json:"url"`
	PaymentIntent     string            `json:"payment_intent"`
	ClientReferenceID string            `json:"client_reference_id"`
	AmountTotal       int64             `json:"amount_total"`
	Currency          string            `json:"currency"`
	PaymentStatus     string            `json:"payment_status"`
	Customer          string            `json:"customer"`
	Metadata          map[string]string `json:"metadata"`
	CustomerDetails   *stripeCustomer   `json:"customer_details"`
	SuccessURL        string            `json:"success_url"`
	CancelURL         string            `json:"cancel_url"`
}

type stripeCustomer struct {
	Email string `json:"email"`
	Name  string `json:"name"`
	Phone string `json:"phone"`
}

type stripePaymentIntent struct {
	ID                  string                     `json:"id"`
	Status              string                     `json:"status"`
	Amount              int64                      `json:"amount"`
	AmountReceived      int64                      `json:"amount_received"`
	Currency            string                     `json:"currency"`
	Created             int64                      `json:"created"`
	Customer            string                     `json:"customer"`
	ReceiptEmail        string                     `json:"receipt_email"`
	Metadata            map[string]string          `json:"metadata"`
	LastPaymentError    *stripePaymentIntentError  `json:"last_payment_error"`
	NextAction          *stripePaymentIntentAction `json:"next_action"`
	PaymentMethod       string                     `json:"payment_method"`
	PaymentMethodTypes  []string                   `json:"payment_method_types"`
	ClientSecret        string                     `json:"client_secret"`
	LatestCharge        string                     `json:"latest_charge"`
	AmountCapturable    int64                      `json:"amount_capturable"`
	AmountDetails       map[string]json.RawMessage `json:"amount_details"`
	StatementDescriptor string                     `json:"statement_descriptor"`
	Description         string                     `json:"description"`
}

type stripePaymentIntentError struct {
	Code    string `json:"code"`
	Message string `json:"message"`
	Type    string `json:"type"`
}

type stripePaymentIntentAction struct {
	Type                 string                          `json:"type"`
	RedirectToURL        *stripeRedirectToURLAction      `json:"redirect_to_url"`
	AlipayHandleRedirect *stripeRedirectToURLAction      `json:"alipay_handle_redirect"`
	WechatPayDisplayQR   *stripeWechatPayDisplayQRAction `json:"wechat_pay_display_qr_code"`
	UseStripeSDK         map[string]json.RawMessage      `json:"use_stripe_sdk"`
}

type stripeRedirectToURLAction struct {
	URL       string `json:"url"`
	ReturnURL string `json:"return_url"`
}

type stripeWechatPayDisplayQRAction struct {
	Data                  string `json:"data"`
	ImageDataURL          string `json:"image_data_url"`
	HostedInstructionsURL string `json:"hosted_instructions_url"`
}

type stripePaymentIntentSearchResponse struct {
	Data []stripePaymentIntent `json:"data"`
}

type stripeRefund struct {
	ID            string `json:"id"`
	Status        string `json:"status"`
	FailureReason string `json:"failure_reason"`
	PaymentIntent string `json:"payment_intent"`
	Amount        int64  `json:"amount"`
}

type stripeWebhookEvent struct {
	Type string `json:"type"`
	Data struct {
		Object json.RawMessage `json:"object"`
	} `json:"data"`
}

type stripeAPIErrorResponse struct {
	Error *struct {
		Message string `json:"message"`
		Type    string `json:"type"`
		Code    string `json:"code"`
		Param   string `json:"param"`
	} `json:"error"`
}

// NewStripeAdapter 创建 Stripe 适配器
func NewStripeAdapter(configJSON json.RawMessage) (PaymentAdapter, error) {
	var m map[string]interface{}
	_ = json.Unmarshal(configJSON, &m)
	if m != nil {
		if _, ok := m["secret_key"]; !ok {
			if v, ok2 := m["api_key"]; ok2 {
				m["secret_key"] = v
			} else if v, ok2 := m["secretKey"]; ok2 {
				m["secret_key"] = v
			} else if v, ok2 := m["appid"]; ok2 {
				m["secret_key"] = v
			}
		}
		if _, ok := m["webhook_secret"]; !ok {
			if v, ok2 := m["webhookSigningSecret"]; ok2 {
				m["webhook_secret"] = v
			} else if v, ok2 := m["signing_secret"]; ok2 {
				m["webhook_secret"] = v
			} else if v, ok2 := m["appkey"]; ok2 {
				m["webhook_secret"] = v
			}
		}
		if _, ok := m["success_url"]; !ok {
			if v, ok2 := m["successUrl"]; ok2 {
				m["success_url"] = v
			}
		}
		if _, ok := m["cancel_url"]; !ok {
			if v, ok2 := m["cancelUrl"]; ok2 {
				m["cancel_url"] = v
			}
		}
		if _, ok := m["payment_method_types"]; !ok {
			if v, ok2 := m["paymentMethods"]; ok2 {
				m["payment_method_types"] = v
			}
		}
		if _, ok := m["currency"]; !ok {
			if v, ok2 := m["currency_code"]; ok2 {
				m["currency"] = v
			}
		}
		if _, ok := m["currency_rate"]; !ok {
			m["currency_rate"] = "1"
		}
		b, _ := json.Marshal(m)
		configJSON = b
	}

	var cfg StripeConfig
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, err
	}

	cfg.SecretKey = strings.TrimSpace(cfg.SecretKey)
	cfg.PublishableKey = strings.TrimSpace(cfg.PublishableKey)
	cfg.WebhookSecret = strings.TrimSpace(cfg.WebhookSecret)
	cfg.SuccessURL = strings.TrimSpace(cfg.SuccessURL)
	cfg.CancelURL = strings.TrimSpace(cfg.CancelURL)
	cfg.Currency = normalizeStripeCurrency(cfg.Currency)
	if cfg.Currency == "" {
		cfg.Currency = defaultStripeCurrency
	}
	cfg.PaymentMethodTypes = strings.TrimSpace(cfg.PaymentMethodTypes)
	if cfg.PaymentMethodTypes == "" {
		cfg.PaymentMethodTypes = defaultStripePayMethods
	}

	rate := decimal.NewFromInt(1)
	if strings.TrimSpace(cfg.CurrencyRate) != "" {
		parsed, err := decimal.NewFromString(strings.TrimSpace(cfg.CurrencyRate))
		if err != nil {
			return nil, errors.New("stripe currency_rate is invalid")
		}
		if parsed.LessThanOrEqual(decimal.Zero) {
			return nil, errors.New("stripe currency_rate must be greater than zero")
		}
		rate = parsed
	}

	if cfg.SecretKey == "" {
		return nil, errors.New("stripe secret_key is required")
	}

	return &StripeAdapter{
		client:       &http.Client{Timeout: 30 * time.Second},
		config:       &cfg,
		currencyRate: rate,
	}, nil
}

// CreateOrder 创建支付订单
func (s *StripeAdapter) CreateOrder(ctx context.Context, req *CreateOrderRequest) (*CreateOrderResponse, error) {
	route := resolveStripeRoute(req.PayType, req.PayMethod)
	switch route.Kind {
	case stripeRouteKindDirect:
		return s.createDirectPaymentIntent(ctx, req, route)
	case stripeRouteKindCheckout:
		return s.createCheckoutSession(ctx, req, route)
	default:
		return nil, errors.New("unsupported stripe route")
	}
}

// QueryOrder 查询订单
func (s *StripeAdapter) QueryOrder(ctx context.Context, tradeNo string) (*QueryOrderResponse, error) {
	intent, err := s.findPaymentIntentByTradeNo(ctx, tradeNo)
	if err != nil {
		return nil, err
	}

	amountMinor := intent.AmountReceived
	if amountMinor == 0 {
		amountMinor = intent.Amount
	}

	paidAt := ""
	if intent.Status == "succeeded" && intent.Created > 0 {
		paidAt = time.Unix(intent.Created, 0).UTC().Format(time.RFC3339)
	}

	return &QueryOrderResponse{
		TradeNo:    tradeNo,
		ApiTradeNo: intent.ID,
		Amount:     s.fromStripeAmount(amountMinor, intent.Currency),
		Status:     mapStripeIntentStatus(intent.Status),
		PaidAt:     paidAt,
	}, nil
}

// Refund 退款
func (s *StripeAdapter) Refund(ctx context.Context, req *RefundRequest) (*RefundResponse, error) {
	intent, err := s.findPaymentIntentByTradeNo(ctx, req.TradeNo)
	if err != nil {
		return nil, err
	}

	amountMinor, err := s.toStripeAmount(req.Amount, intent.Currency)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("payment_intent", intent.ID)
	form.Set("amount", strconv.FormatInt(amountMinor, 10))
	form.Set("metadata[trade_no]", req.TradeNo)
	form.Set("metadata[refund_no]", req.RefundNo)
	if req.RefundDesc != "" {
		form.Set("metadata[refund_desc]", req.RefundDesc)
	}

	var refund stripeRefund
	if err := s.doRequest(ctx, http.MethodPost, "/refunds", nil, form, &refund); err != nil {
		return nil, err
	}

	status := "processing"
	switch refund.Status {
	case "succeeded":
		status = "success"
	case "failed", "canceled":
		status = "failed"
	}

	return &RefundResponse{
		RefundNo:     req.RefundNo,
		ApiRefundNo:  refund.ID,
		Status:       status,
		ErrorMessage: refund.FailureReason,
	}, nil
}

// ParseNotify 解析异步回调通知
func (s *StripeAdapter) ParseNotify(ctx context.Context, r *http.Request) (*NotifyResult, error) {
	_ = ctx

	if strings.TrimSpace(s.config.WebhookSecret) == "" {
		return nil, errors.New("stripe webhook_secret is required")
	}

	payload, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}

	if err := verifyStripeWebhookSignature(s.config.WebhookSecret, r.Header.Get("Stripe-Signature"), payload, stripeWebhookTolerance); err != nil {
		return nil, err
	}

	var event stripeWebhookEvent
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}

	switch event.Type {
	case "checkout.session.completed", "checkout.session.async_payment_succeeded":
		var session stripeCheckoutSession
		if err := json.Unmarshal(event.Data.Object, &session); err != nil {
			return nil, err
		}

		tradeNo := firstNonEmpty(session.ClientReferenceID, session.Metadata[stripeMetadataTradeNo])
		if tradeNo == "" {
			return nil, errors.New("stripe checkout session missing trade_no")
		}

		status := "ignored"
		if session.PaymentStatus == "paid" || event.Type == "checkout.session.async_payment_succeeded" {
			status = "success"
		}

		buyer := session.Customer
		if session.CustomerDetails != nil {
			buyer = firstNonEmpty(session.CustomerDetails.Email, session.CustomerDetails.Name, session.CustomerDetails.Phone, buyer)
		}

		return &NotifyResult{
			TradeNo:    tradeNo,
			ApiTradeNo: firstNonEmpty(session.PaymentIntent, session.ID),
			Amount:     s.fromStripeAmount(session.AmountTotal, session.Currency),
			Buyer:      buyer,
			Status:     status,
		}, nil

	case "payment_intent.succeeded", "payment_intent.payment_failed":
		var intent stripePaymentIntent
		if err := json.Unmarshal(event.Data.Object, &intent); err != nil {
			return nil, err
		}

		tradeNo := intent.Metadata[stripeMetadataTradeNo]
		if tradeNo == "" {
			return nil, errors.New("stripe payment intent missing trade_no")
		}

		status := "fail"
		if event.Type == "payment_intent.succeeded" || intent.Status == "succeeded" {
			status = "success"
		}

		amountMinor := intent.AmountReceived
		if amountMinor == 0 {
			amountMinor = intent.Amount
		}

		return &NotifyResult{
			TradeNo:    tradeNo,
			ApiTradeNo: intent.ID,
			Amount:     s.fromStripeAmount(amountMinor, intent.Currency),
			Buyer:      firstNonEmpty(intent.ReceiptEmail, intent.Customer),
			Status:     status,
		}, nil

	default:
		return &NotifyResult{Status: "ignored"}, nil
	}
}

// NotifySuccess 返回回调成功响应
func (s *StripeAdapter) NotifySuccess() string {
	return "ok"
}

type stripeRouteKind string

const (
	stripeRouteKindDirect   stripeRouteKind = "direct"
	stripeRouteKindCheckout stripeRouteKind = "checkout"
)

type stripeRoute struct {
	Kind                stripeRouteKind
	PaymentMethod       string
	CheckoutMethodTypes []string
	ResponsePayType     string
}

func resolveStripeRoute(payType, payMethod string) stripeRoute {
	normalizedType := strings.ToLower(strings.TrimSpace(payType))
	normalizedMethod := normalizeStripePayMethod(payMethod)

	switch normalizedType {
	case "alipay":
		return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodAlipay, ResponsePayType: "redirect"}
	case "wxpay", "wechat":
		return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodWechatPay, ResponsePayType: stripeWechatRoutePayType(normalizedMethod)}
	case "paypal":
		return stripeRoute{Kind: stripeRouteKindCheckout, CheckoutMethodTypes: []string{stripeCheckoutMethodPaypal}, ResponsePayType: "redirect"}
	case "bank":
		return stripeRoute{Kind: stripeRouteKindCheckout, CheckoutMethodTypes: []string{stripeCheckoutMethodCard}, ResponsePayType: "redirect"}
	case "stripe":
		if normalizedMethod == stripeDirectMethodAlipay {
			return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodAlipay, ResponsePayType: "redirect"}
		}
		if normalizedMethod == stripeDirectMethodWechatPay {
			return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodWechatPay, ResponsePayType: "qrcode"}
		}
		return stripeRoute{Kind: stripeRouteKindCheckout, CheckoutMethodTypes: checkoutMethodTypesFromConfig(""), ResponsePayType: "redirect"}
	default:
		if normalizedMethod == stripeDirectMethodAlipay {
			return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodAlipay, ResponsePayType: "redirect"}
		}
		if normalizedMethod == stripeDirectMethodWechatPay {
			return stripeRoute{Kind: stripeRouteKindDirect, PaymentMethod: stripeDirectMethodWechatPay, ResponsePayType: "qrcode"}
		}
		if normalizedMethod == stripeCheckoutMethodPaypal {
			return stripeRoute{Kind: stripeRouteKindCheckout, CheckoutMethodTypes: []string{stripeCheckoutMethodPaypal}, ResponsePayType: "redirect"}
		}
		return stripeRoute{Kind: stripeRouteKindCheckout, CheckoutMethodTypes: checkoutMethodTypesFromConfig(normalizedMethod), ResponsePayType: "redirect"}
	}
}

func stripeWechatRoutePayType(payMethod string) string {
	switch payMethod {
	case "h5", "web":
		return "redirect"
	default:
		return "qrcode"
	}
}

func (s *StripeAdapter) createCheckoutSession(ctx context.Context, req *CreateOrderRequest, route stripeRoute) (*CreateOrderResponse, error) {
	successURL, cancelURL, err := s.resolveCheckoutURLs(req)
	if err != nil {
		return nil, err
	}

	unitAmount, err := s.toStripeAmount(req.Amount, s.config.Currency)
	if err != nil {
		return nil, err
	}

	form := url.Values{}
	form.Set("mode", "payment")
	form.Set("success_url", successURL)
	form.Set("cancel_url", cancelURL)
	form.Set("client_reference_id", req.TradeNo)
	form.Set("metadata[trade_no]", req.TradeNo)
	form.Set("metadata[subject]", req.Subject)
	form.Set("payment_intent_data[metadata][trade_no]", req.TradeNo)
	form.Set("payment_intent_data[metadata][subject]", req.Subject)
	form.Set("line_items[0][price_data][currency]", s.config.Currency)
	form.Set("line_items[0][price_data][product_data][name]", req.Subject)
	form.Set("line_items[0][price_data][unit_amount]", strconv.FormatInt(unitAmount, 10))
	form.Set("line_items[0][quantity]", "1")

	if req.NotifyURL != "" {
		form.Set("payment_intent_data[metadata][notify_url]", req.NotifyURL)
	}
	if customerEmail := strings.TrimSpace(req.Extra["customer_email"]); customerEmail != "" {
		form.Set("customer_email", customerEmail)
	}

	methods := route.CheckoutMethodTypes
	if len(methods) == 0 {
		methods = checkoutMethodTypesFromConfig(s.config.PaymentMethodTypes)
	}
	for i, method := range methods {
		form.Set(fmt.Sprintf("payment_method_types[%d]", i), method)
	}

	var session stripeCheckoutSession
	if err := s.doRequest(ctx, http.MethodPost, "/checkout/sessions", nil, form, &session); err != nil {
		return nil, err
	}
	if session.URL == "" {
		return nil, errors.New("stripe checkout session missing redirect url")
	}

	return &CreateOrderResponse{
		PayType: "redirect",
		PayURL:  session.URL,
	}, nil
}

func (s *StripeAdapter) createDirectPaymentIntent(ctx context.Context, req *CreateOrderRequest, route stripeRoute) (*CreateOrderResponse, error) {
	amountMinor, err := s.toStripeAmount(req.Amount, s.config.Currency)
	if err != nil {
		return nil, err
	}

	returnURL := req.ReturnURL
	if returnURL == "" {
		returnURL = s.config.SuccessURL
	}

	form := url.Values{}
	form.Set("amount", strconv.FormatInt(amountMinor, 10))
	form.Set("currency", s.config.Currency)
	form.Set("confirm", "true")
	form.Set("payment_method_types[0]", route.PaymentMethod)
	form.Set("description", req.Subject)
	form.Set("metadata[trade_no]", req.TradeNo)
	form.Set("metadata[subject]", req.Subject)
	if req.NotifyURL != "" {
		form.Set("metadata[notify_url]", req.NotifyURL)
	}
	if returnURL != "" {
		form.Set("return_url", returnURL)
	}

	if route.PaymentMethod == stripeDirectMethodWechatPay {
		form.Set("payment_method_options[wechat_pay][client]", stripeWechatClient(req.PayMethod))
	}

	var intent stripePaymentIntent
	if err := s.doRequest(ctx, http.MethodPost, "/payment_intents", nil, form, &intent); err != nil {
		return nil, err
	}

	if intent.Status == "succeeded" {
		return &CreateOrderResponse{PayType: "redirect", PayURL: returnURL}, nil
	}

	payURL, payType, err := extractStripeDirectPayTarget(intent, req.PayMethod, returnURL)
	if err != nil {
		return nil, err
	}

	return &CreateOrderResponse{
		PayType: payType,
		PayURL:  payURL,
	}, nil
}

func extractStripeDirectPayTarget(intent stripePaymentIntent, payMethod, fallbackReturnURL string) (string, string, error) {
	if intent.NextAction == nil {
		if intent.ClientSecret != "" {
			return intent.ClientSecret, "redirect", nil
		}
		return "", "", errors.New("stripe payment intent missing next_action")
	}

	if intent.NextAction.AlipayHandleRedirect != nil && intent.NextAction.AlipayHandleRedirect.URL != "" {
		return intent.NextAction.AlipayHandleRedirect.URL, "redirect", nil
	}
	if intent.NextAction.RedirectToURL != nil && intent.NextAction.RedirectToURL.URL != "" {
		return intent.NextAction.RedirectToURL.URL, "redirect", nil
	}
	if intent.NextAction.WechatPayDisplayQR != nil {
		if strings.TrimSpace(payMethod) == "h5" || strings.TrimSpace(payMethod) == "web" {
			if intent.NextAction.WechatPayDisplayQR.HostedInstructionsURL != "" {
				return intent.NextAction.WechatPayDisplayQR.HostedInstructionsURL, "redirect", nil
			}
		}
		if intent.NextAction.WechatPayDisplayQR.ImageDataURL != "" {
			return intent.NextAction.WechatPayDisplayQR.ImageDataURL, "qrcode", nil
		}
		if intent.NextAction.WechatPayDisplayQR.Data != "" {
			return intent.NextAction.WechatPayDisplayQR.Data, "qrcode", nil
		}
	}
	if fallbackReturnURL != "" {
		return fallbackReturnURL, "redirect", nil
	}
	return "", "", errors.New("stripe payment intent missing payable target")
}

func stripeWechatClient(payMethod string) string {
	switch strings.ToLower(strings.TrimSpace(payMethod)) {
	case "h5", "web":
		return "web"
	default:
		return "web"
	}
}

func (s *StripeAdapter) resolveCheckoutURLs(req *CreateOrderRequest) (string, string, error) {
	successURL := firstNonEmpty(
		strings.TrimSpace(req.Extra["success_url"]),
		strings.TrimSpace(req.ReturnURL),
		s.config.SuccessURL,
	)
	if !isAbsoluteHTTPURL(successURL) {
		return "", "", errors.New("stripe success_url is required and must be an absolute http/https url")
	}

	cancelURL := firstNonEmpty(
		strings.TrimSpace(req.Extra["cancel_url"]),
		s.config.CancelURL,
		successURL,
	)
	if !isAbsoluteHTTPURL(cancelURL) {
		return "", "", errors.New("stripe cancel_url must be an absolute http/https url")
	}

	return successURL, cancelURL, nil
}

func (s *StripeAdapter) findPaymentIntentByTradeNo(ctx context.Context, tradeNo string) (*stripePaymentIntent, error) {
	query := url.Values{}
	query.Set("limit", "1")
	query.Set("query", fmt.Sprintf("metadata['trade_no']:'%s'", escapeStripeSearchValue(tradeNo)))

	var resp stripePaymentIntentSearchResponse
	if err := s.doRequest(ctx, http.MethodGet, "/payment_intents/search", query, nil, &resp); err != nil {
		return nil, err
	}
	if len(resp.Data) == 0 {
		return nil, errors.New("stripe payment intent not found")
	}

	return &resp.Data[0], nil
}

func (s *StripeAdapter) doRequest(ctx context.Context, method, path string, query url.Values, form url.Values, out interface{}) error {
	endpoint := stripeAPIBaseURL + path
	if len(query) > 0 {
		endpoint += "?" + query.Encode()
	}

	var body io.Reader = http.NoBody
	if form != nil {
		body = strings.NewReader(form.Encode())
	}

	req, err := http.NewRequestWithContext(ctx, method, endpoint, body)
	if err != nil {
		return err
	}
	req.Header.Set("Authorization", "Bearer "+s.config.SecretKey)
	req.Header.Set("Accept", "application/json")
	if form != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	resp, err := s.client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()

	bodyBytes, err := io.ReadAll(resp.Body)
	if err != nil {
		return err
	}

	if resp.StatusCode >= 400 {
		var apiErr stripeAPIErrorResponse
		if json.Unmarshal(bodyBytes, &apiErr) == nil && apiErr.Error != nil && apiErr.Error.Message != "" {
			message := apiErr.Error.Message
			if apiErr.Error.Param != "" {
				message += " (param: " + apiErr.Error.Param + ")"
			}
			return fmt.Errorf("stripe api error: %s", message)
		}
		if len(bodyBytes) > 0 {
			return fmt.Errorf("stripe api error: status=%d body=%s", resp.StatusCode, strings.TrimSpace(string(bodyBytes)))
		}
		return fmt.Errorf("stripe api error: status=%d", resp.StatusCode)
	}

	if out == nil || len(bodyBytes) == 0 {
		return nil
	}

	if err := json.Unmarshal(bodyBytes, out); err != nil {
		return fmt.Errorf("stripe api response decode error: %w", err)
	}
	return nil
}

func normalizeStripeCurrency(currency string) string {
	return strings.ToLower(strings.TrimSpace(currency))
}

func normalizeStripePayMethod(payMethod string) string {
	switch strings.ToLower(strings.TrimSpace(payMethod)) {
	case "", "checkout", "card", "hosted_checkout":
		return "checkout"
	case "web":
		return "web"
	case "h5":
		return "h5"
	case "paypal":
		return "paypal"
	case "bank":
		return "bank"
	case "alipay":
		return stripeDirectMethodAlipay
	case "wxpay", "wechat", "wechat_pay", "native":
		return stripeDirectMethodWechatPay
	default:
		return strings.ToLower(strings.TrimSpace(payMethod))
	}
}

func checkoutMethodTypesFromConfig(raw string) []string {
	if strings.TrimSpace(raw) == "" {
		return []string{defaultStripePayMethods}
	}

	seen := make(map[string]struct{})
	values := make([]string, 0)
	for _, item := range strings.Split(raw, ",") {
		value := strings.ToLower(strings.TrimSpace(item))
		if value == "" {
			continue
		}
		switch value {
		case "bank":
			value = stripeCheckoutMethodCard
		case "checkout", "hosted_checkout":
			value = stripeCheckoutMethodCard
		case stripeDirectMethodAlipay, stripeDirectMethodWechatPay:
			continue
		}
		if _, ok := seen[value]; ok {
			continue
		}
		seen[value] = struct{}{}
		values = append(values, value)
	}
	if len(values) == 0 {
		return []string{defaultStripePayMethods}
	}
	return values
}

func stripeAmountToMinor(amount decimal.Decimal, currency string) (int64, error) {
	currency = normalizeStripeCurrency(currency)
	if currency == "" {
		currency = defaultStripeCurrency
	}

	if _, ok := stripeZeroDecimalCurrencies[currency]; ok {
		rounded := amount.Round(0)
		if !rounded.Equal(amount) {
			return 0, fmt.Errorf("%s is a zero-decimal currency and requires an integer amount", strings.ToUpper(currency))
		}
		return rounded.IntPart(), nil
	}

	return amount.Mul(decimal.NewFromInt(100)).Round(0).IntPart(), nil
}

func stripeMinorToAmount(amount int64, currency string) decimal.Decimal {
	currency = normalizeStripeCurrency(currency)
	value := decimal.NewFromInt(amount)
	if _, ok := stripeZeroDecimalCurrencies[currency]; ok {
		return value
	}
	return value.Div(decimal.NewFromInt(100))
}

func (s *StripeAdapter) toStripeAmount(amount decimal.Decimal, currency string) (int64, error) {
	converted := amount
	if !s.currencyRate.Equal(decimal.NewFromInt(1)) {
		converted = amount.Mul(s.currencyRate)
	}
	return stripeAmountToMinor(converted, currency)
}

func (s *StripeAdapter) fromStripeAmount(amount int64, currency string) decimal.Decimal {
	converted := stripeMinorToAmount(amount, currency)
	if s.currencyRate.Equal(decimal.NewFromInt(1)) {
		return converted
	}
	return converted.Div(s.currencyRate).Round(2)
}

func mapStripeIntentStatus(status string) string {
	switch status {
	case "succeeded":
		return "paid"
	case "canceled":
		return "closed"
	case "requires_payment_method", "requires_action", "requires_confirmation", "processing", "requires_capture":
		return "pending"
	default:
		return "pending"
	}
}

func verifyStripeWebhookSignature(secret, signatureHeader string, payload []byte, tolerance time.Duration) error {
	secret = strings.TrimSpace(secret)
	if secret == "" {
		return errors.New("missing stripe webhook secret")
	}

	timestamp, signatures, err := parseStripeSignatureHeader(signatureHeader)
	if err != nil {
		return err
	}

	now := time.Now()
	signedAt := time.Unix(timestamp, 0)
	if now.Sub(signedAt) > tolerance || signedAt.Sub(now) > tolerance {
		return errors.New("stripe webhook signature timestamp expired")
	}

	mac := hmac.New(sha256.New, []byte(secret))
	mac.Write([]byte(strconv.FormatInt(timestamp, 10)))
	mac.Write([]byte("."))
	mac.Write(payload)
	expected := hex.EncodeToString(mac.Sum(nil))

	for _, sig := range signatures {
		if hmac.Equal([]byte(expected), []byte(sig)) {
			return nil
		}
	}

	return errors.New("invalid stripe webhook signature")
}

func parseStripeSignatureHeader(signatureHeader string) (int64, []string, error) {
	parts := strings.Split(signatureHeader, ",")
	var timestamp int64
	signatures := make([]string, 0, 1)

	for _, part := range parts {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}

		key, value, ok := strings.Cut(part, "=")
		if !ok {
			continue
		}

		switch key {
		case "t":
			parsed, err := strconv.ParseInt(value, 10, 64)
			if err != nil {
				return 0, nil, errors.New("invalid stripe webhook timestamp")
			}
			timestamp = parsed
		case "v1":
			signatures = append(signatures, value)
		}
	}

	if timestamp == 0 {
		return 0, nil, errors.New("missing stripe webhook timestamp")
	}
	if len(signatures) == 0 {
		return 0, nil, errors.New("missing stripe webhook signature")
	}

	return timestamp, signatures, nil
}

func escapeStripeSearchValue(value string) string {
	return strings.ReplaceAll(value, "'", "\\'")
}

func isAbsoluteHTTPURL(raw string) bool {
	u, err := url.Parse(raw)
	if err != nil {
		return false
	}
	if u.Host == "" {
		return false
	}
	return u.Scheme == "http" || u.Scheme == "https"
}

func firstNonEmpty(values ...string) string {
	for _, value := range values {
		if strings.TrimSpace(value) != "" {
			return strings.TrimSpace(value)
		}
	}
	return ""
}

func init() {
	Register("stripe", NewStripeAdapter)
}
