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

RUN apk add --no-cache \
    ca-certificates \
    gcc \
    musl-dev \
    python3 \
    py3-pip \
    libffi-dev \
    openssl-dev \
    bash \
    curl \
    tzdata

# Create and activate virtual environment, then install certbot inside it
RUN python3 -m venv /opt/venv \
    && /opt/venv/bin/pip install --upgrade pip \
    && /opt/venv/bin/pip install certbot certbot-dns-cloudflare \
    && mkdir -p /var/www/html

ENV PATH="/opt/venv/bin:$PATH"

WORKDIR /app

# Copy your Go binary
COPY --from=builder /app/latios .

EXPOSE 80 443

CMD ["./latios"]
