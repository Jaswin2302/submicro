package engine

import (
	"time"

	"github.com/shopspring/decimal"
)

func (ob *OrderBook) AddOrder(order *Order) []Trade {
	start := time.Now()

	ob.Mu.Lock()
	defer ob.Mu.Unlock()

	ob.orderIndex[order.ID] = order

	var trades []Trade
	if order.Side == Buy {
		trades = ob.matchBuy(order)
	} else {
		trades = ob.matchSell(order)
	}

	OrdersProcessed.Inc()
	TradesMatched.Add(float64(len(trades)))
	OrderBookDepth.WithLabelValues("bid").Set(float64(len(ob.Bids)))
	OrderBookDepth.WithLabelValues("ask").Set(float64(len(ob.Asks)))
	MatchLatency.Observe(float64(time.Since(start).Nanoseconds()))

	return trades
}

func (ob *OrderBook) matchBuy(order *Order) []Trade {
	trades := make([]Trade, 0)

	for len(ob.Asks) > 0 && order.Status != Filled {
		bestAsk := ob.Asks[0]

		if order.Price.LessThan(bestAsk.Price) {
			break
		}

		for len(bestAsk.Orders) > 0 && order.Status != Filled {
			trades = append(trades, ob.fillOrders(order, bestAsk.Orders[0]))
			if bestAsk.Orders[0].Status == Filled {
				bestAsk.Orders = bestAsk.Orders[1:]
			}
		}

		if len(bestAsk.Orders) == 0 {
			ob.Asks = ob.Asks[1:]
		}
	}

	if order.Status != Filled {
		ob.insertBid(order)
	}

	return trades
}

func (ob *OrderBook) matchSell(order *Order) []Trade {
	trades := make([]Trade, 0)

	for len(ob.Bids) > 0 && order.Status != Filled {
		bestBid := ob.Bids[0]

		if order.Price.GreaterThan(bestBid.Price) {
			break
		}

		for len(bestBid.Orders) > 0 && order.Status != Filled {
			trades = append(trades, ob.fillOrders(bestBid.Orders[0], order))
			if bestBid.Orders[0].Status == Filled {
				bestBid.Orders = bestBid.Orders[1:]
			}
		}

		if len(bestBid.Orders) == 0 {
			ob.Bids = ob.Bids[1:]
		}
	}

	if order.Status != Filled {
		ob.insertAsk(order)
	}

	return trades
}

func (ob *OrderBook) fillOrders(buy *Order, sell *Order) Trade {
	quantity := decimal.Min(
		buy.Quantity.Sub(buy.Filled),
		sell.Quantity.Sub(sell.Filled),
	)

	buy.Filled = buy.Filled.Add(quantity)
	sell.Filled = sell.Filled.Add(quantity)

	if buy.Filled.Equal(buy.Quantity) {
		buy.Status = Filled
	} else {
		buy.Status = PartiallyFilled
	}

	if sell.Filled.Equal(sell.Quantity) {
		sell.Status = Filled
	} else {
		sell.Status = PartiallyFilled
	}

	return Trade{
		BuyOrderID:  buy.ID,
		SellOrderID: sell.ID,
		Price:       sell.Price,
		Quantity:    quantity,
		Timestamp:   time.Now(),
	}
}

func (ob *OrderBook) insertBid(order *Order) {
	for i, level := range ob.Bids {
		if order.Price.GreaterThan(level.Price) {
			ob.Bids = append(ob.Bids[:i], append([]*PriceLevel{{Price: order.Price, Orders: []*Order{order}}}, ob.Bids[i:]...)...)
			return
		}
		if order.Price.Equal(level.Price) {
			level.Orders = append(level.Orders, order)
			return
		}
	}
	ob.Bids = append(ob.Bids, &PriceLevel{Price: order.Price, Orders: []*Order{order}})
}

func (ob *OrderBook) insertAsk(order *Order) {
	for i, level := range ob.Asks {
		if order.Price.LessThan(level.Price) {
			ob.Asks = append(ob.Asks[:i], append([]*PriceLevel{{Price: order.Price, Orders: []*Order{order}}}, ob.Asks[i:]...)...)
			return
		}
		if order.Price.Equal(level.Price) {
			level.Orders = append(level.Orders, order)
			return
		}
	}
	ob.Asks = append(ob.Asks, &PriceLevel{Price: order.Price, Orders: []*Order{order}})
}