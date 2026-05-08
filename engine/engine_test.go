package engine

import (
	"fmt"
	"testing"
	"time"

	"github.com/shopspring/decimal"
)

func makeOrder(id string, side Side, price string, quantity string) *Order {
	return &Order{
		ID:        id,
		Side:      side,
		Price:     decimal.RequireFromString(price),
		Quantity:  decimal.RequireFromString(quantity),
		Filled:    decimal.Zero,
		Status:    Open,
		Timestamp: time.Now(),
	}
}

func TestBasicMatch(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-1", Sell, "50000", "1"))
	trades := ob.AddOrder(makeOrder("buy-1", Buy, "50000", "1"))

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if !trades[0].Quantity.Equal(decimal.RequireFromString("1")) {
		t.Errorf("expected trade quantity 1, got %s", trades[0].Quantity)
	}
	if !trades[0].Price.Equal(decimal.RequireFromString("50000")) {
		t.Errorf("expected trade price 50000, got %s", trades[0].Price)
	}
}

func TestNoMatchWhenPriceTooLow(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-1", Sell, "50000", "1"))
	trades := ob.AddOrder(makeOrder("buy-1", Buy, "49000", "1"))

	if len(trades) != 0 {
		t.Fatalf("expected 0 trades, got %d", len(trades))
	}
	if len(ob.Bids) != 1 {
		t.Errorf("expected buy order to rest in book")
	}
	if len(ob.Asks) != 1 {
		t.Errorf("expected sell order to remain in book")
	}
}

func TestPartialFill(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-1", Sell, "50000", "0.5"))
	trades := ob.AddOrder(makeOrder("buy-1", Buy, "50000", "1"))

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if !trades[0].Quantity.Equal(decimal.RequireFromString("0.5")) {
		t.Errorf("expected partial fill of 0.5, got %s", trades[0].Quantity)
	}
	if len(ob.Bids) != 1 {
		t.Errorf("expected remaining buy order to rest in book")
	}
}

func TestTimePriority(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-first", Sell, "50000", "1"))
	ob.AddOrder(makeOrder("sell-second", Sell, "50000", "1"))
	trades := ob.AddOrder(makeOrder("buy-1", Buy, "50000", "1"))

	if len(trades) != 1 {
		t.Fatalf("expected 1 trade, got %d", len(trades))
	}
	if trades[0].SellOrderID != "sell-first" {
		t.Errorf("expected sell-first to fill first, got %s", trades[0].SellOrderID)
	}
}

func TestCancelOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("buy-1", Buy, "50000", "1"))
	err := ob.CancelOrder("buy-1")

	if err != nil {
		t.Fatalf("expected no error, got %v", err)
	}
	if len(ob.Bids) != 0 {
		t.Errorf("expected book to be empty after cancel")
	}
}

func TestCancelNonExistentOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	err := ob.CancelOrder("does-not-exist")
	if err != ErrOrderNotFound {
		t.Errorf("expected ErrOrderNotFound, got %v", err)
	}
}

func TestCancelAlreadyFilledOrder(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-1", Sell, "50000", "1"))
	ob.AddOrder(makeOrder("buy-1", Buy, "50000", "1"))

	err := ob.CancelOrder("buy-1")
	if err != ErrOrderNotOpen {
		t.Errorf("expected ErrOrderNotOpen, got %v", err)
	}
}

func TestMultipleFillsAcrossPriceLevels(t *testing.T) {
	ob := NewOrderBook("BTCUSD")

	ob.AddOrder(makeOrder("sell-1", Sell, "49000", "1"))
	ob.AddOrder(makeOrder("sell-2", Sell, "50000", "1"))
	trades := ob.AddOrder(makeOrder("buy-1", Buy, "50000", "2"))

	if len(trades) != 2 {
		t.Fatalf("expected 2 trades, got %d", len(trades))
	}
	if !trades[0].Price.Equal(decimal.RequireFromString("49000")) {
		t.Errorf("expected first fill at 49000, got %s", trades[0].Price)
	}
	if !trades[1].Price.Equal(decimal.RequireFromString("50000")) {
		t.Errorf("expected second fill at 50000, got %s", trades[1].Price)
	}
}

func BenchmarkAddOrder(b *testing.B) {
	ob := NewOrderBook("BTCUSD")
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ob.AddOrder(makeOrder(
			fmt.Sprintf("order-%d", i),
			Buy,
			"50000",
			"1",
		))
	}
}

func BenchmarkMatchOrder(b *testing.B) {
	ob := NewOrderBook("BTCUSD")
	for i := 0; i < b.N; i++ {
		ob.AddOrder(makeOrder(
			fmt.Sprintf("sell-%d", i),
			Sell,
			"50000",
			"1",
		))
	}
	b.ResetTimer()
	for i := 0; i < b.N; i++ {
		ob.AddOrder(makeOrder(
			fmt.Sprintf("buy-%d", i),
			Buy,
			"50000",
			"1",
		))
	}
}