#!/bin/bash
echo "1. Clearing Orderbook..."
curl -X DELETE http://localhost:8080/orderbook
echo -e "\n"

echo "2. Placing Sell Limit: 10 @ 100.0"
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side": "sell", "size": 10, "price": 100}'
echo -e "\n"

echo "3. Placing Buy Limit: 5 @ 100.0 (Should match)"
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side": "buy", "size": 5, "price": 100}'
echo -e "\n"

echo "4. Checking Orderbook (Should have 5 Sell left)"
curl -X GET http://localhost:8080/orderbook
echo -e "\n"
