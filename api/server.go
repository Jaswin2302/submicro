package api

import (
	"encoding/json"
	"log/slog"
	"net/http"
	"sync"

	"github.com/Jaswin2302/submicro/engine"
	"github.com/gorilla/websocket"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

var upgrader = websocket.Upgrader{
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

type client struct {
	conn *websocket.Conn
	mu   sync.Mutex
}

func (c *client) write(msg Message) error {
	data, err := json.Marshal(msg)
	if err != nil {
		return err
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	return c.conn.WriteMessage(websocket.TextMessage, data)
}

type Server struct {
	book    *engine.OrderBook
	clients map[*websocket.Conn]*client
	mu      sync.Mutex
}

type Message struct {
	Type    string      `json:"type"`
	Payload interface{} `json:"payload"`
}

type BookSnapshot struct {
	Bids []PriceLevel `json:"bids"`
	Asks []PriceLevel `json:"asks"`
}

type PriceLevel struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
}

type TradePayload struct {
	Price    string `json:"price"`
	Quantity string `json:"quantity"`
	Side     string `json:"side"`
}

func New(book *engine.OrderBook) *Server {
	return &Server{
		book:    book,
		clients: make(map[*websocket.Conn]*client),
	}
}

func (s *Server) Start(addr string) error {
	http.HandleFunc("/ws", s.handleWS)
	http.HandleFunc("/health", s.handleHealth)
	http.Handle("/metrics", promhttp.Handler())
	slog.Info("api server starting", "addr", addr)
	return http.ListenAndServe(addr, nil)
}

func (s *Server) handleHealth(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(http.StatusOK)
	w.Write([]byte("ok"))
}

func (s *Server) handleWS(w http.ResponseWriter, r *http.Request) {
	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		slog.Error("failed to upgrade connection", "err", err)
		return
	}
	defer conn.Close()

	c := &client{conn: conn}

	s.mu.Lock()
	s.clients[conn] = c
	s.mu.Unlock()

	slog.Info("client connected", "addr", conn.RemoteAddr())

	snapshot := s.buildSnapshot()
	c.write(Message{Type: "snapshot", Payload: snapshot})

	for {
		_, _, err := conn.ReadMessage()
		if err != nil {
			break
		}
	}

	s.mu.Lock()
	delete(s.clients, conn)
	s.mu.Unlock()

	slog.Info("client disconnected", "addr", conn.RemoteAddr())
}

func (s *Server) buildSnapshot() BookSnapshot {
	s.book.Mu.RLock()
	defer s.book.Mu.RUnlock()

	snapshot := BookSnapshot{
		Bids: make([]PriceLevel, 0),
		Asks: make([]PriceLevel, 0),
	}

	for _, level := range s.book.Bids {
		total := level.TotalQuantity()
		snapshot.Bids = append(snapshot.Bids, PriceLevel{
			Price:    level.Price.String(),
			Quantity: total.String(),
		})
	}

	for _, level := range s.book.Asks {
		total := level.TotalQuantity()
		snapshot.Asks = append(snapshot.Asks, PriceLevel{
			Price:    level.Price.String(),
			Quantity: total.String(),
		})
	}

	return snapshot
}

func (s *Server) BroadcastTrade(price, quantity, side string) {
	s.broadcast(Message{
		Type: "trade",
		Payload: TradePayload{
			Price:    price,
			Quantity: quantity,
			Side:     side,
		},
	})
}

func (s *Server) BroadcastBook() {
	snapshot := s.buildSnapshot()
	s.broadcast(Message{Type: "book_update", Payload: snapshot})
}

func (s *Server) broadcast(msg Message) {
	s.mu.Lock()
	defer s.mu.Unlock()

	for conn, c := range s.clients {
		if err := c.write(msg); err != nil {
			slog.Error("failed to write to client", "err", err)
			conn.Close()
			delete(s.clients, conn)
		}
	}
}