package main

import (
	"fmt"
	"matching-engine/orderbook"
)

func main() {
	ob := orderbook.NewOrderBook()

	fmt.Println("1. Placing Sell Limit Order: 10 @ 100.0")
	trades, order := ob.ProcessLimitOrder(orderbook.Sell, 10.0, 100.0)
	printTrades(trades)
	if order != nil {
		fmt.Printf("Order placed on book: ID %d\n", order.ID)
	}
	fmt.Println(ob)

	fmt.Println("\n2. Placing Sell Limit Order: 5 @ 102.0")
	trades, order = ob.ProcessLimitOrder(orderbook.Sell, 5.0, 102.0)
	printTrades(trades)
	if order != nil {
		fmt.Printf("Order placed on book: ID %d\n", order.ID)
	}
	fmt.Println(ob)

	fmt.Println("\n3. Placing Buy Limit Order: 12 @ 101.0")
	// Should match the 10 @ 100 first.
	// Remaining 2 should be placed at 101.
	trades, order = ob.ProcessLimitOrder(orderbook.Buy, 12.0, 101.0)
	printTrades(trades)
	if order != nil {
		fmt.Printf("Remaining Order placed on book: ID %d, Size: %.2f\n", order.ID, order.Size)
	}
	fmt.Println(ob)
}

func printTrades(trades []orderbook.Trade) {
	if len(trades) == 0 {
		return
	}
	fmt.Println("Trades Executed:")
	for _, t := range trades {
		fmt.Printf("- Price: %.2f, Size: %.2f\n", t.Price, t.Size)
	}
}
