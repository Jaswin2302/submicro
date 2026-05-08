package engine

import (
	"sync"

	"github.com/shopspring/decimal"
)

type PriceLevel struct {
	Price  decimal.Decimal
	Orders []*Order
}

func (pl *PriceLevel) TotalQuantity() decimal.Decimal {
	total := decimal.Zero
	for _, o := range pl.Orders {
		total = total.Add(o.Quantity.Sub(o.Filled))
	}
	return total
}

type OrderBook struct {
	Symbol     string
	Bids       []*PriceLevel
	Asks       []*PriceLevel
	orderIndex map[string]*Order
	Mu         sync.RWMutex
}

func NewOrderBook(symbol string) *OrderBook {
	return &OrderBook{
		Symbol:     symbol,
		Bids:       make([]*PriceLevel, 0),
		Asks:       make([]*PriceLevel, 0),
		orderIndex: make(map[string]*Order),
	}
}