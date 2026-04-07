FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=1.0.0" -o core-service .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/core-service .

RUN chown -R app:app /app

USER app

EXPOSE 8081

ENV PORT=8081

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8081/health || exit 1

CMD ["sh", "-c", "./core-service"]
