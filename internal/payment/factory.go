// internal/payment/factory.go
package payment

import (
	"encoding/json"
	"fmt"
)

// AdapterFactory 适配器工厂函数类型
type AdapterFactory func(config json.RawMessage) (PaymentAdapter, error)

// 注册的适配器工厂
var adapters = make(map[string]AdapterFactory)

// Register 注册适配器
func Register(plugin string, factory AdapterFactory) {
	adapters[plugin] = factory
}

// NewAdapter 创建适配器实例
func NewAdapter(plugin string, config json.RawMessage) (PaymentAdapter, error) {
	factory, ok := adapters[plugin]
	if !ok {
		return nil, fmt.Errorf("unsupported payment plugin: %s", plugin)
	}
	return factory(config)
}

// GetSupportedPlugins 获取支持的插件列表
func GetSupportedPlugins() []string {
	plugins := make([]string, 0, len(adapters))
	for plugin := range adapters {
		plugins = append(plugins, plugin)
	}
	return plugins
}
