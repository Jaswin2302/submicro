package feed

import (
	"encoding/json"
	"log/slog"
	"time"

	"github.com/gorilla/websocket"
)

const geminiWSURL = "wss://api.gemini.com/v1/marketdata/BTCUSD"

type Handler func(update MarketUpdate)

type Feed struct {
	url     string
	handler Handler
	conn    *websocket.Conn
}

type MarketUpdate struct {
	Type    string  `json:"type"`
	EventID int64   `json:"eventId"`
	Events  []Event `json:"events"`
}

type Event struct {
	Type      string `json:"type"`
	Side      string `json:"side"`
	Price     string `json:"price"`
	Remaining string `json:"remaining"`
	Delta     string `json:"delta"`
	Reason    string `json:"reason"`
}

func New(handler Handler) *Feed {
	return &Feed{
		url:     geminiWSURL,
		handler: handler,
	}
}

func (f *Feed) Connect() error {
	conn, _, err := websocket.DefaultDialer.Dial(f.url, nil)
	if err != nil {
		return err
	}
	f.conn = conn
	slog.Info("connected to Gemini WebSocket feed", "url", f.url)
	return nil
}

func (f *Feed) Start() {
	for {
		if err := f.Connect(); err != nil {
			slog.Error("failed to connect, retrying in 5s", "err", err)
			time.Sleep(5 * time.Second)
			continue
		}
		f.listen()
		slog.Warn("feed disconnected, reconnecting in 5s")
		time.Sleep(5 * time.Second)
	}
}

func (f *Feed) listen() {
	defer f.conn.Close()
	for {
		_, msg, err := f.conn.ReadMessage()
		if err != nil {
			slog.Error("read error", "err", err)
			return
		}
		var update MarketUpdate
		if err := json.Unmarshal(msg, &update); err != nil {
			slog.Error("failed to parse message", "err", err)
			continue
		}
		f.handler(update)
	}
}