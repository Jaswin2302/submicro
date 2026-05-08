package engine

import (
	"time"

	"github.com/shopspring/decimal"
)

type Side int

const (
	Buy  Side = iota
	Sell
)

type OrderStatus int

const (
	Open      OrderStatus = iota
	Filled
	Cancelled
	PartiallyFilled
)

type Order struct {
	ID        string
	Side      Side
	Price     decimal.Decimal
	Quantity  decimal.Decimal
	Filled    decimal.Decimal
	Status    OrderStatus
	Timestamp time.Time
}

type Trade struct {
	BuyOrderID  string
	SellOrderID string
	Price       decimal.Decimal
	Quantity    decimal.Decimal
	Timestamp   time.Time
}