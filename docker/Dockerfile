FROM golang:1.22-alpine AS builder
WORKDIR /src
COPY go.mod go.sum ./
RUN go mod download
COPY . .
RUN CGO_ENABLED=0 GOOS=linux go build -o /out/mikrotik-collector ./cmd/collector

FROM alpine:3.20
RUN adduser -D -H appuser
USER appuser
COPY --from=builder /out/mikrotik-collector /usr/local/bin/mikrotik-collector
COPY configs/config.example.yaml /etc/mikrotik-collector/config.yaml
EXPOSE 9107
ENTRYPOINT ["/usr/local/bin/mikrotik-collector"]
CMD ["--config", "/etc/mikrotik-collector/config.yaml"]
