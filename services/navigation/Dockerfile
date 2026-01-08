FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/navigation ./services/navigation

# Build
WORKDIR /app/services/navigation
RUN go build -o /app/navigation-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /app
# Create certs directory
RUN mkdir -p /app/certs
COPY --from=builder /app/navigation-service .
CMD ["./navigation-service"]

