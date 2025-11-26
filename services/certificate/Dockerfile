FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/certificate ./services/certificate

# Build
WORKDIR /app/services/certificate
RUN go build -o /app/certificate-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/certificate-service .
CMD ["./certificate-service"]


