FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/inventory ./services/inventory

# Build
WORKDIR /app/services/inventory
RUN go build -o /app/inventory-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Create certs directory
RUN mkdir -p /app/certs
COPY --from=builder /app/inventory-service .
CMD ["./inventory-service"]

