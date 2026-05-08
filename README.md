# SubMicro

A high-performance cryptocurrency limit order book matching engine written in Go, fed by Gemini's live WebSocket API, with a React/TypeScript frontend showing real-time order book depth and trade tape.

**[Live Demo →](https://submicro.vercel.app)**

![Go](https://img.shields.io/badge/Go-1.26-blue)
![React](https://img.shields.io/badge/React-TypeScript-blue)
![Docker](https://img.shields.io/badge/Docker-containerised-blue)

## Benchmark Results

Run on 12th Gen Intel Core i5-12500H, Go 1.26.2, Windows 11:
BenchmarkAddOrder-16      8638550    716 ns/op    368 B/op    7 allocs/op
BenchmarkMatchOrder-16    5164989   1285 ns/op    718 B/op   24 allocs/op

1.4M order insertions/sec at 716ns per insertion. 778K matches/sec at 1285ns per match. 54% of orders matched in under 100 nanoseconds, measured via Prometheus histogram. 5,000+ orders processed per connection with zero crashes.

## Architecture

Live Bitcoin/USD market data is streamed from Gemini's public WebSocket API into the feed package, which maintains a persistent reconnecting connection. The feed passes each market event into the engine package, a price-time priority matching engine that processes orders and generates trades. The api package exposes the engine state to the world via a WebSocket server for the frontend and a Prometheus metrics endpoint for observability. The web package is a React/TypeScript frontend that connects to the API and renders the live order book, trade tape, and spread in real time.

## Features

Price-time priority matching engine in Go, the same algorithm used by real exchanges. Fed by Gemini's live BTC/USD WebSocket API with zero mock data. Sub-microsecond matching latency measured via Prometheus histograms. Real-time React/TypeScript frontend with live order book depth bars and trade tape. Production-grade reconnecting WebSocket feed with automatic reconnection. Prometheus metrics at /metrics covering orders/sec, match latency histogram, and book depth. Multi-stage Docker build producing a ~15MB production image. Full test suite covering partial fills, time priority, and cancellation edge cases.

## Running Locally

Without Docker, open two terminals. In the first terminal from the project root run:

```bash
go run cmd/submicro/main.go
```

In the second terminal from the web/ folder run:

```bash
npm install && npm run dev
```

Then open http://localhost:5173.

With Docker, run this single command from the project root:

```bash
docker compose up --build
```

Then open http://localhost:5173.

## Observability

Prometheus metrics are exposed at http://localhost:8080/metrics in development and https://submicro-production.up.railway.app/metrics in production.

| Metric | Type | Description |
|--------|------|-------------|
| `submicro_orders_processed_total` | Counter | Total orders processed by the engine |
| `submicro_trades_matched_total` | Counter | Total trades matched |
| `submicro_order_book_depth` | Gauge | Current price levels per side |
| `submicro_match_latency_nanoseconds` | Histogram | Per-order matching latency |

## Design Decisions

**Why Go?** Go's goroutine model maps perfectly to the concurrent nature of an exchange. The feed consumer, matching engine, and API server run as independent goroutines. Go is also Gemini's primary backend language.

**Why shopspring/decimal over float64?** Floating point arithmetic has rounding errors. 0.1 + 0.2 equals 0.30000000000000004 in float64. In a financial system that is catastrophic. Every price and quantity uses decimal.Decimal.

**Why RWMutex?** The feed writes to the order book while the API reads it to build snapshots for connected clients. sync.RWMutex allows many concurrent readers but only one writer, which is perfect for this access pattern.

**Known limitation:** The current slice-based price level structure has O(n) insertion. At the scale of a real exchange this would be replaced with a skip list or balanced BST for O(log n) insertion and cancellation. I discovered this empirically through benchmarking rather than theory, which is the right way to find real bottlenecks in performance-critical systems.

## What I Learned

The most surprising discovery was that the bottleneck in a naive order book implementation is not the matching algorithm but the data structure used to store price levels. I found this through benchmarking rather than theory, which is the right way to find real bottlenecks in performance-critical systems.
