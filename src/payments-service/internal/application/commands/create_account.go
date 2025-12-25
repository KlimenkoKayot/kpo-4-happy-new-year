package commands

import (
	"context"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
)

type CreateAccountCommand struct {
	UserID string
}

type CreateAccountResult struct {
	ID        string  `json:"id"`
	UserID    string  `json:"user_id"`
	Balance   float64 `json:"balance"`
	CreatedAt string  `json:"created_at"`
}

type CreateAccountHandler struct {
	accountRepo account.Repository
}

func NewCreateAccountHandler(accountRepo account.Repository) *CreateAccountHandler {
	return &CreateAccountHandler{accountRepo: accountRepo}
}

func (h *CreateAccountHandler) Handle(ctx context.Context, cmd CreateAccountCommand) (*CreateAccountResult, error) {
	existing, _ := h.accountRepo.FindByUserID(ctx, account.UserID(cmd.UserID))
	if existing != nil {
		return nil, account.ErrAccountExists
	}

	acc, err := account.NewAccount(cmd.UserID)
	if err != nil {
		return nil, err
	}

	if err := h.accountRepo.Save(ctx, acc); err != nil {
		return nil, err
	}

	return &CreateAccountResult{
		ID:        string(acc.ID()),
		UserID:    string(acc.UserID()),
		Balance:   acc.Balance().Float64(),
		CreatedAt: acc.CreatedAt().Format("2006-01-02T15:04:05Z"),
	}, nil
}
