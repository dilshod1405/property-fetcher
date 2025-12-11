# 1) Builder stage
FROM golang:1.23 AS builder

# Work inside /app/pfservice
WORKDIR /app/pfservice

# Copy only module files first
COPY pfservice/go.mod pfservice/go.sum ./
RUN go mod download

# Now copy full project
COPY pfservice/ .

# Build binary
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o /app/pf-sync ./cmd/pf_sync



# 2) Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata busybox

WORKDIR /app

COPY --from=builder /app/pf-sync /app/pf-sync

RUN mkdir -p /var/www/mhp-api/media/property_images
RUN mkdir -p /var/log

# Cron job: run at 00:00 every day
RUN echo "0 0 * * * /app/pf-sync >> /var/log/pf-sync.log 2>&1" > /etc/crontabs/root
RUN touch /var/log/pf-sync.log

ENV TZ=Asia/Tashkent

CMD ["crond", "-f", "-d", "8"]
