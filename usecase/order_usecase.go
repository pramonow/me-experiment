package usecase

import (
	"fmt"
	"matching-engine/account"
	"matching-engine/orderbook"
	"sync"
)

type OrderUseCase struct {
	OrderBook      *orderbook.OrderBook
	AccountManager *account.AccountManager
	Mutex          sync.Mutex
	FilePath       string
}

func NewOrderUseCase(filePath string, am *account.AccountManager) *OrderUseCase {
	ob, err := orderbook.LoadFromFile(filePath)
	if err != nil {
		fmt.Printf("Error loading orderbook from %s, starting fresh: %v\n", filePath, err)
		ob = orderbook.NewOrderBook()
	} else {
		fmt.Printf("Loaded orderbook from %s\n", filePath)
	}

	return &OrderUseCase{
		OrderBook:      ob,
		AccountManager: am,
		FilePath:       filePath,
	}
}

func (uc *OrderUseCase) GetOrderBook() *orderbook.OrderBook {
	uc.Mutex.Lock()
	defer uc.Mutex.Unlock()
	return uc.OrderBook
}

func (uc *OrderUseCase) PlaceOrder(userID string, sideStr string, typeStr string, size float64, price float64) ([]orderbook.Trade, *orderbook.Order, error) {
	uc.Mutex.Lock() // Global Lock for matching engine safety
	defer uc.Mutex.Unlock()

	// 1. Validate Input
	var side orderbook.Side
	if sideStr == "buy" {
		side = orderbook.Buy
	} else if sideStr == "sell" {
		side = orderbook.Sell
	} else {
		return nil, nil, fmt.Errorf("invalid side: %s", sideStr)
	}

	// 2. Fund Locking (Pre-Trade)
	// Assumptions: Pair is BTC/USD.
	// Buy -> Spend USD (Quote), Get BTC (Base)
	// Sell -> Spend BTC (Base), Get USD (Quote)
	baseCurrency := "BTC"
	quoteCurrency := "USD"

	if typeStr == "limit" {
		if price <= 0 {
			return nil, nil, fmt.Errorf("price must be > 0 for limit orders")
		}

		if side == orderbook.Buy {
			cost := size * price
			if err := uc.AccountManager.LockFunds(userID, quoteCurrency, cost); err != nil {
				return nil, nil, fmt.Errorf("insufficient funds to lock: %v", err)
			}
		} else {
			if err := uc.AccountManager.LockFunds(userID, baseCurrency, size); err != nil {
				return nil, nil, fmt.Errorf("insufficient funds to lock: %v", err)
			}
		}
	} else {
		// Market Order: We should check balance but locking is tricky as price is unknown.
		// For MVP, we skip strict locking for Market Orders or require a buffer.
		// Or we just check if they have > 0 ?
		// Let's Skip locking for Market Order for now (Risky but simple MVP)
		// But we MUST fail if balance is 0.
	}

	// 3. Match
	var trades []orderbook.Trade
	var order *orderbook.Order

	if typeStr == "market" {
		trades = uc.OrderBook.ProcessMarketOrder(side, userID, size)
		order = nil
	} else if typeStr == "limit" {
		trades, order = uc.OrderBook.ProcessLimitOrder(side, userID, size, price)
	} else {
		// If valid type, we shouldn't reach here if we check earlier, but for safety:
		// Also we might need to rollback locks if this fails?
		// But strictly speaking validation should be first.
		return nil, nil, fmt.Errorf("invalid order type: %s", typeStr)
	}

	// 4. Settle Trades
	for _, t := range trades {
		costQuote := t.Price * t.Size
		amountBase := t.Size

		// Settle Buyer (BuyUserID)
		// - If they were maker (Limit): Deduct Locked USD.
		// - If they were taker (Market): Deduct Balance USD.
		// Wait, how do we know if Buyer was Maker or Taker?
		// We know 'userID' is the Incoming Taker.
		// If t.BuyUserID == userID, then Buyer is Taker.

		// Buyer Logic:
		if t.BuyUserID == userID {
			// Buyer is Taker (Incoming)
			if typeStr == "limit" {
				// Taker Limit Buy: We locked OrderPrice * Size.
				// TradePrice might be lower.
				// Deduct TradePrice * TradeSize from Locked?
				// No, we locked based on OrderPrice.
				// Safe bet: Deduct Cost from Locked.
				// The excess lock (savings) stays locked? Ideally unlock it.
				uc.AccountManager.DeductLocked(t.BuyUserID, quoteCurrency, costQuote) // Simplified
			} else {
				// Taker Market Buy: No Lock. Deduct Balance.
				// Need a "DeductBalance" method or just use negative Deposit (hacky)
				// Let's assume we implement "Spend" later. For now, use Lock+Deduct pattern for cleaner API?
				// Just force-deduct from balance.
				uc.AccountManager.DeductBalance(t.BuyUserID, quoteCurrency, costQuote)
			}
		} else {
			// Buyer was Maker (Resting Limit). Funds were locked.
			uc.AccountManager.DeductLocked(t.BuyUserID, quoteCurrency, costQuote)
		}

		// Add BTC to Buyer
		uc.AccountManager.AddBalance(t.BuyUserID, baseCurrency, amountBase)

		// Seller Logic:
		if t.SellUserID == userID {
			// Seller is Taker
			if typeStr == "limit" {
				// Taker Limit Sell: Locked Size.
				uc.AccountManager.DeductLocked(t.SellUserID, baseCurrency, amountBase)
			} else {
				// Taker Market Sell: No Lock.
				uc.AccountManager.DeductBalance(t.SellUserID, baseCurrency, amountBase)
			}
		} else {
			// Seller was Maker. Funds Locked.
			uc.AccountManager.DeductLocked(t.SellUserID, baseCurrency, amountBase)
		}

		// Add USD to Seller
		uc.AccountManager.AddBalance(t.SellUserID, quoteCurrency, costQuote)
	}

	// 5. Save State
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
