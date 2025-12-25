package account

import "errors"

var (
	ErrAccountNotFound     = errors.New("account not found")
	ErrAccountExists       = errors.New("account already exists")
	ErrInvalidUserID       = errors.New("invalid user ID")
	ErrInvalidAmount       = errors.New("invalid amount")
	ErrInsufficientFunds   = errors.New("insufficient funds")
	ErrConcurrencyConflict = errors.New("concurrency conflict")
)
