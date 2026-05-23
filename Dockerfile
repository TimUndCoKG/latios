# Stage 1: Build
FROM golang:1.24.0 AS builder
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

# Install ca-vertificates for tls management
RUN apk --no-cache add ca-certificates tzdata

WORKDIR /app

# Copy your Go binary
COPY --from=builder /app/latios .

EXPOSE 80 443

CMD ["./latios"]
