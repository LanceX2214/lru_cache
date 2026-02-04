# syntax=docker/dockerfile:1
FROM golang:1.24 AS builder

WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download

COPY . .
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /out/lru_cache ./example

FROM alpine:3.20
RUN apk add --no-cache ca-certificates
WORKDIR /app
COPY --from=builder /out/lru_cache /app/lru_cache

EXPOSE 9000
ENTRYPOINT ["/app/lru_cache"]
