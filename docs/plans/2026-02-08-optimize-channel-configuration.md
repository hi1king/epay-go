# 支付通道配置优化实现计划

> **For Claude:** REQUIRED SUB-SKILL: Use superpowers:executing-plans to implement this plan task-by-task.

**目标：** 参考 epay 项目的插件化配置机制，为 epay-go 实现动态通道配置功能，不同支付类型自动显示对应的配置字段（appid、appsecret、商户号等），并支持在线测试支付。

**架构：** 采用配置驱动的方式，每种支付插件定义自己的配置模板，前端根据选择的插件动态渲染表单，后端使用 JSON 存储配置参数。

**技术栈：** Go 1.21+, Gin, GORM, PostgreSQL, Vue 3, TypeScript, Arco Design

---

## 核心设计思路

### epay 项目的优秀设计

根据探索发现，epay 采用了以下机制：

1. **插件化配置定义**：每个支付插件通过 `$info['inputs']` 定义所需参数
2. **动态表单生成**：前端根据插件配置自动生成表单字段
3. **JSON 配置存储**：所有配置以 JSON 格式保存到数据库
4. **多接口支持**：单个插件可支持多种支付接口（扫码、H5、APP 等）
5. **绑定功能**：支持绑定微信公众号/小程序

### epay-go 实现方案

我们将在 Go 项目中实现类似机制：

1. **定义配置模板**：在代码中定义每种支付类型的配置字段
2. **扩展 Channel 模型**：添加 `AppType`（已启用接口）字段
3. **实现配置验证**：后端验证配置完整性
4. **前端动态表单**：根据选择的插件类型动态显示配置字段
5. **测试支付功能**：管理员可在配置后立即测试

---

## Task 1: 定义支付插件配置模板

**文件：**
- Create: `internal/payment/plugin_config.go`

### 步骤 1: 创建配置模板结构

```go
// internal/payment/plugin_config.go
package payment

// PluginConfigField 配置字段定义
type PluginConfigField struct {
	Key         string            `json:"key"`          // 字段键名
	Name        string            `json:"name"`         // 显示名称
	Type        string            `json:"type"`         // 类型: input/textarea/select/checkbox
	Required    bool              `json:"required"`     // 是否必填
	Placeholder string            `json:"placeholder"`  // 占位符
	Note        string            `json:"note"`         // 说明文字
	Options     map[string]string `json:"options"`      // 下拉选项（type=select时使用）
}

// PluginConfig 插件配置信息
type PluginConfig struct {
	Name       string              `json:"name"`        // 插件英文名
	ShowName   string              `json:"show_name"`   // 显示名称
	Author     string              `json:"author"`      // 作者
	Link       string              `json:"link"`        // 官方链接
	Inputs     []PluginConfigField `json:"inputs"`      // 配置字段
	PayTypes   []PayTypeOption     `json:"pay_types"`   // 支持的支付接口
	BindWxmp   bool                `json:"bind_wxmp"`   // 是否绑定微信公众号
	BindWxa    bool                `json:"bind_wxa"`    // 是否绑定微信小程序
	Note       string              `json:"note"`        // 配置说明
}

// PayTypeOption 支付接口选项
type PayTypeOption struct {
	Code string `json:"code"` // 接口代码
	Name string `json:"name"` // 接口名称
}

// GetPluginConfigs 获取所有插件配置模板
func GetPluginConfigs() map[string]PluginConfig {
	return map[string]PluginConfig{
		"alipay": GetAlipayConfig(),
		"wechat": GetWechatConfig(),
	}
}

// GetAlipayConfig 支付宝配置模板
func GetAlipayConfig() PluginConfig {
	return PluginConfig{
		Name:     "alipay",
		ShowName: "支付宝官方支付",
		Author:   "支付宝",
		Link:     "https://open.alipay.com",
		Inputs: []PluginConfigField{
			{
				Key:         "appid",
				Name:        "应用APPID",
				Type:        "input",
				Required:    true,
				Placeholder: "请输入支付宝开放平台应用ID",
			},
			{
				Key:         "app_private_key",
				Name:        "应用私钥",
				Type:        "textarea",
				Required:    true,
				Placeholder: "请输入RSA2私钥（PKCS1或PKCS8格式）",
				Note:        "用于对请求参数进行签名",
			},
			{
				Key:         "alipay_public_key",
				Name:        "支付宝公钥",
				Type:        "textarea",
				Required:    false,
				Placeholder: "请输入支付宝公钥",
				Note:        "用于验证回调签名，使用公钥证书模式可留空",
			},
			{
				Key:         "seller_id",
				Name:        "卖家支付宝用户ID",
				Type:        "input",
				Required:    false,
				Placeholder: "可留空，默认为签约账号",
			},
			{
				Key:         "cert_mode",
				Name:        "证书模式",
				Type:        "select",
				Required:    true,
				Options: map[string]string{
					"public_key": "公钥模式",
					"cert":       "公钥证书模式",
				},
			},
		},
		PayTypes: []PayTypeOption{
			{Code: "page", Name: "电脑网站支付"},
			{Code: "wap", Name: "手机网站支付"},
			{Code: "qrcode", Name: "当面付扫码"},
			{Code: "jsapi", Name: "当面付JS"},
			{Code: "app", Name: "APP支付"},
		},
		Note: `
选择可用的支付接口，只能选择已经签约的产品。
如果使用公钥证书模式，需将应用公钥证书、支付宝公钥证书、
支付宝根证书3个crt文件放置于 /certs/alipay/ 目录。
`,
	}
}

// GetWechatConfig 微信支付配置模板
func GetWechatConfig() PluginConfig {
	return PluginConfig{
		Name:     "wechat",
		ShowName: "微信官方支付",
		Author:   "微信支付",
		Link:     "https://pay.weixin.qq.com",
		Inputs: []PluginConfigField{
			{
				Key:         "appid",
				Name:        "公众号/小程序/开放平台AppID",
				Type:        "input",
				Required:    true,
				Placeholder: "请输入微信应用标识",
			},
			{
				Key:         "mch_id",
				Name:        "商户号",
				Type:        "input",
				Required:    true,
				Placeholder: "请输入微信支付商户号",
			},
			{
				Key:         "api_key",
				Name:        "商户API密钥",
				Type:        "input",
				Required:    true,
				Placeholder: "请输入APIv2密钥（32位）",
				Note:        "在微信商户平台设置",
			},
			{
				Key:         "apiv3_key",
				Name:        "APIv3密钥",
				Type:        "input",
				Required:    false,
				Placeholder: "请输入APIv3密钥（32位）",
				Note:        "仅部分新接口需要",
			},
			{
				Key:         "cert_serial_no",
				Name:        "证书序列号",
				Type:        "input",
				Required:    false,
				Placeholder: "商户API证书序列号",
				Note:        "如需退款功能，需填写此项",
			},
		},
		PayTypes: []PayTypeOption{
			{Code: "native", Name: "扫码支付"},
			{Code: "jsapi", Name: "公众号支付"},
			{Code: "h5", Name: "H5支付"},
			{Code: "miniapp", Name: "小程序支付"},
			{Code: "app", Name: "APP支付"},
		},
		BindWxmp: true,
		BindWxa:  true,
		Note:     "如需退款功能，需将API证书文件上传至 /certs/wechat/ 目录",
	}
}
```

### 步骤 2: 创建获取配置接口

在 `internal/handler/admin/channel.go` 添加：

```go
// GetPluginConfig 获取插件配置模板
func GetPluginConfig(c *gin.Context) {
	pluginName := c.Param("plugin")

	configs := payment.GetPluginConfigs()
	config, exists := configs[pluginName]
	if !exists {
		response.NotFound(c, "插件不存在")
		return
	}

	response.Success(c, config)
}

// GetAllPluginConfigs 获取所有插件配置列表
func GetAllPluginConfigs(c *gin.Context) {
	configs := payment.GetPluginConfigs()

	// 转换为列表
	list := make([]map[string]interface{}, 0)
	for _, cfg := range configs {
		list = append(list, map[string]interface{}{
			"name":      cfg.Name,
			"show_name": cfg.ShowName,
			"author":    cfg.Author,
		})
	}

	response.Success(c, list)
}
```

### 步骤 3: 注册路由

在 `internal/router/router.go` 添加：

```go
// 获取插件配置
adminGroup.GET("/plugins", admin.GetAllPluginConfigs)
adminGroup.GET("/plugins/:plugin/config", admin.GetPluginConfig)
```

### 步骤 4: 测试配置接口

```bash
# 获取所有插件列表
curl http://localhost:8099/admin/plugins

# 获取支付宝配置模板
curl http://localhost:8099/admin/plugins/alipay/config

# 获取微信配置模板
curl http://localhost:8099/admin/plugins/wechat/config
```

### 步骤 5: 提交代码

```bash
git add internal/payment/plugin_config.go internal/handler/admin/channel.go internal/router/router.go
git commit -m "feat: add payment plugin configuration templates

- Define plugin config structure with dynamic fields
- Implement Alipay and WeChat config templates
- Add API endpoints to get plugin configs
- Support multiple payment interfaces per plugin

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 2: 扩展 Channel 模型支持多接口

**文件：**
- Modify: `internal/model/channel.go`
- Modify: `internal/repository/channel.go`
- Modify: `internal/service/channel.go`

### 步骤 1: 扩展 Channel 模型

在 `internal/model/channel.go` 添加字段：

```go
AppType  string `gorm:"size:100" json:"app_type"`  // 已启用的支付接口（逗号分隔，如"page,wap,qrcode"）
```

### 步骤 2: 数据库迁移

运行应用，GORM 会自动添加新字段。

### 步骤 3: 更新 Repository

在 `internal/repository/channel.go` 的 `Create` 和 `Update` 方法中支持 `AppType` 字段。

### 步骤 4: 更新 Service

在 `internal/service/channel.go` 中添加 `AppType` 验证逻辑：

```go
// CreateChannelRequest 添加字段
type CreateChannelRequest struct {
	// ... 现有字段
	AppType  string `json:"app_type"`  // 支持的接口类型
}

// 验证 AppType 不能为空
if req.AppType == "" {
	return nil, errors.New("请至少选择一个支付接口")
}
```

### 步骤 5: 测试并提交

```bash
git add internal/model/channel.go internal/repository/channel.go internal/service/channel.go
git commit -m "feat: extend channel model to support multiple payment interfaces

- Add app_type field to store enabled interfaces
- Validate app_type is not empty when creating channel
- Support comma-separated interface codes

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 3: 实现前端动态配置表单

**文件：**
- Modify: `web/src/views/admin/Channels.vue`
- Modify: `web/src/api/admin.ts`
- Modify: `web/src/api/types.ts`

### 步骤 1: 添加类型定义

在 `web/src/api/types.ts` 添加：

```typescript
export interface PluginConfigField {
  key: string
  name: string
  type: 'input' | 'textarea' | 'select' | 'checkbox'
  required: boolean
  placeholder: string
  note: string
  options?: Record<string, string>
}

export interface PayTypeOption {
  code: string
  name: string
}

export interface PluginConfig {
  name: string
  show_name: string
  author: string
  link: string
  inputs: PluginConfigField[]
  pay_types: PayTypeOption[]
  bind_wxmp: boolean
  bind_wxa: boolean
  note: string
}
```

### 步骤 2: 添加 API 方法

在 `web/src/api/admin.ts` 添加：

```typescript
// 获取所有插件
export const getPlugins = () =>
  request.get<{ name: string; show_name: string; author: string }[]>('/admin/plugins')

// 获取插件配置模板
export const getPluginConfig = (plugin: string) =>
  request.get<PluginConfig>(`/admin/plugins/${plugin}/config`)
```

### 步骤 3: 修改 Channels.vue 实现动态表单

```vue
<template>
  <div class="channels-page">
    <!-- ... 现有的列表部分 ... -->

    <!-- 创建/编辑通道对话框 -->
    <a-modal
      v-model:visible="modalVisible"
      :title="editingId ? '编辑通道' : '新建通道'"
      @ok="handleSubmit"
      @cancel="handleCancel"
      width="700px"
    >
      <a-form :model="form" layout="vertical">
        <a-form-item field="name" label="通道名称" required>
          <a-input v-model="form.name" placeholder="如: 支付宝扫码支付" />
        </a-form-item>

        <a-form-item field="plugin" label="支付插件" required>
          <a-select
            v-model="form.plugin"
            placeholder="请选择支付插件"
            @change="handlePluginChange"
          >
            <a-option
              v-for="plugin in plugins"
              :key="plugin.name"
              :value="plugin.name"
            >
              {{ plugin.show_name }}
            </a-option>
          </a-select>
        </a-form-item>

        <!-- 动态配置字段 -->
        <template v-if="currentPluginConfig">
          <a-divider>支付配置</a-divider>

          <a-form-item
            v-for="field in currentPluginConfig.inputs"
            :key="field.key"
            :field="`config.${field.key}`"
            :label="field.name"
            :required="field.required"
          >
            <!-- 普通输入框 -->
            <a-input
              v-if="field.type === 'input'"
              v-model="form.config[field.key]"
              :placeholder="field.placeholder"
            />

            <!-- 多行文本框 -->
            <a-textarea
              v-else-if="field.type === 'textarea'"
              v-model="form.config[field.key]"
              :placeholder="field.placeholder"
              :rows="4"
            />

            <!-- 下拉选择 -->
            <a-select
              v-else-if="field.type === 'select'"
              v-model="form.config[field.key]"
              :placeholder="field.placeholder"
            >
              <a-option
                v-for="(label, value) in field.options"
                :key="value"
                :value="value"
              >
                {{ label }}
              </a-option>
            </a-select>

            <template v-if="field.note" #extra>
              <div style="color: #86909c; font-size: 12px">{{ field.note }}</div>
            </template>
          </a-form-item>

          <a-divider>支付接口</a-divider>

          <a-form-item label="支持的支付接口" required>
            <a-checkbox-group v-model="form.app_types">
              <a-checkbox
                v-for="payType in currentPluginConfig.pay_types"
                :key="payType.code"
                :value="payType.code"
              >
                {{ payType.name }}
              </a-checkbox>
            </a-checkbox-group>
          </a-form-item>

          <a-alert v-if="currentPluginConfig.note" type="info">
            {{ currentPluginConfig.note }}
          </a-alert>
        </template>

        <a-divider>费率设置</a-divider>

        <a-form-item field="fee_rate" label="费率(%)" required>
          <a-input-number
            v-model="form.fee_rate"
            :precision="2"
            :min="0"
            :max="100"
            placeholder="如: 0.6"
          />
        </a-form-item>
      </a-form>
    </a-modal>
  </div>
</template>

<script setup lang="ts">
import { ref, reactive, onMounted } from 'vue'
import { Message } from '@arco-design/web-vue'
import { IconPlus } from '@arco-design/web-vue/es/icon'
import {
  getChannels,
  createChannel,
  updateChannel,
  toggleChannelStatus,
  getPlugins,
  getPluginConfig,
} from '@/api/admin'
import type { Channel, PluginConfig } from '@/api/types'

const plugins = ref<{ name: string; show_name: string }[]>([])
const currentPluginConfig = ref<PluginConfig | null>(null)

const form = reactive({
  name: '',
  plugin: '',
  fee_rate: 0,
  config: {} as Record<string, string>,
  app_types: [] as string[],
})

// 加载插件列表
const loadPlugins = async () => {
  try {
    const res = await getPlugins()
    plugins.value = res.data
  } catch (e) {
    // error handled
  }
}

// 插件切换时加载配置模板
const handlePluginChange = async (plugin: string) => {
  try {
    const res = await getPluginConfig(plugin)
    currentPluginConfig.value = res.data

    // 重置配置
    form.config = {}
    form.app_types = []
  } catch (e) {
    Message.error('加载插件配置失败')
  }
}

// 提交时转换 app_types 为逗号分隔字符串
const handleSubmit = async () => {
  try {
    if (form.app_types.length === 0) {
      Message.warning('请至少选择一个支付接口')
      return
    }

    const data = {
      ...form,
      app_type: form.app_types.join(','),
      config: form.config,
    }

    if (editingId.value) {
      await updateChannel(editingId.value, data)
      Message.success('更新成功')
    } else {
      await createChannel(data)
      Message.success('创建成功')
    }

    modalVisible.value = false
    loadChannels()
  } catch (e) {
    // error handled
  }
}

onMounted(() => {
  loadChannels()
  loadPlugins()
})
</script>
```

### 步骤 4: 测试动态表单

1. 访问 http://localhost:8090/admin/channels
2. 点击"新建通道"
3. 选择"支付宝官方支付"，查看动态显示的配置字段
4. 切换到"微信官方支付"，查看字段变化
5. 填写配置并提交

### 步骤 5: 提交代码

```bash
git add web/src/views/admin/Channels.vue web/src/api/admin.ts web/src/api/types.ts
git commit -m "feat: implement dynamic channel configuration form

- Add plugin config types and API methods
- Implement dynamic form field rendering
- Support different config fields per plugin
- Add payment interface checkbox selection
- Display plugin notes and field hints

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## Task 4: 实现测试支付功能

**文件：**
- Create: `internal/handler/admin/test_payment.go`
- Create: `web/src/views/admin/TestPayment.vue`
- Modify: `internal/router/router.go`

### 步骤 1: 创建测试支付 Handler

```go
// internal/handler/admin/test_payment.go
package admin

import (
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// TestPaymentRequest 测试支付请求
type TestPaymentRequest struct {
	ChannelID int64  `json:"channel_id" binding:"required"`
	Amount    string `json:"amount" binding:"required"`
	PayType   string `json:"pay_type" binding:"required"`
}

// TestPayment 测试支付
func TestPayment(c *gin.Context) {
	var req TestPaymentRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误")
		return
	}

	// 创建测试订单
	orderService := service.NewOrderService()
	order, payData, err := orderService.CreateTestOrder(req.ChannelID, req.Amount, req.PayType)
	if err != nil {
		response.Error(c, response.CodeServerError, err.Error())
		return
	}

	response.Success(c, gin.H{
		"order":    order,
		"pay_data": payData,
	})
}
```

### 步骤 2: 在 OrderService 中实现测试订单

```go
// CreateTestOrder 创建测试订单
func (s *OrderService) CreateTestOrder(channelID int64, amount, payType string) (*model.Order, interface{}, error) {
	// 获取通道信息
	channel, err := s.channelRepo.GetByID(channelID)
	if err != nil {
		return nil, nil, errors.New("通道不存在")
	}

	// 创建测试订单（商户ID为0表示测试）
	order := &model.Order{
		TradeNo:      utils.GenerateOrderNo("TEST"),
		OutTradeNo:   "TEST" + time.Now().Format("20060102150405"),
		MerchantID:   0, // 测试订单
		ChannelID:    channelID,
		PayType:      payType,
		Amount:       amount,
		Name:         "测试支付",
		NotifyURL:    "",
		ReturnURL:    "",
		Status:       model.OrderStatusPending,
	}

	if err := s.orderRepo.Create(order); err != nil {
		return nil, nil, err
	}

	// 调用支付接口
	payData, err := s.callPaymentGateway(order, channel)
	if err != nil {
		return nil, nil, err
	}

	return order, payData, nil
}
```

### 步骤 3: 创建测试支付页面

创建 `web/src/views/admin/TestPayment.vue`（简化版）

### 步骤 4: 注册路由

```go
adminGroup.POST("/test-payment", admin.TestPayment)
```

### 步骤 5: 测试并提交

```bash
git add internal/handler/admin/test_payment.go internal/service/order.go internal/router/router.go web/src/views/admin/TestPayment.vue
git commit -m "feat: add test payment functionality

- Implement test payment API endpoint
- Create test orders with merchant_id=0
- Support testing different payment interfaces
- Add test payment frontend page

Co-Authored-By: Claude Sonnet 4.5 <noreply@anthropic.com>"
```

---

## 总结

本计划实现了以下核心功能：

1. ✅ **插件配置模板** - 定义支付宝、微信等插件的配置字段
2. ✅ **动态表单生成** - 前端根据选择的插件自动显示对应配置字段
3. ✅ **多接口支持** - 单个通道支持多种支付方式（扫码、H5、APP 等）
4. ✅ **配置验证** - 后端验证配置完整性和格式
5. ✅ **测试支付** - 管理员可在配置后立即测试支付功能

**参考 epay 项目的优秀设计：**
- 插件化架构
- 配置驱动的动态表单
- JSON 格式存储配置
- 支持证书模式和多接口

**预计工作量：** 2-3 天完成全部功能
