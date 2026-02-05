package account

import (
	"fmt"
	"sync"
)

type Account struct {
	UserID   string
	Balances map[string]float64
	Locked   map[string]float64
}

type AccountManager struct {
	Accounts map[string]*Account
	Mutex    sync.RWMutex
}

func NewAccountManager() *AccountManager {
	return &AccountManager{
		Accounts: make(map[string]*Account),
	}
}

func (am *AccountManager) CreateAccount(userID string) *Account {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	if acc, exists := am.Accounts[userID]; exists {
		return acc
	}

	acc := &Account{
		UserID:   userID,
		Balances: make(map[string]float64),
		Locked:   make(map[string]float64),
	}
	am.Accounts[userID] = acc
	return acc
}

func (am *AccountManager) GetAccount(userID string) (*Account, error) {
	am.Mutex.RLock()
	defer am.Mutex.RUnlock()

	if acc, exists := am.Accounts[userID]; exists {
		return acc, nil
	}
	return nil, fmt.Errorf("account not found: %s", userID)
}

func (am *AccountManager) Deposit(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("account not found")
	}

	acc.Balances[currency] += amount
	return nil
}

// LockFunds moves funds from Balance to Locked
func (am *AccountManager) LockFunds(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("account not found")
	}

	if acc.Balances[currency] < amount {
		return fmt.Errorf("insufficient funds")
	}

	acc.Balances[currency] -= amount
	acc.Locked[currency] += amount
	return nil
}

// UnlockFunds moves funds from Locked back to Balance
func (am *AccountManager) UnlockFunds(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("account not found")
	}

	if acc.Locked[currency] < amount {
		return fmt.Errorf("insufficient locked funds")
	}

	acc.Locked[currency] -= amount
	acc.Balances[currency] += amount
	return nil
}

// DeductLocked permanently removes locked funds (spent)
func (am *AccountManager) DeductLocked(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if acc.Locked[currency] < amount {
		return fmt.Errorf("insufficient locked funds to deduct")
	}
	acc.Locked[currency] -= amount
	return nil
}

// DeductBalance removes funds from balance (e.g. market order taker)
func (am *AccountManager) DeductBalance(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	if acc.Balances[currency] < amount {
		return fmt.Errorf("insufficient funds")
	}
	acc.Balances[currency] -= amount
	return nil
}

// AddBalance adds funds (received from trade)
func (am *AccountManager) AddBalance(userID string, currency string, amount float64) error {
	am.Mutex.Lock()
	defer am.Mutex.Unlock()

	acc, exists := am.Accounts[userID]
	if !exists {
		return fmt.Errorf("user not found")
	}

	acc.Balances[currency] += amount
	return nil
}
