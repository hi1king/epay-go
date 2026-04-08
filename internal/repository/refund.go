// internal/repository/refund.go
package repository

import (
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type RefundRepository struct {
	db *gorm.DB
}

func NewRefundRepository() *RefundRepository {
	return &RefundRepository{db: GetDB()}
}

// Create 创建退款单
func (r *RefundRepository) Create(refund *model.Refund) error {
	return r.db.Create(refund).Error
}

// GetByRefundNo 根据退款单号查询
func (r *RefundRepository) GetByRefundNo(refundNo string) (*model.Refund, error) {
	var refund model.Refund
	err := r.db.Where("refund_no = ?", refundNo).First(&refund).Error
	return &refund, err
}

// GetByTradeNo 根据订单号查询退款记录
func (r *RefundRepository) GetByTradeNo(tradeNo string) ([]*model.Refund, error) {
	var refunds []*model.Refund
	err := r.db.Where("trade_no = ?", tradeNo).Find(&refunds).Error
	return refunds, err
}

// List 分页查询退款列表
func (r *RefundRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]*model.Refund, int64, error) {
	var refunds []*model.Refund
	var total int64

	query := r.db.Model(&model.Refund{})

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
	err = query.Order("id DESC").Limit(pageSize).Offset(offset).Find(&refunds).Error
	return refunds, total, err
}

// Update 更新退款单
func (r *RefundRepository) Update(refund *model.Refund) error {
	return r.db.Save(refund).Error
}
