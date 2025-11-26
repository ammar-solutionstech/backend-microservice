# Certificate Management Microservice

This microservice provides certificate lifecycle management (RA/CA/VA roles) by integrating with Smallstep step-ca. It acts as an abstraction layer and API gateway on top of step-ca to manage certificate requests, approvals, issuance, renewal, and revocation.

## Features

- **Registration Authority (RA)**: Submit, approve, and reject Certificate Signing Requests (CSRs)
- **Certificate Authority (CA)**: Issue, revoke, and renew certificates via step-ca
- **Validation Authority (VA)**: OCSP status lookup and CRL retrieval
- **REST API**: Clean REST endpoints for all certificate operations
- **PostgreSQL**: Stores CSR metadata, certificates, and revocation records
- **step-ca Integration**: Uses REST API calls (portable approach)

## Architecture

The service delegates all cryptographic operations to step-ca:
- No private keys stored in this service
- All certificate signing done by step-ca
- Service only stores metadata and orchestrates operations

## Prerequisites

1. **PostgreSQL Database**: A database named `certificate_db` must exist
2. **step-ca Instance**: Either:
   - An existing step-ca instance (configure URL in `.env`)
   - Docker-managed step-ca (optional, see Docker setup)

## Configuration

Add these environment variables to your `.env` file:

```env
# Certificate Service Database
CERTIFICATE_DB_HOST=localhost
CERTIFICATE_DB_PORT=5432
CERTIFICATE_DB_NAME=certificate_db
CERTIFICATE_DB_USER=postgres
CERTIFICATE_DB_PASSWORD=postgres

# Certificate Service Ports
CERTIFICATE_PORT=8004
CERTIFICATE_GRPC_PORT=9004

# step-ca Connection (REST API - portable approach)
STEP_CA_URL=http://step-ca:9000  # or https://your-step-ca:9000
STEP_CA_TOKEN=your-provisioner-jwt-token
STEP_CA_PROVISIONER=admin

# Optional: mTLS (if needed)
STEP_CA_USE_MTLS=false
STEP_CA_ROOT_CA=/path/to/root-ca.crt
STEP_CA_CERT=/path/to/client.crt
STEP_CA_KEY=/path/to/client.key

# Optional: Docker step-ca Management
STEP_CA_MANAGE_DOCKER=false
```

## Setup

### 1. Database Setup

Create the database and run migrations:

```bash
# Create database
psql -h localhost -U postgres -c "CREATE DATABASE certificate_db;"

# Run migrations
psql -h localhost -U postgres -d certificate_db -f services/certificate/migrations/001_initial_schema.sql
```

Or using Docker:

```bash
docker exec -i postgres-itaas psql -U postgres -d certificate_db < services/certificate/migrations/001_initial_schema.sql
```

### 2. step-ca Setup

#### Option A: Use Existing step-ca Instance

1. Ensure step-ca is running and accessible
2. Get a JWT token from your step-ca provisioner
3. Set `STEP_CA_URL` and `STEP_CA_TOKEN` in `.env`

#### Option B: Use Docker step-ca (Optional)

To start step-ca via Docker Compose:

```bash
docker-compose --profile step-ca up -d step-ca
```

**Note**: The step-ca service is in a profile, so it won't start by default. You need to explicitly enable it.

For production, you should:
1. Initialize step-ca with proper configuration
2. Store CA keys securely
3. Configure provisioners
4. Generate JWT tokens for authentication

See [step-ca documentation](https://smallstep.com/docs/step-ca) for detailed setup.

### 3. Running the Service

#### Local Development

```bash
cd services/certificate
go run ./cmd/server
```

#### Docker

```bash
docker-compose up -d certificate-service
```

## API Endpoints

### Registration Authority (RA)

- `POST /api/csr` - Submit a new CSR
- `GET /api/csr` - List CSRs (with filters: status, requester_user_id)
- `GET /api/csr/:id` - Get CSR details
- `POST /api/csr/:id/approve` - Approve a CSR
- `POST /api/csr/:id/reject` - Reject a CSR

### Certificate Authority (CA)

- `POST /api/certificates` - Issue certificate from approved CSR
- `GET /api/certificates` - List certificates (with filters: status)
- `GET /api/certificates/:serial` - Get certificate details
- `POST /api/certificates/:serial/revoke` - Revoke certificate
- `POST /api/certificates/renew` - Renew certificate

### Validation Authority (VA)

- `GET /api/ocsp/:serial` - Get OCSP status
- `GET /api/crl` - Get Certificate Revocation List
- `GET /api/validate/:serial` - Validate certificate

### Health Check

- `GET /health` - Service health check

## Example Usage

### Submit a CSR

```bash
curl -X POST http://localhost:8004/api/csr \
  -H "Content-Type: application/json" \
  -d '{
    "csr_pem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----",
    "requester_email": "user@example.com",
    "requester_user_id": 1
  }'
```

### Approve and Issue Certificate

```bash
# Approve CSR
curl -X POST http://localhost:8004/api/csr/1/approve \
  -H "Content-Type: application/json" \
  -d '{
    "approver_id": 1,
    "auto_issue": true
  }'

# Or issue separately
curl -X POST http://localhost:8004/api/certificates \
  -H "Content-Type: application/json" \
  -d '{
    "csr_id": 1
  }'
```

### Check Certificate Status

```bash
curl http://localhost:8004/api/ocsp/ABC123
```

## Integration with Gateway

To integrate with the API Gateway, add routes in `services/gateway/internal/routes/router.go`:

```go
router.Route("/api/certificates", func(r chi.Router) {
    r.Use(middleware.AuthMiddleware(clients.NewAuthClient(cfg.AuthServiceGRPC)))
    // Proxy to certificate service
})
```

## Security Notes

- **JWT Tokens**: Primary authentication method with step-ca (most portable)
- **mTLS**: Optional for additional security
- **No Private Keys**: This service never stores or handles private keys
- **Input Validation**: All CSRs are validated before processing
- **Error Handling**: Safe error messages that don't leak sensitive information

## Troubleshooting

### step-ca Connection Issues

1. Check `STEP_CA_URL` is correct
2. Verify step-ca is running: `curl http://step-ca:9000/health`
3. Check JWT token is valid
4. Review service logs: `docker-compose logs certificate-service`

### Database Connection Issues

1. Verify database exists: `psql -l | grep certificate_db`
2. Check migrations ran successfully
3. Verify connection credentials in `.env`

### Certificate Operations Fail

1. Ensure step-ca is accessible
2. Check JWT token has proper permissions
3. Verify CSR format is valid
4. Review step-ca logs for detailed errors

## Development

### Project Structure

```
services/certificate/
├── cmd/server/main.go          # Entry point
├── internal/
│   ├── config/                  # Configuration
│   ├── models/                  # Database models
│   ├── services/                # Business logic
│   │   ├── stepca_client.go    # step-ca HTTP client
│   │   ├── ra_service.go       # RA operations
│   │   ├── ca_service.go       # CA operations
│   │   └── va_service.go       # VA operations
│   ├── routes/                  # HTTP controllers
│   └── middleware/              # HTTP middleware
├── migrations/                  # Database migrations
└── Dockerfile                   # Docker build
```

### Testing

```bash
# Test compilation
go build ./services/certificate/cmd/server

# Run tests (when implemented)
go test ./services/certificate/...
```

## References

- [Smallstep step-ca Documentation](https://smallstep.com/docs/step-ca)
- [step-ca REST API](https://smallstep.com/docs/step-ca/certificate-authority-server-api)
- [RFC 5280 - X.509 Certificate Profile](https://tools.ietf.org/html/rfc5280)


