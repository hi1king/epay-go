// internal/api/admin/channel.go
package admin

import (
	"strconv"

	payment "github.com/example/epay-go/internal/plugin"
	"github.com/example/epay-go/internal/service"
	"github.com/example/epay-go/pkg/response"
	"github.com/gin-gonic/gin"
)

// ListChannels 通道列表
func ListChannels(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	pageSize, _ := strconv.Atoi(c.DefaultQuery("page_size", "20"))

	channelService := service.NewChannelService()
	channels, total, err := channelService.List(page, pageSize)
	if err != nil {
		response.ServerError(c, "获取通道列表失败")
		return
	}

	response.SuccessPage(c, channels, total, page, pageSize)
}

// GetChannel 获取通道详情
func GetChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	channelService := service.NewChannelService()
	channel, err := channelService.GetByID(id)
	if err != nil {
		response.NotFound(c, "通道不存在")
		return
	}

	response.Success(c, channel)
}

// CreateChannel 创建通道
func CreateChannel(c *gin.Context) {
	var req service.CreateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	channelService := service.NewChannelService()
	channel, err := channelService.Create(&req)
	if err != nil {
		response.ServerError(c, "创建通道失败: "+err.Error())
		return
	}

	response.Success(c, channel)
}

// UpdateChannel 更新通道
func UpdateChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	var req service.UpdateChannelRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ParamError(c, "参数错误: "+err.Error())
		return
	}

	channelService := service.NewChannelService()
	if err := channelService.Update(id, &req); err != nil {
		response.ServerError(c, "更新通道失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

// DeleteChannel 删除通道
func DeleteChannel(c *gin.Context) {
	id, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ParamError(c, "无效的通道ID")
		return
	}

	channelService := service.NewChannelService()
	if err := channelService.Delete(id); err != nil {
		response.ServerError(c, "删除通道失败: "+err.Error())
		return
	}

	response.Success(c, nil)
}

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
