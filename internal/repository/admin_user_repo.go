// internal/repository/admin_user_repo.go
package repository

import (
	"time"

	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type AdminUserRepository struct {
	db *gorm.DB
}

func NewAdminUserRepository() *AdminUserRepository {
	return &AdminUserRepository{db: GetDB()}
}

// Create 创建管理员
func (r *AdminUserRepository) Create(admin *model.AdminUser) error {
	return r.db.Create(admin).Error
}

func (r *AdminUserRepository) Update(admin *model.AdminUser) error {
	return r.db.Save(admin).Error
}

// GetByID 根据ID获取管理员
func (r *AdminUserRepository) GetByID(id int64) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *AdminUserRepository) GetByUsername(username string) (*model.AdminUser, error) {
	var admin model.AdminUser
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminUserRepository) UpdateLastLogin(id int64) error {
	now := time.Now()
	return r.db.Model(&model.AdminUser{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *AdminUserRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.AdminUser{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Count 统计管理员数量
func (r *AdminUserRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.AdminUser{}).Count(&count).Error
	return count, err
}
