// internal/repository/admin.go
package repository

import (
	"time"

	"github.com/example/epay-go/internal/database"
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type AdminRepository struct {
	db *gorm.DB
}

func NewAdminRepository() *AdminRepository {
	return &AdminRepository{db: database.Get()}
}

// Create 创建管理员
func (r *AdminRepository) Create(admin *model.Admin) error {
	return r.db.Create(admin).Error
}

func (r *AdminRepository) Update(admin *model.Admin) error {
	return r.db.Save(admin).Error
}

// GetByID 根据ID获取管理员
func (r *AdminRepository) GetByID(id int64) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.First(&admin, id).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// GetByUsername 根据用户名获取管理员
func (r *AdminRepository) GetByUsername(username string) (*model.Admin, error) {
	var admin model.Admin
	err := r.db.Where("username = ?", username).First(&admin).Error
	if err != nil {
		return nil, err
	}
	return &admin, nil
}

// UpdateLastLogin 更新最后登录时间
func (r *AdminRepository) UpdateLastLogin(id int64) error {
	now := time.Now()
	return r.db.Model(&model.Admin{}).Where("id = ?", id).Update("last_login_at", &now).Error
}

// ExistsByUsername 检查用户名是否存在
func (r *AdminRepository) ExistsByUsername(username string) (bool, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Where("username = ?", username).Count(&count).Error
	return count > 0, err
}

// Count 统计管理员数量
func (r *AdminRepository) Count() (int64, error) {
	var count int64
	err := r.db.Model(&model.Admin{}).Count(&count).Error
	return count, err
}
