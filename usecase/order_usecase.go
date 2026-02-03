package usecase

import (
	"fmt"
	"matching-engine/orderbook"
	"sync"
)

type OrderUseCase struct {
	OrderBook *orderbook.OrderBook
	Mutex     sync.Mutex
	FilePath  string
}

func NewOrderUseCase(filePath string) *OrderUseCase {
	ob, err := orderbook.LoadFromFile(filePath)
	if err != nil {
		fmt.Printf("Error loading orderbook from %s, starting fresh: %v\n", filePath, err)
		ob = orderbook.NewOrderBook()
	} else {
		fmt.Printf("Loaded orderbook from %s\n", filePath)
	}

	return &OrderUseCase{
		OrderBook: ob,
		FilePath:  filePath,
	}
}

func (uc *OrderUseCase) GetOrderBook() *orderbook.OrderBook {
	uc.Mutex.Lock()
	defer uc.Mutex.Unlock()
	return uc.OrderBook
}

func (uc *OrderUseCase) PlaceLimitOrder(sideStr string, size float64, price float64) ([]orderbook.Trade, *orderbook.Order, error) {
	uc.Mutex.Lock()
	defer uc.Mutex.Unlock()

	var side orderbook.Side
	if sideStr == "buy" {
		side = orderbook.Buy
	} else if sideStr == "sell" {
		side = orderbook.Sell
	} else {
		return nil, nil, fmt.Errorf("invalid side: %s", sideStr)
	}

	trades, order := uc.OrderBook.ProcessLimitOrder(side, size, price)

	if err := uc.OrderBook.SaveToFile(uc.FilePath); err != nil {
		fmt.Printf("Error saving orderbook: %v\n", err)
	}

	return trades, order, nil
}

func (uc *OrderUseCase) ClearOrderBook() error {
	uc.Mutex.Lock()
	defer uc.Mutex.Unlock()

	uc.OrderBook = orderbook.NewOrderBook()
	return uc.OrderBook.SaveToFile(uc.FilePath)
}
