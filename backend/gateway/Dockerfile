FROM golang:1.22-alpine AS builder

WORKDIR /app

RUN apk add --no-cache git ca-certificates

COPY go.mod go.sum* ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=1.0.0" -o gateway .

FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/gateway .

RUN chown -R app:app /app

USER app

EXPOSE 8080

ENV GATEWAY_PORT=8080
ENV CORE_SERVICE_URL=http://core-service:8081
ENV KEYCLOAK_URL=http://keycloak:8080
ENV KEYCLOAK_REALM=master
ENV KEYCLOAK_CLIENT_ID=gateway

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

CMD ["sh", "-c", "./gateway"]
