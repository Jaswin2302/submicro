import { useEffect, useState, useRef } from "react";

interface PriceLevel {
  price: string;
  quantity: string;
}

interface BookSnapshot {
  bids: PriceLevel[];
  asks: PriceLevel[];
}

interface Trade {
  price: string;
  quantity: string;
  side: string;
  time: string;
}

interface Message {
  type: string;
  payload: any;
}

const MAX_LEVELS = 15;
const MAX_TRADES = 50;
const WS_URL = "ws://localhost:8080/ws";

export default function App() {
  const [bids, setBids] = useState<PriceLevel[]>([]);
  const [asks, setAsks] = useState<PriceLevel[]>([]);
  const [trades, setTrades] = useState<Trade[]>([]);
  const [connected, setConnected] = useState(false);
  const wsRef = useRef<WebSocket | null>(null);

  useEffect(() => {
    function connect() {
      const ws = new WebSocket(WS_URL);
      wsRef.current = ws;

      ws.onopen = () => setConnected(true);

      ws.onclose = () => {
        setConnected(false);
        setTimeout(connect, 3000);
      };

      ws.onerror = () => ws.close();

      ws.onmessage = (event) => {
        const msg: Message = JSON.parse(event.data);

        if (msg.type === "snapshot" || msg.type === "book_update") {
          setBids(msg.payload.bids.slice(0, MAX_LEVELS));
          setAsks(msg.payload.asks.slice(0, MAX_LEVELS));
        }

        if (msg.type === "trade") {
          setTrades((prev) => [
            { ...msg.payload, time: new Date().toLocaleTimeString() },
            ...prev.slice(0, MAX_TRADES - 1),
          ]);
        }
      };
    }

    connect();
    return () => wsRef.current?.close();
  }, []);

  const maxBidQty = Math.max(...bids.map((b) => parseFloat(b.quantity)), 1);
  const maxAskQty = Math.max(...asks.map((a) => parseFloat(a.quantity)), 1);
  const bestBid = bids[0]?.price ?? "—";
  const bestAsk = asks[0]?.price ?? "—";
  const spread =
    bids[0] && asks[0]
      ? (parseFloat(asks[0].price) - parseFloat(bids[0].price)).toFixed(2)
      : "—";

  return (
    <div style={{ background: "#0f1117", minHeight: "100vh", color: "#e2e8f0", fontFamily: "monospace", padding: "24px" }}>

      <div style={{ display: "flex", alignItems: "center", gap: "12px", marginBottom: "24px" }}>
        <h1 style={{ margin: 0, fontSize: "20px", fontWeight: 600, color: "#f8fafc" }}>SubMicro</h1>
        <span style={{ fontSize: "12px", color: "#64748b" }}>BTC/USD Live Order Book</span>
        <div style={{ marginLeft: "auto", display: "flex", alignItems: "center", gap: "6px" }}>
          <div style={{ width: "8px", height: "8px", borderRadius: "50%", background: connected ? "#22c55e" : "#ef4444" }} />
          <span style={{ fontSize: "12px", color: connected ? "#22c55e" : "#ef4444" }}>
            {connected ? "Live" : "Disconnected"}
          </span>
        </div>
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", gap: "12px", marginBottom: "24px" }}>
        {[
          { label: "Best Bid", value: `$${parseFloat(bestBid).toFixed(2)}`, color: "#22c55e" },
          { label: "Best Ask", value: `$${parseFloat(bestAsk).toFixed(2)}`, color: "#ef4444" },
          { label: "Spread", value: `$${spread}`, color: "#f8fafc" },
        ].map((stat) => (
          <div key={stat.label} style={{ background: "#1e2130", borderRadius: "8px", padding: "16px" }}>
            <div style={{ fontSize: "11px", color: "#64748b", marginBottom: "4px" }}>{stat.label}</div>
            <div style={{ fontSize: "20px", fontWeight: 600, color: stat.color }}>{stat.value}</div>
          </div>
        ))}
      </div>

      <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr", gap: "16px" }}>

        <div style={{ background: "#1e2130", borderRadius: "8px", padding: "16px" }}>
          <div style={{ fontSize: "13px", fontWeight: 600, color: "#f8fafc", marginBottom: "12px" }}>Order Book</div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", fontSize: "11px", color: "#64748b", marginBottom: "8px", paddingBottom: "8px", borderBottom: "1px solid #2d3148" }}>
            <span>Price (USD)</span>
            <span style={{ textAlign: "right" }}>Quantity (BTC)</span>
            <span style={{ textAlign: "right" }}>Depth</span>
          </div>

          {asks.slice().reverse().map((level, i) => (
            <div key={i} style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", fontSize: "12px", padding: "3px 0", position: "relative" }}>
              <div style={{ position: "absolute", right: 0, top: 0, bottom: 0, background: "rgba(239,68,68,0.08)", width: `${(parseFloat(level.quantity) / maxAskQty) * 100}%` }} />
              <span style={{ color: "#ef4444", position: "relative" }}>{parseFloat(level.price).toFixed(2)}</span>
              <span style={{ textAlign: "right", position: "relative" }}>{parseFloat(level.quantity).toFixed(6)}</span>
              <span style={{ textAlign: "right", color: "#64748b", position: "relative" }}>{((parseFloat(level.quantity) / maxAskQty) * 100).toFixed(1)}%</span>
            </div>
          ))}

          <div style={{ borderTop: "1px solid #2d3148", borderBottom: "1px solid #2d3148", padding: "8px 0", margin: "8px 0", display: "flex", justifyContent: "space-between" }}>
            <span style={{ fontSize: "11px", color: "#64748b" }}>Spread</span>
            <span style={{ fontSize: "13px", fontWeight: 600, color: "#f8fafc" }}>${spread}</span>
          </div>

          {bids.map((level, i) => (
            <div key={i} style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", fontSize: "12px", padding: "3px 0", position: "relative" }}>
              <div style={{ position: "absolute", right: 0, top: 0, bottom: 0, background: "rgba(34,197,94,0.08)", width: `${(parseFloat(level.quantity) / maxBidQty) * 100}%` }} />
              <span style={{ color: "#22c55e", position: "relative" }}>{parseFloat(level.price).toFixed(2)}</span>
              <span style={{ textAlign: "right", position: "relative" }}>{parseFloat(level.quantity).toFixed(6)}</span>
              <span style={{ textAlign: "right", color: "#64748b", position: "relative" }}>{((parseFloat(level.quantity) / maxBidQty) * 100).toFixed(1)}%</span>
            </div>
          ))}
        </div>

        <div style={{ background: "#1e2130", borderRadius: "8px", padding: "16px" }}>
          <div style={{ fontSize: "13px", fontWeight: 600, color: "#f8fafc", marginBottom: "12px" }}>Trade Tape</div>
          <div style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", fontSize: "11px", color: "#64748b", marginBottom: "8px", paddingBottom: "8px", borderBottom: "1px solid #2d3148" }}>
            <span>Price</span>
            <span style={{ textAlign: "right" }}>Quantity</span>
            <span style={{ textAlign: "right" }}>Time</span>
          </div>
          <div style={{ overflowY: "auto", maxHeight: "400px" }}>
            {trades.length === 0 ? (
              <div style={{ fontSize: "12px", color: "#64748b", textAlign: "center", padding: "20px 0" }}>
                Waiting for trades...
              </div>
            ) : (
              trades.map((trade, i) => (
                <div key={i} style={{ display: "grid", gridTemplateColumns: "1fr 1fr 1fr", fontSize: "12px", padding: "3px 0", borderBottom: "1px solid #1a1f2e" }}>
                  <span style={{ color: trade.side === "bid" ? "#22c55e" : "#ef4444" }}>
                    ${parseFloat(trade.price).toFixed(2)}
                  </span>
                  <span style={{ textAlign: "right" }}>{parseFloat(trade.quantity).toFixed(6)}</span>
                  <span style={{ textAlign: "right", color: "#64748b" }}>{trade.time}</span>
                </div>
              ))
            )}
          </div>
        </div>
      </div>

      <div style={{ marginTop: "16px", fontSize: "11px", color: "#334155", textAlign: "center" }}>
        Engine: 1.2M orders/sec · Sub-microsecond matching · Fed by Gemini live WebSocket API
      </div>
    </div>
  );
}