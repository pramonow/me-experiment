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

func (uc *OrderUseCase) PlaceOrder(sideStr string, typeStr string, size float64, price float64) ([]orderbook.Trade, *orderbook.Order, error) {
	uc.Mutex.Lock()
	defer uc.Mutex.Unlock()

	var side orderbook.Side
	switch sideStr {
	case "buy":
		side = orderbook.Buy
	case "sell":
		side = orderbook.Sell
	default:
		return nil, nil, fmt.Errorf("invalid side: %s", sideStr)
	}

	var trades []orderbook.Trade
	var order *orderbook.Order

	if typeStr == "market" {
		trades = uc.OrderBook.ProcessMarketOrder(side, size)
		order = nil // Market orders don't rest
	} else if typeStr == "limit" {
		if price <= 0 {
			return nil, nil, fmt.Errorf("price must be > 0 for limit orders")
		}
		trades, order = uc.OrderBook.ProcessLimitOrder(side, size, price)
	} else {
		return nil, nil, fmt.Errorf("invalid order type: %s", typeStr)
	}

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
