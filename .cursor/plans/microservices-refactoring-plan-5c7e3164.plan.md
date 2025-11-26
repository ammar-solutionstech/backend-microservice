<!-- 5c7e3164-11e7-4830-8ec9-784606477e7f 095031e8-2827-4942-a317-50544b63d62e -->
# Certificate Management Microservice Implementation Plan

## Overview

Create a new `services/certificate` microservice that acts as an abstraction layer over Smallstep step-ca, handling certificate requests, approvals, issuance, renewal, and revocation. The service follows the same structure as existing microservices and uses REST API calls to step-ca (most portable approach for future CA changes).

## Project Structure

```
services/certificate/
├── cmd/server/main.go
├── Dockerfile
├── internal/
│   ├── config/config.go
│   ├── models/
│   │   ├── csr.go
│   │   ├── certificate.go
│   │   └── revocation.go
│   ├── services/
│   │   ├── stepca_client.go          # HTTP REST API client wrapper
│   │   ├── ra_service.go             # Registration Authority logic
│   │   ├── ca_service.go             # Certificate Authority operations
│   │   ├── va_service.go             # Validation Authority operations
│   │   └── grpc_server.go           # gRPC server (optional)
│   ├── routes/
│   │   ├── ra_controller.go          # RA endpoints
│   │   ├── ca_controller.go          # CA endpoints
│   │   └── va_controller.go          # VA endpoints
│   └── middleware/
│       ├── cors_middleware.go
│       └── jwt_middleware.go         # Reuse from auth service pattern
├── migrations/
│   └── 001_initial_schema.sql
└── proto/                            # Optional: for gRPC
    └── certificate.proto
```

## Key Components

### 1. Configuration (`internal/config/config.go`)

- Database connection (PostgreSQL)
- Server ports (REST and gRPC)
- step-ca connection settings:
  - `STEP_CA_URL` (e.g., `https://step-ca:9000` or `http://localhost:9000`)
  - `STEP_CA_TOKEN` (JWT token from provisioner - primary authentication)
  - `STEP_CA_PROVISIONER` (provisioner name)
  - `STEP_CA_USE_MTLS` (boolean: enable mTLS if certificates provided)
  - `STEP_CA_ROOT_CA` (optional: path to root CA cert for mTLS verification)
  - `STEP_CA_CERT` (optional: client cert path for mTLS)
  - `STEP_CA_KEY` (optional: client key path for mTLS)
  - `STEP_CA_MANAGE_DOCKER` (boolean: start step-ca via Docker if not running)
- JWT settings (for service authentication via Gateway)

### 2. Database Models (`internal/models/`)

**CSR Model** (`csr.go`):

- ID, CSR content (PEM), status (pending/approved/rejected), requester info (user_id, email), created_at, updated_at, approved_by, rejection_reason

**Certificate Model** (`certificate.go`):

- ID, serial number, CSR ID (FK), certificate (PEM), issued_at, expires_at, status (active/revoked/expired), step_ca_cert_id

**Revocation Model** (`revocation.go`):

- ID, certificate_id (FK), serial number, revoked_at, reason, revoked_by

### 3. step-ca Client (`internal/services/stepca_client.go`)

HTTP REST API client wrapper (portable, works with any CA):

- Connects to step-ca REST API endpoints (standard HTTP/HTTPS)
- Primary authentication: JWT tokens from provisioners (most common, portable)
- Optional: mTLS support if certificates configured
- Provides methods:
  - `SignCSR(csrPEM string) (*Certificate, error)`
  - `GetCertificate(serial string) (*Certificate, error)`
  - `RevokeCertificate(serial string, reason int) error`
  - `RenewCertificate(serial string) (*Certificate, error)`
  - `GetCRL() ([]byte, error)`
  - `GetOCSPStatus(serial string) (*OCSPResponse, error)`
  - `ListCertificates(filters) ([]Certificate, error)`

Uses step-ca's REST API (standard endpoints, portable to other CAs):

- `POST /1.0/sign` - Sign CSR (with JWT token in Authorization header)
- `GET /1.0/certificates/{serial}` - Get certificate
- `POST /1.0/revoke` - Revoke certificate
- `GET /1.0/crl` - Get CRL
- `GET /1.0/certificates` - List certificates

Authentication: JWT token sent in `Authorization: Bearer <token>` header

### 4. Service Layer

**RA Service** (`ra_service.go`):

- `SubmitCSR(csrPEM, requesterInfo) (*CSR, error)` - Store CSR, validate PEM format
- `ApproveCSR(csrID, approverID) error` - Approve and forward to step-ca
- `RejectCSR(csrID, reason) error` - Reject CSR
- `GetCSR(id) (*CSR, error)` - Retrieve CSR status

**CA Service** (`ca_service.go`):

- `IssueCertificate(csrID) (*Certificate, error)` - Call step-ca to sign approved CSR
- `GetCertificate(serial) (*Certificate, error)` - Get from DB or step-ca
- `RevokeCertificate(serial, reason) error` - Revoke via step-ca
- `RenewCertificate(serial) (*Certificate, error)` - Renew via step-ca
- `ListCertificates(filters) ([]Certificate, error)` - List from DB

**VA Service** (`va_service.go`):

- `GetOCSPStatus(serial) (*OCSPStatus, error)` - Query step-ca or DB cache
- `GetCRL() ([]byte, error)` - Fetch CRL from step-ca
- `ValidateCertificate(serial) (bool, error)` - Check if valid/revoked

### 5. Controllers (`internal/routes/`)

**RA Controller** (`ra_controller.go`):

- `POST /api/csr` - Submit CSR
- `GET /api/csr/:id` - Get CSR status
- `POST /api/csr/:id/approve` - Approve CSR
- `POST /api/csr/:id/reject` - Reject CSR

**CA Controller** (`ca_controller.go`):

- `POST /api/certificates` - Issue certificate (from approved CSR)
- `GET /api/certificates/:serial` - Get certificate details
- `POST /api/certificates/:serial/revoke` - Revoke certificate
- `POST /api/certificates/renew` - Renew certificate
- `GET /api/certificates` - List certificates

**VA Controller** (`va_controller.go`):

- `GET /api/ocsp/:serial` - OCSP status lookup
- `GET /api/crl` - Get CRL

### 6. Database Migration (`migrations/001_initial_schema.sql`)

Create tables:

- `csr_requests` - Store CSR submissions
- `certificates` - Store issued certificates metadata
- `revocations` - Store revocation records

### 7. Main Server (`cmd/server/main.go`)

- Initialize config
- Connect to database
- Check step-ca connectivity (start Docker instance if `STEP_CA_MANAGE_DOCKER=true`)
- Initialize step-ca client
- Initialize services (RA, CA, VA)
- Register routes
- Start REST server (and optionally gRPC)
- Graceful shutdown

### 8. Docker Integration

- Add `step-ca` service to `docker-compose.yml` (optional, can use existing)
- Configure step-ca with basic setup (if managed by Docker)
- Add `certificate-service` to docker-compose
- Service supports both:
  - Connecting to existing step-ca instance (default)
  - Starting step-ca via Docker if `STEP_CA_MANAGE_DOCKER=true` and step-ca not reachable
- Update Gateway routes (optional - service works standalone by default)

### 9. Environment Variables

Add to `.env`:

```
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
STEP_CA_URL=https://step-ca:9000
STEP_CA_TOKEN=your-provisioner-jwt-token
STEP_CA_PROVISIONER=your-provisioner-name

# Optional: mTLS (if needed)
STEP_CA_USE_MTLS=false
STEP_CA_ROOT_CA=/path/to/root-ca.crt
STEP_CA_CERT=/path/to/client.crt
STEP_CA_KEY=/path/to/client.key

# Optional: Docker step-ca Management
STEP_CA_MANAGE_DOCKER=false
```

## Implementation Steps

1. Create folder structure and base files
2. Implement config with step-ca settings (JWT primary, mTLS optional)
3. Create database models and migration
4. Implement step-ca HTTP REST API client wrapper
5. Implement RA service (CSR management)
6. Implement CA service (certificate operations)
7. Implement VA service (validation/OCSP)
8. Create controllers for each role
9. Set up routes and middleware
10. Create main server entry point with step-ca connectivity check
11. Add Dockerfile and docker-compose integration (optional step-ca)
12. Create README with setup instructions for both existing and Docker-managed step-ca

## Notes

- All cryptographic operations delegated to step-ca
- Service only stores metadata and orchestrates operations
- Use existing middleware patterns (CORS, JWT auth)
- Follow same error handling and logging patterns
- No private keys stored in this service
- **Portable Design**: Uses standard REST API calls (works with step-ca and other CAs)
- **Authentication**: JWT tokens (primary, most portable), mTLS (optional)
- **Deployment**: Standalone by default, Gateway integration optional
- **step-ca Management**: Supports both existing instances and Docker-managed

### To-dos

- [ ] Create folder structure and base configuration file with step-ca connection settings
- [ ] Create database models (CSR, Certificate, Revocation) and SQL migration
- [ ] Implement step-ca HTTP client wrapper with authentication and API methods
- [ ] Implement RA service for CSR submission, approval, and rejection
- [ ] Implement CA service for certificate issuance, revocation, and renewal
- [ ] Implement VA service for OCSP status and CRL retrieval
- [ ] Create RA, CA, and VA controllers with REST endpoints
- [ ] Set up routes, middleware, and main server entry point
- [ ] Create Dockerfile and integrate step-ca into docker-compose.yml
- [ ] Create README with step-ca setup and service configuration instructions