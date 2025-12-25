package commands

import (
	"context"
	"encoding/json"
	"log"
	"time"

	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/inbox"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/outbox"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/transaction"
)

type ProcessPaymentCommand struct {
	OrderID        string
	UserID         string
	Amount         float64
	IdempotencyKey string
}

type PaymentResultMessage struct {
	OrderID      string    `json:"order_id"`
	Success      bool      `json:"success"`
	ErrorMessage string    `json:"error_message,omitempty"`
	ProcessedAt  time.Time `json:"processed_at"`
}

type ProcessPaymentHandler struct {
	accountRepo     account.Repository
	transactionRepo transaction.Repository
	inboxRepo       inbox.Repository
	outboxRepo      outbox.Repository
	paymentService  *account.PaymentService
	unitOfWork      UnitOfWork
}

func NewProcessPaymentHandler(
	accountRepo account.Repository,
	transactionRepo transaction.Repository,
	inboxRepo inbox.Repository,
	outboxRepo outbox.Repository,
	uow UnitOfWork,
) *ProcessPaymentHandler {
	return &ProcessPaymentHandler{
		accountRepo:     accountRepo,
		transactionRepo: transactionRepo,
		inboxRepo:       inboxRepo,
		outboxRepo:      outboxRepo,
		paymentService:  account.NewPaymentService(),
		unitOfWork:      uow,
	}
}

func (h *ProcessPaymentHandler) Handle(ctx context.Context, cmd ProcessPaymentCommand) error {
	existingInbox, _ := h.inboxRepo.FindByIdempotencyKey(ctx, cmd.IdempotencyKey)
	if existingInbox != nil && existingInbox.IsProcessed() {
		log.Printf("Payment already processed for order %s", cmd.OrderID)
		return nil
	}

	exists, _ := h.transactionRepo.ExistsByOrderID(ctx, cmd.OrderID)
	if exists {
		log.Printf("Transaction already exists for order %s", cmd.OrderID)
		return nil
	}

	txCtx, err := h.unitOfWork.Begin(ctx)
	if err != nil {
		return err
	}

	if existingInbox == nil {
		inboxMsg := inbox.NewMessage(cmd.IdempotencyKey, "PaymentRequest", "")
		if err := h.inboxRepo.Save(txCtx, inboxMsg); err != nil {
			h.unitOfWork.Rollback(txCtx)
			return err
		}
		existingInbox = inboxMsg
	}

	acc, err := h.accountRepo.FindByUserID(txCtx, account.UserID(cmd.UserID))

	var result PaymentResultMessage
	result.OrderID = cmd.OrderID
	result.ProcessedAt = time.Now().UTC()

	money := account.Money(cmd.Amount)

	if err != nil || acc == nil {
		log.Printf("Account not found for user %s", cmd.UserID)
		result.Success = false
		result.ErrorMessage = "Account not found"
	} else if err := h.paymentService.CanProcessPayment(acc, money); err != nil {
		log.Printf("Cannot process payment for user %s: %v", cmd.UserID, err)
		result.Success = false
		result.ErrorMessage = err.Error()
	} else {
		if err := acc.Withdraw(money); err != nil {
			result.Success = false
			result.ErrorMessage = err.Error()
		} else {
			if err := h.accountRepo.Update(txCtx, acc); err != nil {
				h.unitOfWork.Rollback(txCtx)
				return err
			}

			tx := transaction.NewPaymentTransaction(string(acc.ID()), cmd.OrderID, cmd.Amount)
			if err := h.transactionRepo.Save(txCtx, tx); err != nil {
				h.unitOfWork.Rollback(txCtx)
				return err
			}

			result.Success = true
			log.Printf("Payment processed for order %s, amount %.2f", cmd.OrderID, cmd.Amount)
		}
	}

	existingInbox.MarkAsProcessed()
	if err := h.inboxRepo.Update(txCtx, existingInbox); err != nil {
		log.Printf("Warning: failed to update inbox: %v", err)
	}

	payload, err := json.Marshal(result)
	if err != nil {
		h.unitOfWork.Rollback(txCtx)
		return err
	}

	outboxMsg := outbox.NewMessage("PaymentResultMessage", string(payload))
	if err := h.outboxRepo.Save(txCtx, outboxMsg); err != nil {
		h.unitOfWork.Rollback(txCtx)
		return err
	}

	return h.unitOfWork.Commit(txCtx)
}
