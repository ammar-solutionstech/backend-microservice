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
# Or use the PowerShell script (recommended):
.\generate-proto.ps1

# Or use Makefile:
make proto

# The script generates code for:
# - Auth Service
# - Help Desk Service
# - Notification Service
# - Inventory Service (NEW)
# - Geography Service (NEW)
# - Navigation Service (NEW)
# - Client Container Service
# - Agent

**Note**: Proto code has already been generated. Re-run if you modify `.proto` files.

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

# For Inventory Service (NEW)
psql -h localhost -U postgres -d inventory_db -f services/inventory/migrations/001_initial_schema.sql

# For Geography Service (NEW)
psql -h localhost -U postgres -d geography_db -f services/geography/migrations/001_initial_schema.sql

# For Navigation Service (NEW)
psql -h localhost -U postgres -d navigation_db -f services/navigation/migrations/001_initial_schema.sql
```

Or use Docker Compose to start databases and run migrations manually.

**Note**: After migrations, you can migrate data from the monolithic database using scripts in `extra/` directory. See `extra/DATA_MIGRATION_README.md` for details.

## Step 4: Start Services with Docker Compose

```bash
docker-compose up -d
```

This will start:
- PostgreSQL databases (auth_db, helpdesk_db, notification_db, inventory_db, geography_db, navigation_db)
- RabbitMQ
- All microservices (auth, helpdesk, notification, inventory, geography, navigation, gateway)

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

# Build Inventory Service (NEW)
cd services/inventory
go build -o ../../bin/inventory-service ./cmd/server
./bin/inventory-service

# Build Geography Service (NEW)
cd services/geography
go build -o ../../bin/geography-service ./cmd/server
./bin/geography-service

# Build Navigation Service (NEW)
cd services/navigation
go build -o ../../bin/navigation-service ./cmd/server
./bin/navigation-service

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

# Inventory Service (NEW)
curl http://localhost:8007/health

# Geography Service (NEW)
curl http://localhost:8008/health

# Navigation Service (NEW)
curl http://localhost:8009/health
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

