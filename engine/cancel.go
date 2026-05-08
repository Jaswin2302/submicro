package engine

import "errors"

var (
	ErrOrderNotFound = errors.New("order not found")
	ErrOrderNotOpen  = errors.New("order is not open")
)

func (ob *OrderBook) CancelOrder(orderID string) error {
	ob.Mu.Lock()
	defer ob.Mu.Unlock()

	order, exists := ob.orderIndex[orderID]
	if !exists {
		return ErrOrderNotFound
	}

	if order.Status != Open && order.Status != PartiallyFilled {
		return ErrOrderNotOpen
	}

	order.Status = Cancelled

	if order.Side == Buy {
		ob.removFromBids(order)
	} else {
		ob.removeFromAsks(order)
	}

	delete(ob.orderIndex, orderID)

	return nil
}

func (ob *OrderBook) removFromBids(order *Order) {
	for i, level := range ob.Bids {
		if level.Price.Equal(order.Price) {
			for j, o := range level.Orders {
				if o.ID == order.ID {
					level.Orders = append(level.Orders[:j], level.Orders[j+1:]...)
					break
				}
			}
			if len(level.Orders) == 0 {
				ob.Bids = append(ob.Bids[:i], ob.Bids[i+1:]...)
			}
			return
		}
	}
}

func (ob *OrderBook) removeFromAsks(order *Order) {
	for i, level := range ob.Asks {
		if level.Price.Equal(order.Price) {
			for j, o := range level.Orders {
				if o.ID == order.ID {
					level.Orders = append(level.Orders[:j], level.Orders[j+1:]...)
					break
				}
			}
			if len(level.Orders) == 0 {
				ob.Asks = append(ob.Asks[:i], ob.Asks[i+1:]...)
			}
			return
		}
	}
}