package main

import (
	"fmt"
	"matching-engine/handler"
	"matching-engine/usecase"

	"github.com/gin-gonic/gin"
)

const OrderBookFile = "orderbook.json"

func main() {
	// 1. Initialize UseCase
	orderUC := usecase.NewOrderUseCase(OrderBookFile)

	// 2. Initialize Handler
	orderHandler := handler.NewOrderHandler(orderUC)

	// 3. Setup Gin Router
	r := gin.Default()

	// 4. Register Routes
	orderHandler.RegisterRoutes(r)

	// Startup message
	fmt.Println("Matching Engine Server starting on port 8080...")

	// 5. Run Server
	if err := r.Run(":8080"); err != nil {
		fmt.Println("Server failed to start:", err)
	}
}
