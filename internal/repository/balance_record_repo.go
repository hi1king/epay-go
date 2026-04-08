// internal/repository/balance_record_repo.go
package repository

import (
	"github.com/example/epay-go/internal/model"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type BalanceRecordRepository struct {
	db *gorm.DB
}

func NewBalanceRecordRepository() *BalanceRecordRepository {
	return &BalanceRecordRepository{db: GetDB()}
}

// Create 创建资金记录
func (r *BalanceRecordRepository) Create(record *model.BalanceRecord) error {
	return r.db.Create(record).Error
}

// CreateWithTx 在事务中创建资金记录
func (r *BalanceRecordRepository) CreateWithTx(tx *gorm.DB, record *model.BalanceRecord) error {
	return tx.Create(record).Error
}

// List 分页查询资金记录
func (r *BalanceRecordRepository) List(page, pageSize int, merchantID int64) ([]model.BalanceRecord, int64, error) {
	var records []model.BalanceRecord
	var total int64

	query := r.db.Model(&model.BalanceRecord{}).Where("merchant_id = ?", merchantID)

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

// AddBalanceRecord 添加资金变动记录 (带事务)
func AddBalanceRecord(tx *gorm.DB, merchantID int64, action int8, amount decimal.Decimal, beforeBalance, afterBalance decimal.Decimal, recordType, tradeNo string) error {
	record := &model.BalanceRecord{
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
