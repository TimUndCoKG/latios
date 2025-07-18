FROM golang:1.23.2 AS builder

WORKDIR /app

COPY go.mod ./
COPY go.sum ./
RUN go mod download

COPY . ./

RUN CGO_ENABLED=0 GOOS=linux go build -o latios main.go

# ---

FROM alpine:3.20.0

WORKDIR /app
RUN apk upgrade --no-cache
COPY --from=builder /app/latios .
COPY .certs ./.certs

EXPOSE 80
EXPOSE 443

CMD ["./latios"]
