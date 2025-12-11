# 1) Builder stage
FROM golang:1.23 AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

# Build from the correct directory!
RUN CGO_ENABLED=0 GOOS=linux GOARCH=amd64 go build -o pf-sync ./cmd/pf_sync



# 2) Runtime stage
FROM alpine:3.19

RUN apk add --no-cache ca-certificates tzdata busybox

WORKDIR /app

COPY --from=builder /app/pf-sync /app/pf-sync

RUN mkdir -p /var/www/mhp-api/media/property_images
RUN mkdir -p /var/log

RUN echo "0 0 * * * /app/pf-sync >> /var/log/pf-sync.log 2>&1" > /etc/crontabs/root
RUN touch /var/log/pf-sync.log

ENV TZ=Asia/Tashkent

CMD ["crond", "-f", "-d", "8"]
