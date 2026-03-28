# QRCode_Gen

A URL shortener and QR code generator service built with Go, featuring two-tier caching (Redis + SQLite), circuit breaker fault tolerance, and singleflight request deduplication.

## Architecture

```
Handler (Chi Router) → Service → Repository
                                    ├── Redis (Cache)
                                    └── SQLite (Persistent)
```

**Key patterns:**
- Clean layered architecture with dependency injection
- Composite repository: cache-aside with Redis + SQLite
- Circuit breaker (gobreaker) wrapping the database layer
- Singleflight deduplication to prevent thundering herd
- Token generation via SHA256 + Base62 encoding with collision retry

## Tech Stack

- **Go 1.25** with [Chi v5](https://github.com/go-chi/chi) router
- **SQLite** (via modernc.org/sqlite) for persistent storage
- **Redis** for caching (go-redis/v9)
- **gobreaker** for circuit breaker
- **skip2/go-qrcode** for QR image generation
- **Docker** multi-stage build + Docker Compose

## API Endpoints

| Method | Endpoint | Description |
|--------|----------|-------------|
| POST | `/v1/qr_code` | Create a QR code from a URL |
| GET | `/v1/qr_code/{qr_token}` | Resolve token to original URL |
| GET | `/v1/qr_code_image/{qr_token}` | Get QR code image (PNG) |
| PUT | `/v1/qr_code/{qr_token}` | Update URL for a token |
| DELETE | `/v1/qr_code/{qr_token}` | Delete a QR code record |
| GET | `/r/{qr_token}` | 302 redirect to original URL |
| GET | `/health` | Health check |

## Run

**Local:**

```bash
go run .
```

**Docker Compose (with Redis):**

```bash
docker-compose up --build
```

## Configuration

All settings are configurable via environment variables:

| Variable | Default | Description |
|----------|---------|-------------|
| `PORT` | `8080` | Server port |
| `BASE_URL` | `http://localhost:8080` | Base URL for short links |
| `DB_PATH` | `qrcode.db` | SQLite database path |
| `REDIS_ADDR` | `localhost:6379` | Redis address |
| `REDIS_TTL` | `24h` | Cache TTL |
| `TOKEN_LENGTH` | `8` | Generated token length |
| `TOKEN_MAX_RETRIES` | `5` | Collision retry attempts |
| `BREAKER_MAX_REQUESTS` | `3` | Circuit breaker half-open requests |
| `BREAKER_INTERVAL` | `10s` | Error count reset interval |
| `BREAKER_TIMEOUT` | `5s` | Open state timeout |
| `BREAKER_FAIL_THRESHOLD` | `5` | Failures to trip breaker |
| `DEFAULT_QR_DIMENSION` | `256` | Default QR image size (px) |
| `MAX_QR_DIMENSION` | `1024` | Max QR image size (px) |

