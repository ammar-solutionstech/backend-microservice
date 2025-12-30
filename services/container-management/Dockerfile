# Build stage
FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/container-management ./services/container-management

# Build the service
WORKDIR /app/services/container-management
RUN CGO_ENABLED=0 GOOS=linux go build -a -installsuffix cgo -o server ./cmd/server

# Runtime stage
FROM alpine:latest

RUN apk --no-cache add ca-certificates

WORKDIR /app

# Create certs directory for mTLS certificates
RUN mkdir -p /app/certs

# Copy the binary
COPY --from=builder /app/services/container-management/server .

EXPOSE 8005 9005

CMD ["./server"]

