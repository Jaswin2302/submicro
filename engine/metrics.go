package engine

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

var (
	OrdersProcessed = promauto.NewCounter(prometheus.CounterOpts{
		Name: "submicro_orders_processed_total",
		Help: "Total number of orders processed by the engine",
	})

	TradesMatched = promauto.NewCounter(prometheus.CounterOpts{
		Name: "submicro_trades_matched_total",
		Help: "Total number of trades matched by the engine",
	})

	OrderBookDepth = promauto.NewGaugeVec(prometheus.GaugeOpts{
		Name: "submicro_order_book_depth",
		Help: "Number of price levels in the order book",
	}, []string{"side"})

	MatchLatency = promauto.NewHistogram(prometheus.HistogramOpts{
		Name:    "submicro_match_latency_nanoseconds",
		Help:    "Latency of order matching in nanoseconds",
		Buckets: []float64{100, 500, 1000, 5000, 10000, 50000, 100000},
	})
)