FROM golang:1.25-alpine AS builder

WORKDIR /app

# Copy go mod files
COPY go.mod go.sum ./
RUN go mod download

# Copy service code
COPY services/helpdesk ./services/helpdesk

# Copy Auth Service proto (Help Desk needs it for auth client)
COPY services/auth/proto ./services/auth/proto

# Build
WORKDIR /app/services/helpdesk
RUN go build -o /app/helpdesk-service ./cmd/server

FROM alpine:latest
RUN apk --no-cache add ca-certificates
WORKDIR /root/
COPY --from=builder /app/helpdesk-service .
CMD ["./helpdesk-service"]

