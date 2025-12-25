package account

import (
	"time"

	"github.com/google/uuid"
)

type AccountID string
type UserID string

type Account struct {
	id        AccountID
	userID    UserID
	balance   Money
	version   int
	createdAt time.Time
	updatedAt *time.Time
}

func NewAccount(userID string) (*Account, error) {
	if userID == "" {
		return nil, ErrInvalidUserID
	}

	return &Account{
		id:        AccountID(uuid.New().String()),
		userID:    UserID(userID),
		balance:   Money(0),
		version:   1,
		createdAt: time.Now().UTC(),
	}, nil
}

func ReconstructAccount(id, userID string, balance float64, version int, createdAt time.Time, updatedAt *time.Time) *Account {
	return &Account{
		id:        AccountID(id),
		userID:    UserID(userID),
		balance:   Money(balance),
		version:   version,
		createdAt: createdAt,
		updatedAt: updatedAt,
	}
}

func (a *Account) Deposit(amount Money) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	a.balance = a.balance.Add(amount)
	a.touch()
	return nil
}

func (a *Account) Withdraw(amount Money) error {
	if amount <= 0 {
		return ErrInvalidAmount
	}

	if !a.balance.CanSubtract(amount) {
		return ErrInsufficientFunds
	}

	a.balance = a.balance.Subtract(amount)
	a.touch()
	return nil
}

func (a *Account) CanWithdraw(amount Money) bool {
	return a.balance.CanSubtract(amount)
}

func (a *Account) touch() {
	now := time.Now().UTC()
	a.updatedAt = &now
	a.version++
}

func (a *Account) ID() AccountID         { return a.id }
func (a *Account) UserID() UserID        { return a.userID }
func (a *Account) Balance() Money        { return a.balance }
func (a *Account) Version() int          { return a.version }
func (a *Account) CreatedAt() time.Time  { return a.createdAt }
func (a *Account) UpdatedAt() *time.Time { return a.updatedAt }
