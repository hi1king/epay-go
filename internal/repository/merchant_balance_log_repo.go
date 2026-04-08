// internal/repository/merchant_balance_log_repo.go
package repository

import (
	"github.com/example/epay-go/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type MerchantBalanceLogRepository struct {
	db *gorm.DB
}

func NewMerchantBalanceLogRepository() *MerchantBalanceLogRepository {
	return &MerchantBalanceLogRepository{db: GetDB()}
}

// Create 创建资金记录
func (r *MerchantBalanceLogRepository) Create(record *model.MerchantBalanceLog) error {
	return r.db.Create(record).Error
}

// CreateWithTx 在事务中创建资金记录
func (r *MerchantBalanceLogRepository) CreateWithTx(tx *gorm.DB, record *model.MerchantBalanceLog) error {
	return tx.Create(record).Error
}

// List 分页查询资金记录
func (r *MerchantBalanceLogRepository) List(page, pageSize int, merchantID int64) ([]model.MerchantBalanceLog, int64, error) {
	var records []model.MerchantBalanceLog
	var total int64

	query := r.db.Model(&model.MerchantBalanceLog{}).Where("merchant_id = ?", merchantID)

	err := query.Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	offset := (page - 1) * pageSize
	err = query.Offset(offset).Limit(pageSize).Order("id DESC").Find(&records).Error
	if err != nil {
		return nil, 0, err
	}

	return records, total, nil
}

// AddMerchantBalanceLog 添加资金变动记录 (带事务)
func AddMerchantBalanceLog(tx *gorm.DB, merchantID int64, action int8, amount decimal.Decimal, beforeBalance, afterBalance decimal.Decimal, recordType, tradeNo string) error {
	record := &model.MerchantBalanceLog{
		MerchantID:    merchantID,
		Action:        action,
		Amount:        amount,
		BeforeBalance: beforeBalance,
		AfterBalance:  afterBalance,
		Type:          recordType,
		TradeNo:       tradeNo,
	}
	return tx.Create(record).Error
}
