FROM --platform=linux/arm64 golang:1.24-alpine AS builder

WORKDIR /src

COPY go.work ./

COPY api/go.mod api/go.sum ./api/
COPY worker/go.mod worker/go.sum ./worker/
COPY shared/jobs/go.mod ./shared/jobs/

COPY shared/jobs ./shared/jobs
COPY worker ./worker

WORKDIR /src/worker

RUN go mod download

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o app-binary .

FROM --platform=linux/arm64 alpine:latest

WORKDIR /app

COPY --from=builder /src/worker/app-binary .

RUN chmod +x /app/app-binary

CMD ["/app/app-binary"]