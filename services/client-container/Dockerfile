# Multi-stage build for client-container service

# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/client-container ./services/client-container

# Build the binary
WORKDIR /app/services/client-container
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o client-container ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates tzdata

WORKDIR /root/

# Copy the binary from builder
COPY --from=builder /app/services/client-container/client-container .

# Create directories for certificates and data
RUN mkdir -p /certs /data

# Expose ports (REST and gRPC)
EXPOSE 8006 9006

# Set entrypoint
ENTRYPOINT ["./client-container"]

