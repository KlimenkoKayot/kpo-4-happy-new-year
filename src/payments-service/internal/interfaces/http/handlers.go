package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/application/commands"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/application/queries"
	"github.com/kayotklimenko/gozon-go/payments-service/internal/domain/account"
)

type Handlers struct {
	createAccountHandler *commands.CreateAccountHandler
	depositHandler       *commands.DepositHandler
	getBalanceHandler    *queries.GetBalanceHandler
}

func NewHandlers(
	createAccountHandler *commands.CreateAccountHandler,
	depositHandler *commands.DepositHandler,
	getBalanceHandler *queries.GetBalanceHandler,
) *Handlers {
	return &Handlers{
		createAccountHandler: createAccountHandler,
		depositHandler:       depositHandler,
		getBalanceHandler:    getBalanceHandler,
	}
}

func (h *Handlers) CreateAccount(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	cmd := commands.CreateAccountCommand{UserID: userID}
	result, err := h.createAccountHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		if err == account.ErrAccountExists {
			c.JSON(http.StatusConflict, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

type DepositRequest struct {
	Amount float64 `json:"amount" binding:"required,gt=0"`
}

func (h *Handlers) Deposit(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	var req DepositRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := commands.DepositCommand{
		UserID: userID,
		Amount: req.Amount,
	}

	result, err := h.depositHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		if err == account.ErrAccountNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handlers) GetBalance(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	query := queries.GetBalanceQuery{UserID: userID}
	result, err := h.getBalanceHandler.Handle(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Account not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}
