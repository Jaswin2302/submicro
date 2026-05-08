package main

import (
	"log/slog"
	"os"
	"time"

	"github.com/Jaswin2302/submicro/api"
	"github.com/Jaswin2302/submicro/engine"
	"github.com/Jaswin2302/submicro/feed"
	"github.com/shopspring/decimal"
)

func main() {
	logger := slog.New(slog.NewTextHandler(os.Stdout, nil))
	slog.SetDefault(logger)

	slog.Info("starting submicro")

	book := engine.NewOrderBook("BTCUSD")
	server := api.New(book)

	go func() {
		if err := server.Start(":8080"); err != nil {
			slog.Error("api server failed", "err", err)
			os.Exit(1)
		}
	}()

	f := feed.New(func(update feed.MarketUpdate) {
		for _, event := range update.Events {
			if event.Type != "change" {
				continue
			}

			price, err := decimal.NewFromString(event.Price)
			if err != nil {
				slog.Error("invalid price", "price", event.Price)
				continue
			}

			remaining, err := decimal.NewFromString(event.Remaining)
			if err != nil {
				slog.Error("invalid remaining", "remaining", event.Remaining)
				continue
			}

			if remaining.IsZero() {
				continue
			}

			var side engine.Side
			if event.Side == "bid" {
				side = engine.Buy
			} else {
				side = engine.Sell
			}

			order := &engine.Order{
				ID:        event.Price + "-" + event.Side + "-" + event.Remaining,
				Side:      side,
				Price:     price,
				Quantity:  remaining,
				Filled:    decimal.Zero,
				Status:    engine.Open,
				Timestamp: time.Now(),
			}

			trades := book.AddOrder(order)

			server.BroadcastBook()

			if len(trades) > 0 {
				for _, trade := range trades {
					slog.Info("TRADE MATCHED",
						"price", trade.Price,
						"quantity", trade.Quantity,
					)
					server.BroadcastTrade(
						trade.Price.String(),
						trade.Quantity.String(),
						event.Side,
					)
				}
			}
		}
	})

	slog.Info("connecting to Gemini WebSocket feed")
	f.Start()
}