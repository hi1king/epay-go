// internal/repository/settlement.go
package repository

import (
	"github.com/example/epay-go/internal/model"
	"gorm.io/gorm"
)

type SettlementRepository struct {
	db *gorm.DB
}

func NewSettlementRepository() *SettlementRepository {
	return &SettlementRepository{db: GetDB()}
}

// Create 创建结算记录
func (r *SettlementRepository) Create(settlement *model.Settlement) error {
	return r.db.Create(settlement).Error
}

// GetByID 根据ID获取结算记录
func (r *SettlementRepository) GetByID(id int64) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.Preload("Merchant").First(&settlement, id).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// GetBySettleNo 根据结算单号获取记录
func (r *SettlementRepository) GetBySettleNo(settleNo string) (*model.Settlement, error) {
	var settlement model.Settlement
	err := r.db.Where("settle_no = ?", settleNo).First(&settlement).Error
	if err != nil {
		return nil, err
	}
	return &settlement, nil
}

// Update 更新结算记录
func (r *SettlementRepository) Update(settlement *model.Settlement) error {
	return r.db.Save(settlement).Error
}

// UpdateStatus 更新结算状态
func (r *SettlementRepository) UpdateStatus(id int64, status int8, remark string) error {
	updates := map[string]interface{}{
		"status": status,
		"remark": remark,
	}
	return r.db.Model(&model.Settlement{}).Where("id = ?", id).Updates(updates).Error
}

// List 分页查询结算列表
func (r *SettlementRepository) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Settlement, int64, error) {
	var settlements []model.Settlement
	var total int64

	query := r.db.Model(&model.Settlement{})
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
	err = query.Preload("Merchant").Offset(offset).Limit(pageSize).Order("id DESC").Find(&settlements).Error
	if err != nil {
		return nil, 0, err
	}

	return settlements, total, nil
}

// HasPendingSettlement 检查是否有待处理的结算
func (r *SettlementRepository) HasPendingSettlement(merchantID int64) (bool, error) {
	var count int64
	err := r.db.Model(&model.Settlement{}).
		Where("merchant_id = ? AND status IN ?", merchantID, []int8{model.SettleStatusPending, model.SettleStatusProcessing}).
		Count(&count).Error
	return count > 0, err
}
