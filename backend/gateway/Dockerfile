FROM golang:1.26-alpine AS builder

WORKDIR /src

RUN apk add --no-cache git ca-certificates

COPY pkg/go.mod pkg/go.sum* ./pkg/
COPY gateway/go.mod gateway/go.sum* ./gateway/
RUN printf 'go 1.26.0\n\nuse (\n\t./pkg\n\t./gateway\n)\n' > go.work \
    && cd gateway && go mod download \
    && cd ../pkg && go mod download

COPY pkg/ ./pkg/
COPY gateway/ ./gateway/
RUN cd gateway && CGO_ENABLED=0 GOOS=linux go build -ldflags="-s -w -X main.version=1.0.0" -o /app/gateway ./cmd/gateway

FROM alpine:3.21

RUN apk add --no-cache ca-certificates tzdata && \
    addgroup -S app && adduser -S app -G app

WORKDIR /app

COPY --from=builder /app/gateway .

RUN chown -R app:app /app

USER app

EXPOSE 8080

ENV GATEWAY_PORT=8080
ENV CORE_SERVICE_GRPC=core-service:9091
ENV KEYCLOAK_URL=http://keycloak:8080
ENV KEYCLOAK_REALM=master

HEALTHCHECK --interval=30s --timeout=5s --start-period=10s --retries=3 \
    CMD wget -qO- http://localhost:8080/health || exit 1

ENTRYPOINT ["./gateway"]
