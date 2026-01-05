# 1) Builder stage
FROM golang:1.23 AS builder

WORKDIR /app/pfservice

# Copy module files and download dependencies
COPY pfservice/go.mod pfservice/go.sum ./
RUN go mod download

# Copy source
COPY pfservice/ .

# Build static binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/pf-sync ./cmd/pf_sync

# 2) Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata busybox

WORKDIR /app

COPY --from=builder /app/pf-sync /app/pf-sync

# Log directory (will be mounted from host)
RUN mkdir -p /var/log && \
    touch /var/log/pf-sync.log && \
    chmod 777 /var/log /var/log/pf-sync.log

# Cron job: run daily at midnight with explicit flag
RUN echo "0 0 * * * /app/pf-sync --sync >> /var/log/pf-sync.log 2>&1" > /etc/crontabs/root

ENV TZ=Asia/Tashkent

CMD ["crond", "-f", "-d", "8"]