// internal/service/admin_user_service.go
package service

import (
	"errors"
	"os"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"gorm.io/gorm"
)

type AdminUserService struct {
	repo *repository.AdminUserRepository
}

func NewAdminUserService() *AdminUserService {
	return &AdminUserService{
		repo: repository.NewAdminUserRepository(),
	}
}

// InitDefaultAdmin 初始化默认管理员（如果不存在）
func (s *AdminUserService) InitDefaultAdmin() error {
	count, err := s.repo.Count()
	if err != nil {
		return err
	}

	if count > 0 {
		return nil // 已有管理员，跳过
	}

	// 创建默认管理员
	username := os.Getenv("DEFAULT_ADMIN_USERNAME")
	if username == "" {
		username = "epay_admin"
	}

	password := os.Getenv("DEFAULT_ADMIN_PASSWORD")
	if password == "" {
		password = "ChangeMe123!"
	}

	hashedPassword, err := utils.HashPassword(password)
	if err != nil {
		return err
	}

	admin := &model.AdminUser{
		Username: username,
		Password: hashedPassword,
		Role:     "super",
	}

	return s.repo.Create(admin)
}

// AdminUserLoginRequest 管理员登录请求
type AdminUserLoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Login 管理员登录
func (s *AdminUserService) Login(req *AdminUserLoginRequest) (*model.AdminUser, error) {
	admin, err := s.repo.GetByUsername(req.Username)
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, errors.New("用户名或密码错误")
		}
		return nil, err
	}

	if !utils.CheckPassword(req.Password, admin.Password) {
		return nil, errors.New("用户名或密码错误")
	}

	// 更新最后登录时间
	_ = s.repo.UpdateLastLogin(admin.ID)

	return admin, nil
}

// GetByID 根据ID获取管理员
func (s *AdminUserService) GetByID(id int64) (*model.AdminUser, error) {
	return s.repo.GetByID(id)
}

// UpdatePassword 更新密码
func (s *AdminUserService) UpdatePassword(id int64, oldPassword, newPassword string) error {
	admin, err := s.repo.GetByID(id)
	if err != nil {
		return err
	}

	if !utils.CheckPassword(oldPassword, admin.Password) {
		return errors.New("原密码错误")
	}

	hashedPassword, err := utils.HashPassword(newPassword)
	if err != nil {
		return err
	}

	admin.Password = hashedPassword
	return s.repo.Update(admin)
}
