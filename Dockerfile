# Stage 1: Build
FROM golang:1.23.2 AS builder
WORKDIR /app

# Copy and download dependencies
COPY go.mod go.sum ./
RUN go mod download

# Copy source
COPY . .

# Build
RUN CGO_ENABLED=0 GOOS=linux go build -o latios main.go

# Stage 2: Run
FROM alpine:latest
RUN apk --no-cache add ca-certificates
RUN apk add --no-cache gcc musl-dev
WORKDIR /app

# Install certbot with Cloudflare DNS plugin
RUN apk update && apk upgrade --no-cache && \
    apk add --no-cache certbot tzdata curl bash py3-pip && \
    pip install --no-cache-dir certbot-dns-cloudflare && \
    mkdir -p /var/www/html

# Copy built binary
COPY --from=builder /app/latios .

# Expose HTTP and HTTPS ports
EXPOSE 80 443

# Run
CMD ["./latios"]
