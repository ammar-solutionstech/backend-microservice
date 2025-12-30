FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/gateway ./services/gateway

# Copy proto directories from all services (Gateway needs them)
COPY services/auth/proto ./services/auth/proto
COPY services/helpdesk/proto ./services/helpdesk/proto
COPY services/notification/proto ./services/notification/proto

# Build
WORKDIR /app/services/gateway
RUN go build -o /app/gateway ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Create certs directory
RUN mkdir -p /app/certs
COPY --from=builder /app/gateway .
CMD ["./gateway"]

