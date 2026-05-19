# ONDC Analytics Gateway

> Real-time seller analytics platform for the ONDC network with PII-safe data processing, edge-computed redaction, and high-throughput event streaming.

## Architecture

```
┌─────────────┐     ┌──────────────────┐     ┌──────────────┐
│ Load Tester  │────▶│  Gateway Ingress  │────▶│   Redpanda   │
│  (Node.js)   │     │   (Go + Chi)     │     │  (Streaming) │
└─────────────┘     └────────┬─────────┘     └──────┬───────┘
                             │ gRPC/UDS              │ Kafka Engine
                    ┌────────▼─────────┐     ┌──────▼───────┐
                    │  Edge Sandbox    │     │  ClickHouse  │
                    │ (Rust+Wasmtime)  │     │   (OLAP)     │
                    └──────────────────┘     └──────┬───────┘
                                                     │ SQL
                                             ┌──────▼───────┐
                                             │ Analytics API │
                                             │  (Go + Chi)  │
                                             └──────┬───────┘
                                                     │ REST
                                             ┌──────▼───────┐
                                             │  Dashboard   │
                                             │ (React+Vite) │
                                             └──────────────┘
```

## Quick Start

```bash
# Clone and start everything
git clone https://github.com/amitav400c/ondc-analytics-gateway.git
cd ondc-analytics-gateway
cp .env.example .env
docker-compose up -d

# Wait for services to be healthy (~30s)
docker-compose ps

# Open the dashboard
open http://localhost:5173

# Login: admin@ondc-analytics.dev / admin123

# Generate test traffic
docker-compose --profile test up load-tester
```

## Tech Stack

| Layer | Technology | Purpose |
|-------|-----------|---------|
| **Frontend** | React 18, TypeScript, Vite, Tailwind, Recharts | Dashboard UI |
| **API** | Go, Chi, JWT, clickhouse-go | REST analytics endpoints |
| **Ingestion** | Go, Chi, kafka-go | Webhook receiver + Redpanda producer |
| **Edge Compute** | Rust, Wasmtime, Tonic gRPC | PII redaction sandbox (UDS) |
| **Streaming** | Redpanda | Kafka-compatible event bus |
| **OLAP** | ClickHouse | Columnar analytics (MergeTree + Kafka Engine) |
| **Auth** | PostgreSQL, bcrypt, JWT | User authentication |
| **DevOps** | Docker, Docker Compose | Single-command deployment |

## API Endpoints

| Method | Endpoint | Auth | Description |
|--------|----------|------|-------------|
| `POST` | `/webhooks/ondc` | No | Beckn payload ingestion |
| `GET` | `/api/v1/health` | No | System health check |
| `POST` | `/api/v1/auth/login` | No | JWT authentication |
| `GET` | `/api/v1/metrics/funnel` | JWT | Order funnel (Search→Confirm) |
| `GET` | `/api/v1/metrics/cancellations?city=BLR` | JWT | Cancellation breakdown |
| `GET` | `/api/v1/metrics/volume?days=7` | JWT | Daily order volume |
| `GET` | `/api/v1/events/recent?limit=20` | JWT | Recent sanitized transactions |

## PII Redaction

All incoming payloads pass through the Edge Sandbox before storage:
- **Phone numbers** → SHA256 hash (truncated to 16 chars)
- **GPS coordinates** → Fuzzed ±0.01° (~1km radius)
- **Raw PII** → Never reaches ClickHouse or the dashboard

## Project Structure

```
├── docker-compose.yml          # Full stack orchestration
├── gateway-ingress/            # Go: HTTP webhook receiver + Redpanda producer
├── edge-sandbox/               # Rust: gRPC PII redaction via Wasmtime
├── analytics-api/              # Go: REST API querying ClickHouse
├── frontend-dashboard/         # React: Seller analytics dashboard
├── load-tester/                # Node.js: Simulated Beckn traffic
├── clickhouse/                 # SQL: DDL + Kafka Engine setup
├── postgres/                   # SQL: Auth schema + seed data
└── docs/                       # Architecture documentation
```

## Development

```bash
# Run Go services locally
cd gateway-ingress && go run ./cmd/server
cd analytics-api && go run ./cmd/server

# Run Rust sandbox locally
cd edge-sandbox && cargo run

# Run frontend dev server
cd frontend-dashboard && npm install && npm run dev

# Run load tester
cd load-tester && node src/index.js
```

## Environment Variables

See [.env.example](.env.example) for all configurable variables.

## References

- [ONDC Protocol Specs](https://ondc.org)
- [Beckn Protocol](https://becknprotocol.io)
- [Vortex WASM Runtime](https://github.com/amitav400c/Vortex_wasm-edge-runtime) — upstream WASM engine

## License

MIT
