# Go Observability Lab

A small production-style observability playground built with Go, Docker Compose, Prometheus, Grafana, and k6.

This project demonstrates foundational backend/platform engineering concepts including:
- HTTP middleware
- Prometheus instrumentation
- request latency tracking
- structured request logging
- request correlation IDs
- containerized services
- Grafana dashboards
- load testing
- operational observability patterns

The goal of this project was to better understand how modern distributed systems and production services are monitored and debugged.

---

# Architecture

```text
                +------------------+
                |      k6          |
                |  Load Testing    |
                +--------+---------+
                         |
                         v
                +------------------+
                |   Go API Server  |
                |                  |
                |  /work           |
                |  /metrics        |
                |  /healthz        |
                +--------+---------+
                         |
                         v
                +------------------+
                |   Prometheus     |
                | Metrics Storage  |
                +--------+---------+
                         |
                         v
                +------------------+
                |     Grafana      |
                | Dashboards/UI    |
                +------------------+
```

---

# Features

## Go HTTP Service

- Go HTTP server
- middleware-based observability
- simulated workload endpoint
- health check endpoint
- Prometheus metrics endpoint

## Prometheus Metrics

Custom metrics include:
- request count
- request duration
- HTTP status labels
- request throughput
- latency histograms

## Request Logging

Middleware-based request logging including:
- HTTP method
- request path
- status code
- request duration
- request correlation ID

## Request Correlation IDs

Each request receives a UUID-based request ID:
- returned in response headers
- included in request logs
- useful for tracing requests across services

## Grafana Dashboards

Dashboards visualize:
- requests per second
- 500 error rates
- P95 latency
- service behavior under load

## Load Testing

k6 is used to simulate concurrent traffic and stress the service under load.

---

# Tech Stack

| Component | Technology |
|---|---|
| API Service | Go |
| Metrics | Prometheus |
| Dashboards | Grafana |
| Containerization | Docker Compose |
| Load Testing | k6 |

---

# Endpoints

| Endpoint | Purpose |
|---|---|
| `/healthz` | Health check |
| `/work` | Simulated workload |
| `/metrics` | Prometheus metrics |

---

# Running Locally

## Start Services

```bash
docker compose up --build
```

Services:
- API: `localhost:8080`
- Prometheus: `localhost:9090`
- Grafana: `localhost:3000`

---

# Grafana Setup

Default credentials:

```text
admin / admin
```

Add Prometheus as a data source:

```text
http://prometheus:9090
```

---

# Load Testing

Install k6:

```bash
brew install k6
```

Run load test:

```bash
k6 run load-test/work.js
```

---

# Example PromQL Queries

## Requests Per Second

```promql
rate(http_requests_total[1m])
```

## 500 Errors

```promql
rate(http_requests_total{status="500"}[1m])
```

## P95 Latency

```promql
histogram_quantile(
  0.95,
  rate(http_request_duration_seconds_bucket[1m])
)
```

---

# Observability Concepts Demonstrated

## Middleware

Cross-cutting request instrumentation implemented using Go HTTP middleware.

## Metrics vs Logs

Metrics answer:
- How many requests failed?
- How slow is the service?
- How much traffic exists?

Logs answer:
- Which specific request failed?
- What happened during execution?

## Request Correlation IDs

Request IDs allow tracing a request through multiple services in distributed systems.

## Prometheus

Prometheus scrapes and stores time-series metrics exposed by the Go service.

## Grafana

Grafana visualizes Prometheus metrics using PromQL-powered dashboards.

---

# Future Improvements

Potential future expansions:
- Redis-backed worker queues
- asynchronous job processing
- OpenTelemetry tracing
- Kubernetes deployment
- Helm charts
- horizontal scaling
- distributed worker services
- ML inference workers
- alerting pipelines

---

# Screenshots

Add screenshots here:
- Grafana dashboard
- Prometheus targets page
- k6 load test output
- request logs

---

# Learning Goals

This project was built to gain hands-on experience with:
- backend observability
- production service instrumentation
- infrastructure tooling
- monitoring systems
- distributed systems concepts
- platform engineering fundamentals
- production debugging workflows
