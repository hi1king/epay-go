// internal/model/admin_user.go
package model

import "time"

// AdminUser 管理员
type AdminUser struct {
	BaseModel
	Username    string     `gorm:"size:64;uniqueIndex;not null" json:"username"`
	Password    string     `gorm:"size:128;not null" json:"-"`
	Role        string     `gorm:"size:20;default:admin" json:"role"` // super, admin
	LastLoginAt *time.Time `json:"last_login_at"`
}

func (AdminUser) TableName() string {
	return "admin_users"
}
