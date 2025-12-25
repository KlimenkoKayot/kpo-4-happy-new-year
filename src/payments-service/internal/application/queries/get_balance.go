package queries

import (
	"context"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
)

type GetBalanceQuery struct {
	UserID string
}

type AccountDTO struct {
	ID        string     `json:"id"`
	UserID    string     `json:"user_id"`
	Balance   float64    `json:"balance"`
	CreatedAt time.Time  `json:"created_at"`
	UpdatedAt *time.Time `json:"updated_at,omitempty"`
}

type GetBalanceHandler struct {
	accountRepo account.Repository
}

func NewGetBalanceHandler(accountRepo account.Repository) *GetBalanceHandler {
	return &GetBalanceHandler{accountRepo: accountRepo}
}

func (h *GetBalanceHandler) Handle(ctx context.Context, query GetBalanceQuery) (*AccountDTO, error) {
	acc, err := h.accountRepo.FindByUserID(ctx, account.UserID(query.UserID))
	if err != nil {
		return nil, account.ErrAccountNotFound
	}

	return &AccountDTO{
		ID:        string(acc.ID()),
		UserID:    string(acc.UserID()),
		Balance:   acc.Balance().Float64(),
		CreatedAt: acc.CreatedAt(),
		UpdatedAt: acc.UpdatedAt(),
	}, nil
}
