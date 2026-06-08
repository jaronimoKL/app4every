# syntax=docker/dockerfile:1
FROM golang:1.21-alpine AS dev
WORKDIR /app
RUN go install github.com/cosmtrek/air@v1.49.0
COPY go.mod ./
RUN go mod download
COPY . .
CMD ["air", "-c", ".air.toml"]

FROM golang:1.21-alpine AS builder
WORKDIR /app
COPY go.mod ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /screenshare ./cmd/server

FROM alpine:3.19 AS prod
COPY --from=builder /screenshare /screenshare
EXPOSE 8083
CMD ["/screenshare"]
