// internal/service/withdraw_service.go
package service

import (
	"errors"

	"github.com/example/epay-go/internal/model"
	"github.com/example/epay-go/internal/repository"
	"github.com/example/epay-go/pkg/utils"
	"github.com/shopspring/decimal"
	"gorm.io/gorm"
)

type WithdrawService struct {
	withdrawRepo *repository.WithdrawRepository
	merchantRepo *repository.MerchantRepository
}

func NewWithdrawService() *WithdrawService {
	return &WithdrawService{
		withdrawRepo: repository.NewWithdrawRepository(),
		merchantRepo: repository.NewMerchantRepository(),
	}
}

// ApplyWithdrawRequest 申请提现请求
type ApplyWithdrawRequest struct {
	MerchantID  int64           `json:"-"`
	Amount      decimal.Decimal `json:"amount" binding:"required"`
	AccountType string          `json:"account_type" binding:"required,oneof=alipay bank"`
	AccountNo   string          `json:"account_no" binding:"required"`
	AccountName string          `json:"account_name" binding:"required"`
}

// Apply 申请提现
func (s *WithdrawService) Apply(req *ApplyWithdrawRequest) (*model.Withdraw, error) {
	// 检查是否有待处理的提现
	hasPending, err := s.withdrawRepo.HasPendingWithdraw(req.MerchantID)
	if err != nil {
		return nil, err
	}
	if hasPending {
		return nil, errors.New("您有待处理的提现申请，请等待处理完成")
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

	// 最小提现金额检查（假设最小10元）
	minAmount := decimal.NewFromInt(10)
	if req.Amount.LessThan(minAmount) {
		return nil, errors.New("最小提现金额为10元")
	}

	// 计算手续费（假设2%）
	feeRate := decimal.NewFromFloat(0.02)
	fee := req.Amount.Mul(feeRate).Round(2)
	actualAmount := req.Amount.Sub(fee)

	// 创建提现记录
	withdraw := &model.Withdraw{
		WithdrawNo:   utils.GenerateWithdrawNo(),
		MerchantID:   req.MerchantID,
		Amount:       req.Amount,
		Fee:          fee,
		ActualAmount: actualAmount,
		AccountType:  req.AccountType,
		AccountNo:    req.AccountNo,
		AccountName:  req.AccountName,
		Status:       model.WithdrawStatusPending,
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
	if err := repository.AddMerchantBalanceLog(tx, req.MerchantID, model.BalanceActionExpense, req.Amount, merchant.Balance, newBalance, "withdraw_freeze", withdraw.WithdrawNo); err != nil {
		tx.Rollback()
		return nil, err
	}

	// 创建提现记录
	if err := tx.Create(withdraw).Error; err != nil {
		tx.Rollback()
		return nil, err
	}

	tx.Commit()
	return withdraw, nil
}

// Approve 审核通过
func (s *WithdrawService) Approve(id int64) error {
	withdraw, err := s.withdrawRepo.GetByID(id)
	if err != nil {
		return err
	}

	if withdraw.Status != model.WithdrawStatusPending {
		return errors.New("提现状态不正确")
	}

	// 更新状态为处理中
	return s.withdrawRepo.UpdateStatus(id, model.WithdrawStatusProcessing, "审核通过，处理中")
}

// Complete 完成提现
func (s *WithdrawService) Complete(id int64) error {
	withdraw, err := s.withdrawRepo.GetByID(id)
	if err != nil {
		return err
	}

	if withdraw.Status != model.WithdrawStatusProcessing {
		return errors.New("提现状态不正确")
	}

	// 开启事务
	tx := repository.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 扣除冻结余额
	if err := tx.Model(&model.Merchant{}).Where("id = ?", withdraw.MerchantID).
		Update("frozen_balance", gorm.Expr("frozen_balance - ?", withdraw.Amount)).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 更新提现状态
	if err := s.withdrawRepo.UpdateStatus(id, model.WithdrawStatusCompleted, "提现完成"); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// Reject 驳回提现
func (s *WithdrawService) Reject(id int64, remark string) error {
	withdraw, err := s.withdrawRepo.GetByID(id)
	if err != nil {
		return err
	}

	if withdraw.Status != model.WithdrawStatusPending {
		return errors.New("提现状态不正确")
	}

	// 开启事务
	tx := repository.GetDB().Begin()
	defer func() {
		if r := recover(); r != nil {
			tx.Rollback()
		}
	}()

	// 获取商户
	merchant, err := s.merchantRepo.GetByID(withdraw.MerchantID)
	if err != nil {
		tx.Rollback()
		return err
	}

	// 解冻余额
	newBalance := merchant.Balance.Add(withdraw.Amount)
	newFrozen := merchant.FrozenBalance.Sub(withdraw.Amount)

	if err := tx.Model(&model.Merchant{}).Where("id = ?", withdraw.MerchantID).Updates(map[string]interface{}{
		"balance":        newBalance,
		"frozen_balance": newFrozen,
	}).Error; err != nil {
		tx.Rollback()
		return err
	}

	// 添加资金记录
	if err := repository.AddMerchantBalanceLog(tx, withdraw.MerchantID, model.BalanceActionIncome, withdraw.Amount, merchant.Balance, newBalance, "withdraw_unfreeze", withdraw.WithdrawNo); err != nil {
		tx.Rollback()
		return err
	}

	// 更新提现状态
	if err := s.withdrawRepo.UpdateStatus(id, model.WithdrawStatusRejected, remark); err != nil {
		tx.Rollback()
		return err
	}

	tx.Commit()
	return nil
}

// List 分页查询提现列表
func (s *WithdrawService) List(page, pageSize int, merchantID *int64, status *int8) ([]model.Withdraw, int64, error) {
	return s.withdrawRepo.List(page, pageSize, merchantID, status)
}

// GetByID 根据ID获取提现记录
func (s *WithdrawService) GetByID(id int64) (*model.Withdraw, error) {
	return s.withdrawRepo.GetByID(id)
}
