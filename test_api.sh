#!/bin/bash
echo "1. Clearing Orderbook..."
curl -X DELETE http://localhost:8080/orderbook
echo -e "\n"

echo "2. Placing Sell Limit: 10 @ 100.0"
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side": "sell", "type": "limit", "size": 10, "price": 100}'
echo -e "\n"

echo "3. Placing Buy Limit: 5 @ 100.0 (Should match)"
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side": "buy", "type": "limit", "size": 5, "price": 100}'
echo -e "\n"

echo "4. Placing Buy Market: 2 (Should match best ask)"
curl -X POST http://localhost:8080/orders \
  -H "Content-Type: application/json" \
  -d '{"side": "buy", "type": "market", "size": 2}'
echo -e "\n"

echo "5. Checking Orderbook (Should have 3 Sell left)"
curl -X GET http://localhost:8080/orderbook
echo -e "\n"
