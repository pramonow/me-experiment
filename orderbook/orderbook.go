package orderbook

import (
	"fmt"
	"sort"
	"time"
)

type Side int

const (
	Buy Side = iota
	Sell
)

type Order struct {
	ID        uint64
	Side      Side
	Size      float64
	Price     float64
	Timestamp int64
}

type Trade struct {
	BuyOrderID  uint64
	SellOrderID uint64
	Price       float64
	Size        float64
	Timestamp   int64
}

type OrderBook struct {
	Bids []*Order // Descending Price
	Asks []*Order // Ascending Price
}

func NewOrderBook() *OrderBook {
	return &OrderBook{
		Bids: []*Order{},
		Asks: []*Order{},
	}
}

// ProcessLimitOrder handles the matching logic for a limit order
// It returns a list of trades executd and the remaining order (if any)
func (ob *OrderBook) ProcessLimitOrder(side Side, size float64, price float64) ([]Trade, *Order) {
	trades := []Trade{}
	remainingSize := size

	if side == Buy {
		// Matching against Asks (Sellers)
		// We want to buy low. We check the Asks from lowest price upwards.
		// As long as BestAsk <= MyLimitPrice, we match.
		for len(ob.Asks) > 0 && remainingSize > 0 {
			bestAsk := ob.Asks[0] // Lowest Ask
			if bestAsk.Price > price {
				break // Cannot match anymore
			}

			// Match found
			matchSize := min(remainingSize, bestAsk.Size)
			trades = append(trades, Trade{
				BuyOrderID:  0, // Temporary ID for incoming
				SellOrderID: bestAsk.ID,
				Price:       bestAsk.Price, // Maker's price usually determines trade price
				Size:        matchSize,
				Timestamp:   time.Now().UnixNano(),
			})

			remainingSize -= matchSize
			bestAsk.Size -= matchSize

			// Clean up filled orders
			if bestAsk.Size == 0 {
				ob.Asks = ob.Asks[1:]
			}
		}

		// If still remaining, add to Bids
		if remainingSize > 0 {
			newOrder := &Order{
				ID:        uint64(time.Now().UnixNano()), // Simple ID generation
				Side:      Buy,
				Size:      remainingSize,
				Price:     price,
				Timestamp: time.Now().UnixNano(),
			}
			ob.Bids = append(ob.Bids, newOrder)
			// Bids are sorted by Price Descending (Highest first)
			sort.Slice(ob.Bids, func(i, j int) bool {
				if ob.Bids[i].Price == ob.Bids[j].Price {
					return ob.Bids[i].Timestamp < ob.Bids[j].Timestamp // Time Priority
				}
				return ob.Bids[i].Price > ob.Bids[j].Price
			})
			return trades, newOrder
		}

	} else {
		// Sell Order
		// Matching against Bids (Buyers)
		// We want to sell high. We check Bids from highest price downwards.
		// As long as BestBid >= MyLimitPrice, we match.
		for len(ob.Bids) > 0 && remainingSize > 0 {
			bestBid := ob.Bids[0] // Highest Bid
			if bestBid.Price < price {
				break // Cannot match anymore
			}

			// Match found
			matchSize := min(remainingSize, bestBid.Size)
			trades = append(trades, Trade{
				BuyOrderID:  bestBid.ID,
				SellOrderID: 0, // Temporary
				Price:       bestBid.Price,
				Size:        matchSize,
				Timestamp:   time.Now().UnixNano(),
			})

			remainingSize -= matchSize
			bestBid.Size -= matchSize

			if bestBid.Size == 0 {
				ob.Bids = ob.Bids[1:]
			}
		}

		// If still remaining, add to Asks
		if remainingSize > 0 {
			newOrder := &Order{
				ID:        uint64(time.Now().UnixNano()),
				Side:      Sell,
				Size:      remainingSize,
				Price:     price,
				Timestamp: time.Now().UnixNano(),
			}
			ob.Asks = append(ob.Asks, newOrder)
			// Asks are sorted by Price Ascending (Lowest first)
			sort.Slice(ob.Asks, func(i, j int) bool {
				if ob.Asks[i].Price == ob.Asks[j].Price {
					return ob.Asks[i].Timestamp < ob.Asks[j].Timestamp
				}
				return ob.Asks[i].Price < ob.Asks[j].Price
			})
			return trades, newOrder
		}
	}

	return trades, nil
}

func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

func (ob *OrderBook) String() string {
	s := "--- Order Book ---\n"
	s += "ASKS (Sellers):\n"
	for i := len(ob.Asks) - 1; i >= 0; i-- {
		s += fmt.Sprintf("  Price: %.2f | Size: %.2f\n", ob.Asks[i].Price, ob.Asks[i].Size)
	}
	s += "------------------\n"
	s += "BIDS (Buyers):\n"
	for _, bid := range ob.Bids {
		s += fmt.Sprintf("  Price: %.2f | Size: %.2f\n", bid.Price, bid.Size)
	}
	s += "------------------\n"
	return s
}
