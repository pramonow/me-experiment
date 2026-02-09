package main

import (
	"fmt"
	"log"
	"matching-engine/handler"
	"matching-engine/internal/domain/account"
	"matching-engine/internal/infrastructure/database"
	"matching-engine/usecase"

	"github.com/gin-gonic/gin"
)

const OrderBookFile = "orderbook.json"

func main() {
	// 1. Connect to Postgres
	// Ideally use environment variables, but hardcoding for local POC for now
	db, err := database.NewPostgresDB("localhost", "5432", "postgres", "postgres", "matching_engine")
	if err != nil {
		log.Printf("Warning: Could not connect to database: %v. Continuing without DB persistence.", err)
	} else {
		defer db.Close()
	}

	// 2. Initialize AccountManager
	accountManager := account.NewAccountManager(db)

	// 2. Initialize UseCase
	orderUC := usecase.NewOrderUseCase(OrderBookFile, accountManager)

	// 3. Initialize Handler
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
