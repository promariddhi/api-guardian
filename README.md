# API Guardian

API Guardian is a lightweight API Gateway written in Go that provides authentication, authorization, rate limiting, load balancing, health checks, request tracing, and observability for backend services.

The project was built to explore the core components of modern API gateway architecture while keeping the implementation small and understandable.

## Features

### Reverse Proxying

* Route-based request forwarding
* Path prefix matching
* Optional path trimming before forwarding

### Authentication & Authorization

* JWT authentication
* Role-Based Access Control (RBAC)
* Protected and public routes

### Rate Limiting

* Redis-backed atomic token bucket rate limiter
* Per-IP and per-user limits
* Distributed rate limiting across gateway instances

### Load Balancing

* Round-robin backend selection
* Multiple backends per route

### Health Checks & Failure Recovery

* Automatic backend failure detection
* Background health checks
* Automatic backend recovery

### Observability

* Request tracing IDs
* Structured request logging
* Request latency metrics
* Status code metrics
* Prometheus metrics endpoint

### Reliability

* Graceful shutdown
* Redis connection cleanup
* Background worker shutdown

---

## Architecture

```text
Client
   │
   ▼
┌─────────────┐
│ API Guardian│
└──────┬──────┘
       │
       ├── Authentication
       ├── RBAC
       ├── Rate Limiting
       ├── Load Balancing
       ├── Health Checks
       ├── Metrics
       │
       ▼
 ┌──────────────┐
 │ Backend Pool │
 └──────┬───────┘
        │
 ┌──────┴──────┐
 ▼             ▼
Service A   Service B
```

---

## Configuration

Routes are configured using YAML.

Example:

```yaml
server:
  port: 8090

routes:
  - path: /auth
    trim_prefix: true
    protected: false
    backends:
      - http://localhost:8081
      - http://localhost:8082

  - path: /payments
    trim_prefix: true
    protected: true
    allowed_roles:
      - admin
    backends:
      - http://localhost:8083
      - http://localhost:8084
```

Environment-specific values are provided through environment variables.

Example `.env`:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0

JWT_SECRET=your-secret-key
```

---

## Metrics

Prometheus metrics are exposed at:

```text
/metrics
```

| Metric                                | Type      | Labels                     | Description                                                               |
| ------------------------------------- | --------- | -------------------------- | ------------------------------------------------------------------------- |
| `gateway_requests_total`              | Counter   | `method`, `path`, `status` | Total number of requests processed by the gateway.                        |
| `gateway_request_duration_seconds`    | Histogram | `method`, `path`           | Distribution of request latency in seconds.                               |
| `gateway_rate_limited_total`          | Counter   | `path`                     | Number of requests rejected by the rate limiter.                          |
| `gateway_backend_failures_total`      | Counter   | `backend`                  | Total backend failures detected by the gateway.                           |
| `gateway_backend_up`                  | Gauge     | `backend`                  | Current backend health status (`1 = healthy`, `0 = unhealthy`).           |
| `gateway_circuit_breaker_trips_total` | Counter   | `backend`                  | Number of times a backend was marked unavailable after repeated failures. |
| `gateway_backends_recovered`          | Counter   | `backend`                  | Number of times an unhealthy backend successfully recovered.              |
| `gateway_active_requests`             | Gauge     | None                       | Current number of requests being processed.                               |
| `gateway_response_bytes_total`        | Counter   | `path`                     | Total response bytes sent by the gateway.                                 |


---

## Running Locally

### Prerequisites

* Go 1.25+
* Redis

### Install Dependencies

```bash
go mod download
```

### Start Redis

```bash
docker run -p 6379:6379 redis
```

### Configure Environment

Create a `.env` file:

```env
REDIS_ADDR=localhost:6379
REDIS_PASSWORD=
REDIS_DB=0
JWT_SECRET=secret
```

### Run

```bash
go run .
```

Gateway will start on:

```text
http://localhost:8090
```

---

## Example Request

```bash
curl http://localhost:8090/auth/login
```

Example log:

```text
trace=f6b5b0b9-b0b8-4792-a45d-95b45deac1a5 GET /auth 200 54ms
```

---

## Project Goals

This project was built to gain hands-on experience with:

* Reverse proxies
* API gateways
* Load balancing
* Distributed rate limiting
* Health checking
* Failure Recovery
* Idempotent Retries
* JWT authentication
* Observability and metrics

---

## Future Improvements

* Admin dashboard
* Docker Compose deployment
* Request caching
* Least connections based load balancing
* Configurable health check

---

## License

MIT
