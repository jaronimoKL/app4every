# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS dev

WORKDIR /app

# Устанавливаем air для hot-reload в режиме разработки
RUN go install github.com/cosmtrek/air@v1.49.0

COPY go.mod go.sum ./
RUN go mod download

COPY . .

CMD ["air", "-c", ".air.toml"]

# ── Продакшн-образ (будет добавлен позже) ──
FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /notebook ./cmd/server

FROM alpine:3.19 AS prod
COPY --from=builder /notebook /notebook
EXPOSE 8081
CMD ["/notebook"]
