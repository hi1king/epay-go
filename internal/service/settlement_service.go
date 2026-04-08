// internal/service/settlement_service.go
package service

import (
	"errors"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type SettlementService struct {
	settleRepo   *repository.SettlementRepository
	merchantRepo *repository.MerchantRepository
}

func NewSettlementService() *SettlementService {
	return &SettlementService{
		settleRepo:   repository.NewSettlementRepository(),
		merchantRepo: repository.NewMerchantRepository(),
	}
}

// ApplyRequest 申请结算请求
type ApplyRequest struct {
	MerchantID  int64           `json:"-"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	AccountType string          `json:"account_type" binding:"required,oneof=alipay bank"`
	AccountNo   string          `json:"account_no" binding:"required"`
	AccountName string          `json:"account_name" binding:"required"`
}

// Apply 申请结算
func (s *SettlementService) Apply(req *ApplyRequest) (*model.Settlement, error) {
	// 检查是否有待处理的结算
	hasPending, err := s.settleRepo.HasPendingSettlement(req.MerchantID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("您有待处理的结算申请，请等待处理完成")
	}

	// 获取商户信息
	merchant, err := s.merchantRepo.GetByID(req.MerchantID)
	if err != nil {
		return nil, err
	}

	// 检查余额是否充足
	if merchant.Balance.LessThan(req.Amount) {
		return nil, errors.New("余额不足")
	}

	// 最小结算金额检查（假设最小10元）
	minAmount := decimal.NewFromInt(10)
	if req.Amount.LessThan(minAmount) {
		return nil, errors.New("最小结算金额为10元")
	}

	// 计算手续费（假设2%）
	feeRate := decimal.NewFromFloat(0.02)
	fee := req.Amount.Mul(feeRate).Round(2)
	actualAmount := req.Amount.Sub(fee)

	// 创建结算记录
	settlement := &model.Settlement{
		SettleNo:     utils.GenerateSettleNo(),
		MerchantID:   req.MerchantID,
		Amount:       req.Amount,
		Fee:          fee,
		ActualAmount: actualAmount,
		AccountType:  req.AccountType,
		AccountNo:    req.AccountNo,
		AccountName:  req.AccountName,
		Status:       model.SettleStatusPending,
	}

	// 开启事务
	tx := repository.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 冻结余额
	newBalance := merchant.Balance.Sub(req.Amount)
	newFrozen := merchant.FrozenBalance.Add(req.Amount)

	if err := tx.Model(&model.Merchant{}).Where("id = ?", req.MerchantID).Updates(map[string]interface{}{
		"balance":        newBalance,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	// 添加资金记录
	if err := repository.AddBalanceRecord(tx, req.MerchantID, model.RecordActionExpense, req.Amount, merchant.Balance, newBalance, "settle_freeze", settlement.SettleNo); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建结算记录
	if err := tx.Create(settlement).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return settlement, nil
}

// Approve 审核通过
func (s *SettlementService) Approve(id int64) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusPending {
		return errors.New("结算状态不正确")
	}

	// 更新状态为处理中
	return s.settleRepo.UpdateStatus(id, model.SettleStatusProcessing, "审核通过，处理中")
}

// Complete 完成结算
func (s *SettlementService) Complete(id int64) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusProcessing {
		return errors.New("结算状态不正确")
	}

	// 开启事务
	tx := repository.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 扣除冻结余额
	if err := tx.Model(&model.Merchant{}).Where("id = ?", settlement.MerchantID).
		Update("frozen_balance", gorm.Expr("frozen_balance - ?", settlement.Amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新结算状态
	if err := s.settleRepo.UpdateStatus(id, model.SettleStatusCompleted, "结算完成"); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// Reject 驳回结算
func (s *SettlementService) Reject(id int64, remark string) error {
	settlement, err := s.settleRepo.GetByID(id)
	if err != nil {
		return err
	}

	if settlement.Status != model.SettleStatusPending {
		return errors.New("结算状态不正确")
	}

	// 开启事务
	tx := repository.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取商户
	merchant, err := s.merchantRepo.GetByID(settlement.MerchantID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 解冻余额
	newBalance := merchant.Balance.Add(settlement.Amount)
	newFrozen := merchant.FrozenBalance.Sub(settlement.Amount)

	if err := tx.Model(&model.Merchant{}).Where("id = ?", settlement.MerchantID).Updates(map[string]interface{}{
		"balance":        newBalance,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加资金记录
	if err := repository.AddBalanceRecord(tx, settlement.MerchantID, model.RecordActionIncome, settlement.Amount, merchant.Balance, newBalance, "settle_unfreeze", settlement.SettleNo); err != nil {
		tx.Rollback()
		return err
	}

	// 更新结算状态
	if err := s.settleRepo.UpdateStatus(id, model.SettleStatusRejected, remark); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// List 分页查询结算列表
func (s *SettlementService) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Settlement, int64, error) {
	return s.settleRepo.List(page, pageSize, merchantID, status)
}

// GetByID 根据ID获取结算记录
func (s *SettlementService) GetByID(id int64) (*model.Settlement, error) {
	return s.settleRepo.GetByID(id)
}
