package transaction

import (
	"time"

	"github.com/google/uuid"
)

type TransactionID string
type TransactionType int

const (
	TypeDeposit TransactionType = iota
	TypeWithdrawal
)

type Transaction struct {
	id          TransactionID
	accountID   string
	orderID     *string
	amount      float64
	txType      TransactionType
	description string
	createdAt   time.Time
}

func NewDepositTransaction(accountID string, amount float64) *Transaction {
	return &Transaction{
		id:          TransactionID(uuid.New().String()),
		accountID:   accountID,
		amount:      amount,
		txType:      TypeDeposit,
		description: "Deposit",
		createdAt:   time.Now().UTC(),
	}
}

func NewPaymentTransaction(accountID, orderID string, amount float64) *Transaction {
	return &Transaction{
		id:          TransactionID(uuid.New().String()),
		accountID:   accountID,
		orderID:     &orderID,
		amount:      amount,
		txType:      TypeWithdrawal,
		description: "Payment for order " + orderID,
		createdAt:   time.Now().UTC(),
	}
}

func (t *Transaction) ID() TransactionID     { return t.id }
func (t *Transaction) AccountID() string     { return t.accountID }
func (t *Transaction) OrderID() *string      { return t.orderID }
func (t *Transaction) Amount() float64       { return t.amount }
func (t *Transaction) Type() TransactionType { return t.txType }
func (t *Transaction) Description() string   { return t.description }
func (t *Transaction) CreatedAt() time.Time  { return t.createdAt }
