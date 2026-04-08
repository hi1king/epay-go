// internal/model/system_config.go
package model

import "time"

// SystemConfig 系统配置
type SystemConfig struct {
	Key       string    `gorm:"size:64;primarykey" json:"key"`
	Value     string    `gorm:"type:text" json:"value"`
	UpdatedAt time.Time `json:"updated_at"`
}

func (SystemConfig) TableName() string {
	return "system_configs"
}
