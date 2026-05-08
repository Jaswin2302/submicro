# SubMicro

A high-performance cryptocurrency limit order book matching engine written in Go, fed by Gemini's live WebSocket API, with a React/TypeScript frontend showing real-time order book depth and trade tape.

![Go](https://img.shields.io/badge/Go-1.26-blue)
![React](https://img.shields.io/badge/React-TypeScript-blue)
![Docker](https://img.shields.io/badge/Docker-containerised-blue)

## Benchmark Results

Run on 12th Gen Intel Core i5-12500H, Go 1.26.2, Windows 11:
BenchmarkAddOrder-16      8638550    716 ns/op    368 B/op    7 allocs/op
BenchmarkMatchOrder-16    5164989   1285 ns/op    718 B/op   24 allocs/op

- **1.4M order insertions/sec** at 716ns per insertion
- **778K matches/sec** at 1285ns per match
- **54% of orders matched in under 100 nanoseconds** — measured via Prometheus histogram
- **5,000+ orders processed** per connection with zero crashes

## Architecture
Gemini WebSocket API (live BTC/USD feed)
│
▼
feed/ package        — reconnecting WebSocket consumer
│
▼
engine/ package       — price-time priority matching engine
│
▼
api/ package         — WebSocket + Prometheus metrics server
│
▼
web/ (React/TS)        — live order book frontend

## Features

- Price-time priority matching engine in Go — the same algorithm used by real exchanges
- Fed by Gemini's live BTC/USD WebSocket API — zero mock data
- Sub-microsecond matching latency measured via Prometheus histograms
- Real-time React/TypeScript frontend with live order book depth bars and trade tape
- Production-grade reconnecting WebSocket feed with automatic reconnection
- Prometheus metrics at `/metrics` — orders/sec, match latency histogram, book depth
- Multi-stage Docker build producing a ~15MB production image
- Full test suite covering partial fills, time priority, and cancellation edge cases

## Running Locally

**Without Docker:**

Terminal 1 — from project root:
```bash
go run cmd/submicro/main.go
```

Terminal 2 — from `web/`:
```bash
npm install && npm run dev
```

Open `http://localhost:5173`

**With Docker (recommended):**
```bash
docker compose up --build
```

Open `http://localhost:5173`

## Observability

Prometheus metrics exposed at `http://localhost:8080/metrics`:

| Metric | Type | Description |
|--------|------|-------------|
| `submicro_orders_processed_total` | Counter | Total orders processed by the engine |
| `submicro_trades_matched_total` | Counter | Total trades matched |
| `submicro_order_book_depth` | Gauge | Current price levels per side |
| `submicro_match_latency_nanoseconds` | Histogram | Per-order matching latency |

## Design Decisions

**Why Go?** Go's goroutine model maps perfectly to the concurrent nature of an exchange — the feed consumer, matching engine, and API server run as independent goroutines. Go is also Gemini's primary backend language.

**Why shopspring/decimal over float64?** Floating point arithmetic has rounding errors — `0.1 + 0.2 = 0.30000000000000004` in float64. In a financial system that's catastrophic. Every price and quantity uses `decimal.Decimal`.

**Why RWMutex?** The feed writes to the order book while the API reads it to build snapshots for connected clients. `sync.RWMutex` allows many concurrent readers but only one writer — perfect for this access pattern.

**Known limitation:** The current slice-based price level structure has O(n) insertion. At the scale of a real exchange this would be replaced with a skip list or balanced BST for O(log n) insertion and cancellation. I discovered this empirically through benchmarking — the cancel benchmark degraded on deep books, revealing the bottleneck.

## What I Learned

The most surprising discovery was that the bottleneck in a naive order book implementation isn't the matching algorithm — it's the data structure used to store price levels. I found this through benchmarking rather than theory, which is the right way to find real bottlenecks in performance-critical systems.
