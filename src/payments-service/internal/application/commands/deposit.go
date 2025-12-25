package commands

import (
	"context"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/transaction"
)

type DepositCommand struct {
	UserID string
	Amount float64
}

type DepositResult struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type UnitOfWork interface {
	Begin(ctx context.Context) (context.Context, error)
	Commit(ctx context.Context) error
	Rollback(ctx context.Context) error
}

type DepositHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
	unitOfWork      UnitOfWork
}

func NewDepositHandler(accountRepo account.Repository, transactionRepo transaction.Repository, uow UnitOfWork) *DepositHandler {
	return &DepositHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		unitOfWork:      uow,
	}
}

func (h *DepositHandler) Handle(ctx context.Context, cmd DepositCommand) (*DepositResult, error) {
	acc, err := h.accountRepo.FindByUserID(ctx, account.UserID(cmd.UserID))
	if err != nil {
		return nil, account.ErrAccountNotFound
	}

	money, err := account.NewMoney(cmd.Amount)
	if err != nil {
		return nil, err
	}

	txCtx, err := h.unitOfWork.Begin(ctx)
	if err != nil {
		return nil, err
	}

	if err := acc.Deposit(money); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return nil, err
	}

	if err := h.accountRepo.Update(txCtx, acc); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return nil, err
	}

	tx := transaction.NewDepositTransaction(string(acc.ID()), cmd.Amount)
	if err := h.transactionRepo.Save(txCtx, tx); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return nil, err
	}

	if err := h.unitOfWork.Commit(txCtx); err != nil {
		return nil, err
	}

	return &DepositResult{
		ID:        string(acc.ID()),
		UserID:    string(acc.UserID()),
		Balance:   acc.Balance().Float64(),
		CreatedAt: acc.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}
