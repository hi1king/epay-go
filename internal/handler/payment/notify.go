// internal/handler/payment/notify.go
package payment

import (
	"bytes"
	"encoding/json"
	"errors"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/example/epay-go/internal/model"
	intPayment "github.com/example/epay-go/internal/payment"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/internal/service"
	"github.com/gin-gonic/gin"
)

// HandleNotify 处理支付回调
func HandleNotify(c *gin.Context) {
	channelPlugin := c.Param("channel")

	channelRepo := repository.NewChannelRepository()
	orderService := service.NewOrderService()
	notifyService := service.NewNotifyService()

	// 获取通道配置
	var channel *model.Channel
	var err error
	if channelPlugin == "stripe" {
		channel, err = resolveStripeNotifyChannel(c, orderService)
		if err != nil {
			log.Printf("Resolve stripe channel failed: %v", err)
			c.String(http.StatusOK, "fail")
			return
		}
	} else {
		channel, err = channelRepo.GetByPluginAndPayType(channelPlugin, "")
		if err != nil {
			log.Printf("Channel not found: %s", channelPlugin)
			c.String(http.StatusOK, "fail")
			return
		}
	}

	// 创建适配器
	adapter, err := intPayment.NewAdapter(channel.Plugin, channel.Config)
	if err != nil {
		log.Printf("Create adapter failed: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	// 解析回调
	result, err := adapter.ParseNotify(c.Request.Context(), c.Request)
	if err != nil {
		log.Printf("Parse notify failed: %v", err)
		c.String(http.StatusOK, "fail")
		return
	}

	// 处理支付结果
	if result.Status == "success" {
		if err := orderService.ProcessPayNotify(result.TradeNo, result.ApiTradeNo, result.Buyer, result.Amount); err != nil {
			log.Printf("Process notify failed: %v", err)
			c.String(http.StatusOK, "fail")
			return
		}

		// 发送商户通知
		order, _ := orderService.GetByTradeNo(result.TradeNo)
		if order != nil && order.Status == model.OrderStatusPaid {
			go notifyService.SendNotify(order)
		}
	}

	// 返回成功响应
	c.String(http.StatusOK, adapter.NotifySuccess())
}

func resolveStripeNotifyChannel(c *gin.Context, orderService *service.OrderService) (*model.Channel, error) {
	payload, err := io.ReadAll(c.Request.Body)
	if err != nil {
		return nil, err
	}
	c.Request.Body = io.NopCloser(bytes.NewReader(payload))

	var event struct {
		Data struct {
			Object json.RawMessage `json:"object"`
		} `json:"data"`
	}
	if err := json.Unmarshal(payload, &event); err != nil {
		return nil, err
	}
	if len(event.Data.Object) == 0 {
		return nil, errors.New("stripe webhook object is empty")
	}

	var object struct {
		ClientReferenceID string            `json:"client_reference_id"`
		Metadata          map[string]string `json:"metadata"`
	}
	if err := json.Unmarshal(event.Data.Object, &object); err != nil {
		return nil, err
	}

	tradeNo := strings.TrimSpace(object.ClientReferenceID)
	if tradeNo == "" {
		tradeNo = strings.TrimSpace(object.Metadata["trade_no"])
	}
	if tradeNo == "" {
		return nil, errors.New("stripe webhook missing trade_no")
	}

	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		return nil, err
	}
	if order.Channel == nil {
		return nil, errors.New("stripe webhook order channel not found")
	}
	if order.Channel.Plugin != "stripe" {
		return nil, errors.New("stripe webhook channel plugin mismatch")
	}

	return order.Channel, nil
}

// HandleReturn 处理同步跳转
func HandleReturn(c *gin.Context) {
	// 从参数获取订单号
	tradeNo := c.Query("out_trade_no")
	if tradeNo == "" {
		// 尝试从 body 读取
		body, _ := io.ReadAll(c.Request.Body)
		log.Printf("Return body: %s", string(body))
	}

	orderService := service.NewOrderService()
	order, err := orderService.GetByTradeNo(tradeNo)
	if err != nil {
		c.Redirect(http.StatusFound, "/")
		return
	}

	// 跳转到商户 return_url
	if order.ReturnURL != "" {
		c.Redirect(http.StatusFound, order.ReturnURL)
		return
	}

	c.String(http.StatusOK, "支付完成")
}
