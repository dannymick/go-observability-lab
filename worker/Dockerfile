FROM --platform=linux/arm64 golang:1.24-alpine AS builder

WORKDIR /src

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux GOARCH=arm64 go build -o app-binary .

FROM --platform=linux/arm64 alpine:latest

WORKDIR /app

COPY --from=builder /src/app-binary .

RUN chmod +x /app/app-binary

CMD ["/app/app-binary"]