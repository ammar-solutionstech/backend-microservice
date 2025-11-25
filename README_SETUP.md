# Microservices Setup Guide

## Prerequisites

1. **Go 1.25.3** or later
2. **Docker** and **Docker Compose**
3. **Protocol Buffers Compiler (protoc)**
   - Download from: https://grpc.io/docs/protoc-installation/
   - Or install via package manager:
     - Windows: `choco install protoc`
     - macOS: `brew install protobuf`
     - Linux: `apt-get install protobuf-compiler`

## Step 1: Generate Protobuf Code

Before building the services, you must generate Go code from the `.proto` files:

```bash
# Install protoc plugins
go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest

# Generate code for all services
protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    services/auth/proto/auth.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    services/helpdesk/proto/helpdesk.proto

protoc --go_out=. --go_opt=paths=source_relative \
    --go-grpc_out=. --go-grpc_opt=paths=source_relative \
    services/notification/proto/notification.proto
```

Or use the PowerShell script:
```powershell
.\generate-proto.ps1
```

**Important**: The placeholder `.pb.go` files will be overwritten with actual generated code.

## Step 2: Set Environment Variables

Create a `.env` file in the root directory:

```env
# JWT Configuration
JWT_SECRET=your-secret-key-change-in-production
JWT_ACCESS_EXPIRY=15m
JWT_REFRESH_EXPIRY=7d
JWT_ISSUER=auth-service

# RabbitMQ
RABBITMQ_URL=amqp://localhost:5672
RABBITMQ_USER=rabbitmq
RABBITMQ_PASSWORD=rabbitmq

# SMTP (for email notifications)
SMTP_HOST=smtp.gmail.com
SMTP_PORT=587
SMTP_USER=your-email@gmail.com
SMTP_PASSWORD=your-app-password
SMTP_TLS=true
```

## Step 3: Run Database Migrations

Each service has its own database. Run migrations:

```bash
# For Auth Service
psql -h localhost -U postgres -d auth_db -f services/auth/migrations/001_initial_schema.sql

# For Help Desk Service
psql -h localhost -U postgres -d helpdesk_db -f services/helpdesk/migrations/001_initial_schema.sql

# For Notification Service
psql -h localhost -U postgres -d notification_db -f services/notification/migrations/001_initial_schema.sql
```

Or use Docker Compose to start databases and run migrations manually.

## Step 4: Start Services with Docker Compose

```bash
docker-compose up -d
```

This will start:
- PostgreSQL databases (auth_db, helpdesk_db, notification_db, legacy_db)
- RabbitMQ
- All microservices (auth, helpdesk, notification, gateway)

## Step 5: Build and Run Services Locally (Alternative)

If you prefer to run services locally without Docker:

```bash
# Build Auth Service
cd services/auth
go build -o ../../bin/auth-service ./cmd/server
./bin/auth-service

# Build Help Desk Service
cd services/helpdesk
go build -o ../../bin/helpdesk-service ./cmd/server
./bin/helpdesk-service

# Build Notification Service
cd services/notification
go build -o ../../bin/notification-service ./cmd/server
./bin/notification-service

# Build Gateway
cd services/gateway
go build -o ../../bin/gateway ./cmd/server
./bin/gateway
```

## Step 6: Verify Services

Check service health:

```bash
# Gateway
curl http://localhost:8080/health

# Auth Service
curl http://localhost:8001/health

# Help Desk Service
curl http://localhost:8002/health

# Notification Service
curl http://localhost:8003/health
```

## Testing

### Test Authentication

```bash
# Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "first_name": "John",
    "latest_name": "Doe",
    "father_name": "Smith",
    "work_email": "john@example.com",
    "password": "password123",
    "work_mobile": "+1234567890"
  }'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "john@example.com",
    "password": "password123"
  }'
```

### Test Help Desk

```bash
# Create a ticket (requires authentication token)
curl -X POST http://localhost:8080/api/help-desk \
  -H "Content-Type: application/json" \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -d '{
    "name": "Laptop Issue",
    "description": "Laptop won't start",
    "help_desk_type_id": 1,
    "portal_user_id": 1,
    "state": "open"
  }'
```

## Troubleshooting

1. **Proto generation errors**: Ensure `protoc` is installed and in PATH
2. **Database connection errors**: Check Docker containers are running: `docker-compose ps`
3. **gRPC connection errors**: Ensure services are started in order (databases → services)
4. **RabbitMQ errors**: Check RabbitMQ management UI at http://localhost:15672

## Next Steps

- Complete legacy handler implementations in Gateway
- Add comprehensive error handling
- Implement request/response logging
- Add monitoring and metrics
- Set up CI/CD pipeline

