package http

import (
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/google/uuid"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/application/commands"
	"github.com/kayotklimenko/gozon-go/orders-service/internal/application/queries"
)

type Handlers struct {
	createOrderHandler *commands.CreateOrderHandler
	getOrderHandler    *queries.GetOrderHandler
	getOrdersHandler   *queries.GetOrdersHandler
}

func NewHandlers(
	createOrderHandler *commands.CreateOrderHandler,
	getOrderHandler *queries.GetOrderHandler,
	getOrdersHandler *queries.GetOrdersHandler,
) *Handlers {
	return &Handlers{
		createOrderHandler: createOrderHandler,
		getOrderHandler:    getOrderHandler,
		getOrdersHandler:   getOrdersHandler,
	}
}

type CreateOrderRequest struct {
	Amount      float64 `json:"amount" binding:"required,gt=0"`
	Description string  `json:"description" binding:"required,max=500"`
}

func (h *Handlers) CreateOrder(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	var req CreateOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	cmd := commands.CreateOrderCommand{
		UserID:      userID,
		Amount:      req.Amount,
		Description: req.Description,
	}

	result, err := h.createOrderHandler.Handle(c.Request.Context(), cmd)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusCreated, result)
}

func (h *Handlers) GetOrders(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	query := queries.GetOrdersQuery{UserID: userID}
	result, err := h.getOrdersHandler.Handle(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	c.JSON(http.StatusOK, result)
}

func (h *Handlers) GetOrder(c *gin.Context) {
	userID := c.GetHeader("X-User-Id")
	if userID == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "X-User-Id header is required"})
		return
	}

	if _, err := uuid.Parse(userID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid User ID format"})
		return
	}

	orderID := c.Param("id")
	if _, err := uuid.Parse(orderID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "Invalid Order ID format"})
		return
	}

	query := queries.GetOrderQuery{
		OrderID: orderID,
		UserID:  userID,
	}

	result, err := h.getOrderHandler.Handle(c.Request.Context(), query)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "Order not found"})
		return
	}

	c.JSON(http.StatusOK, result)
}
