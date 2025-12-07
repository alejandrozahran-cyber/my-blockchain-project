# Build stage
FROM golang:1.21-alpine AS builder

WORKDIR /app

COPY go.mod go.sum ./
RUN go mod download

COPY . .

RUN CGO_ENABLED=0 GOOS=linux go build -o nusa-node ./cmd/node

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /root/

COPY --from=builder /app/nusa-node .
COPY --from=builder /app/config ./config

EXPOSE 8545 30303

CMD ["./nusa-node", "--dev"]