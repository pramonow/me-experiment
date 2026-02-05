package handler

import (
	"matching-engine/usecase"
	"net/http"

	"github.com/gin-gonic/gin"
)

type OrderHandler struct {
	UseCase *usecase.OrderUseCase
}

func NewOrderHandler(uc *usecase.OrderUseCase) *OrderHandler {
	return &OrderHandler{
		UseCase: uc,
	}
}

type PlaceOrderRequest struct {
	UserID string  `json:"user_id" binding:"required"`
	Side   string  `json:"side" binding:"required,oneof=buy sell"`
	Type   string  `json:"type" binding:"required,oneof=limit market"`
	Size   float64 `json:"size" binding:"required,gt=0"`
	Price  float64 `json:"price"` // Optional for market orders
}

type OrderResponse struct {
	OrderID uint64      `json:"order_id"`
	Trades  interface{} `json:"trades"` // Use interface{} to avoid direct dependency on orderbook types if desired, but here we just pass it through
}

func (h *OrderHandler) RegisterRoutes(r *gin.Engine) {
	r.GET("/orderbook", h.GetOrderBook)
	r.POST("/orders", h.PlaceOrder)
	r.DELETE("/orderbook", h.ClearOrderBook)

	// Account Routes
	r.POST("/accounts", h.CreateAccount)
	r.POST("/accounts/deposit", h.Deposit)
	r.GET("/accounts/:user_id", h.GetAccount)
}

// ... existing methods ...

func (h *OrderHandler) CreateAccount(c *gin.Context) {
	var req struct {
		UserID string `json:"user_id" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	h.UseCase.AccountManager.CreateAccount(req.UserID)
	c.JSON(http.StatusOK, gin.H{"message": "account created"})
}

func (h *OrderHandler) Deposit(c *gin.Context) {
	var req struct {
		UserID   string  `json:"user_id" binding:"required"`
		Currency string  `json:"currency" binding:"required"`
		Amount   float64 `json:"amount" binding:"required,gt=0"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	if err := h.UseCase.AccountManager.Deposit(req.UserID, req.Currency, req.Amount); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "deposit successful"})
}

func (h *OrderHandler) GetAccount(c *gin.Context) {
	userID := c.Param("user_id")
	acc, err := h.UseCase.AccountManager.GetAccount(userID)
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, acc)
}

func (h *OrderHandler) GetOrderBook(c *gin.Context) {
	ob := h.UseCase.GetOrderBook()
	c.JSON(http.StatusOK, ob)
}

func (h *OrderHandler) PlaceOrder(c *gin.Context) {
	var req PlaceOrderRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	trades, order, err := h.UseCase.PlaceOrder(req.UserID, req.Side, req.Type, req.Size, req.Price)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	var orderID uint64
	if order != nil {
		orderID = order.ID
	}

	resp := OrderResponse{
		OrderID: orderID,
		Trades:  trades,
	}

	c.JSON(http.StatusOK, resp)
}

func (h *OrderHandler) ClearOrderBook(c *gin.Context) {
	if err := h.UseCase.ClearOrderBook(); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "failed to clear"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "orderbook cleared"})
}
