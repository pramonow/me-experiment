#!/bin/bash
HOST="http://localhost:8080"

echo "1. Clearing Orderbook..."
curl -X DELETE $HOST/orderbook
echo -e "\n"

echo "2. Creating Accounts..."
curl -X POST $HOST/accounts -H "Content-Type: application/json" -d '{"user_id": "userA"}'
curl -X POST $HOST/accounts -H "Content-Type: application/json" -d '{"user_id": "userB"}'
echo -e "\n"

echo "3. Depositing Funds..."
# UserA has BTC to sell
curl -X POST $HOST/accounts/deposit -H "Content-Type: application/json" -d '{"user_id": "userA", "currency": "BTC", "amount": 10}'
# UserB has USD to buy
curl -X POST $HOST/accounts/deposit -H "Content-Type: application/json" -d '{"user_id": "userB", "currency": "USD", "amount": 2000}'
echo -e "\n"

echo "4. Checking Initial Balances..."
curl $HOST/accounts/userA
curl $HOST/accounts/userB
echo -e "\n"

echo "5. UserA Places Sell Limit: 5 BTC @ 100 USD (Locks 5 BTC)"
curl -X POST $HOST/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": "userA", "side": "sell", "type": "limit", "size": 5, "price": 100}'
echo -e "\n"

echo "6. UserB Places Buy Market: 2 BTC (Should match 2 @ 100)"
curl -X POST $HOST/orders \
  -H "Content-Type: application/json" \
  -d '{"user_id": "userB", "side": "buy", "type": "market", "size": 2}'
echo -e "\n"

echo "7. Checking Final Balances..."
echo "UserA (Seller): Should have 5 BTC remaining (3 Locked in Order + 2 Sold -> Gone?), +200 USD"
# Wait, UserA had 10 BTC. Sold 5 Limit. Locked 5. 
# Matched 2. 
# Locked: 3 BTC. 
# Balance: 5 BTC (original unused).
# Total BTC: 8? No. 
# Initial: 10.
# Sell Order 5: Balance 5, Locked 5.
# Match 2: Locked 5 -> 3. Sold 2. 
# UserA Balance: 5 BTC. Locked: 3 BTC. USD: 200.
curl $HOST/accounts/userA
echo -e "\n"

echo "UserB (Buyer): Should have 2000 - 200 = 1800 USD, +2 BTC"
curl $HOST/accounts/userB
echo -e "\n"

echo "8. Checking Orderbook"
curl $HOST/orderbook
echo -e "\n"
