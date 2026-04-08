// internal/repository/withdraw_repo.go
package repository

import (
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type WithdrawRepository struct {
	db *gorm.DB
}

func NewWithdrawRepository() *WithdrawRepository {
	return &WithdrawRepository{db: GetDB()}
}

// Create 创建提现记录
func (r *WithdrawRepository) Create(withdraw *model.Withdraw) error {
	return r.db.Create(withdraw).Error
}

// GetByID 根据ID获取提现记录
func (r *WithdrawRepository) GetByID(id int64) (*model.Withdraw, error) {
	var withdraw model.Withdraw
	err := r.db.Preload("Merchant").First(&withdraw, id).Error
	if err != nil {
		return nil, err
	}
	return &withdraw, nil
}

// GetByWithdrawNo 根据提现单号获取记录
func (r *WithdrawRepository) GetByWithdrawNo(withdrawNo string) (*model.Withdraw, error) {
	var withdraw model.Withdraw
	err := r.db.Where("withdraw_no = ?", withdrawNo).First(&withdraw).Error
	if err != nil {
		return nil, err
	}
	return &withdraw, nil
}

// Update 更新提现记录
func (r *WithdrawRepository) Update(withdraw *model.Withdraw) error {
	return r.db.Save(withdraw).Error
}

// UpdateStatus 更新提现状态
func (r *WithdrawRepository) UpdateStatus(id int64, status int8, remark string) error {
	updates := map[string]interface{}{
		"status": status,
		"remark": remark,
	}
	return r.db.Model(&model.Withdraw{}).Where("id = ?", id).Updates(updates).Error
}

// List 分页查询提现列表
func (r *WithdrawRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Withdraw, int64, error) {
	var withdraws []model.Withdraw
	var total int64

	query := r.db.Model(&model.Withdraw{})
	if merchantID != nil {
		query = query.Where("merchant_id = ?", *merchantID)
	}
	if status != nil {
		query = query.Where("status = ?", *status)
	}

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Preload("Merchant").Offset(offset).Limit(pageSize).Order("id DESC").Find(&withdraws).Error
	if err != nil {
		return nil, 0, err
	}

	return withdraws, total, nil
}

// HasPendingWithdraw 检查是否有待处理的提现
func (r *WithdrawRepository) HasPendingWithdraw(merchantID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Withdraw{}).
		Where("merchant_id = ? AND status IN ?", merchantID, []int8{model.WithdrawStatusPending, model.WithdrawStatusProcessing}).
		Count(&count).Error
	return count > 0, err
}
