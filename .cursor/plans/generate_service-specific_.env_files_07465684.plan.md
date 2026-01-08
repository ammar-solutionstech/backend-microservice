---
name: Generate Service-Specific .env Files
overview: Create a plan to generate individual .env files for each microservice containing only that service's configuration, including all TLS/mTLS settings. Start with container-management service as the first implementation.
todos:
  - id: docker-client
    content: Add Docker SDK dependency and create Docker client service with container operations (create, start, stop, restart, remove, status, update, list)
    status: completed
  - id: container-orchestrator
    content: Create container orchestrator service for generating Docker container configurations and creating client-container instances
    status: completed
    dependencies:
      - docker-client
  - id: service-manager
    content: Create service manager for managing services/plugins within client-container instances (add, remove, enable, disable, list, update)
    status: completed
    dependencies:
      - docker-client
  - id: add-model-comments
    content: Add detailed comments to all existing model files (organization.go, client_container.go) - clarify each field purpose, format, and list static value options where applicable
    status: completed
  - id: update-models
    content: Update client_container model with Docker fields (docker_container_id, docker_container_name, docker_status, services, container_config) and create container_services model - both with detailed field comments
    status: completed
    dependencies:
      - add-model-comments
  - id: update-migration
    content: Update migration to add Docker fields to client_containers table and create container_services table
    status: completed
    dependencies:
      - update-models
  - id: update-registration
    content: Update RegisterClientContainer to create Docker container via orchestrator, configure certificates/data, and start container
    status: completed
    dependencies:
      - container-orchestrator
      - update-models
  - id: lifecycle-endpoints
    content: Add lifecycle management endpoints (start, stop, restart, update, remove, status) to client_container_controller
    status: completed
    dependencies:
      - container-orchestrator
  - id: service-endpoints
    content: Create service_controller with endpoints for managing services in containers (add, list, get, update, enable, disable, remove)
    status: completed
    dependencies:
      - service-manager
  - id: client-container-dockerfile
    content: Create Dockerfile for client-container service with multi-stage build
    status: completed
  - id: service-registry
    content: Implement service registry and service interface in client-container for plugin management
    status: completed
  - id: internal-service-api
    content: Add internal service management API endpoints in client-container (requires mTLS) for container-management to manage services
    status: completed
    dependencies:
      - service-registry
  - id: update-config
    content: Update container-management config with Docker settings (DockerHost, ClientContainerImage, DockerNetwork)
    status: completed
  - id: update-docker-compose
    content: Update docker-compose.yml to mount Docker socket to container-management and add client-container build configuration
    status: completed
    dependencies:
      - client-container-dockerfile
  - id: update-main-server
    content: Update container-management main server to initialize Docker client, orchestrator, and service manager, and register new routes
    status: completed
    dependencies:
      - docker-client
      - container-orchestrator
      - service-manager
      - lifecycle-endpoints
      - service-endpoints
---

# Generate Service-Specific .env Files

## Overview

Create individual `.env` files for each microservice in their respective service directories. Each `.env` file will contain only that service's configuration, making it easier to manage, deploy, and maintain service-specific settings. Start with `container-management` service.

## Goals

1. **Service Isolation**: Each service has its own `.env` file with only its configuration
2. **Complete TLS/mTLS Configuration**: All certificate paths, keys, and TLS settings included
3. **Documentation**: Clear comments explaining each variable
4. **Environment-Specific**: Support for development, staging, and production
5. **Security**: Sensitive values clearly marked, with placeholders for production

## Implementation Plan

### Phase 1: Container-Management Service .env File

**File Location:** `services/container-management/.env.example`Create a comprehensive `.env.example` file with all required variables organized by category:

#### 1. Database Configuration

- `CONTAINER_MGMT_DB_HOST` - Database host
- `CONTAINER_MGMT_DB_PORT` - Database port
- `CONTAINER_MGMT_DB_NAME` - Database name
- `CONTAINER_MGMT_DB_USER` - Database user
- `CONTAINER_MGMT_DB_PASSWORD` - Database password

#### 2. Server Configuration

- `CONTAINER_MGMT_PORT` - HTTP server port

#### 3. Certificate Service Connection (mTLS Client)

- `CERTIFICATE_SERVICE_URL` - Certificate service HTTPS URL
- `CERTIFICATE_SERVICE_CA` - Path to CA cert for verifying certificate service
- `CERTIFICATE_SERVICE_CERT` - Path to client cert for mTLS
- `CERTIFICATE_SERVICE_KEY` - Path to client private key
- `CERTIFICATE_SERVICE_KEY_PASSWORD` - Optional password for encrypted key

#### 4. mTLS Server Configuration

- `MTLS_CA_CERT` - Path to CA cert for validating client certificates
- `MTLS_SERVER_CERT` - Path to server certificate
- `MTLS_SERVER_KEY` - Path to server private key
- `MTLS_SERVER_KEY_PASSWORD` - Optional password for encrypted server key

#### 5. Bootstrap Token

- `BOOTSTRAP_TOKEN_SECRET` - Secret key for generating/validating bootstrap tokens

#### 6. CSR Validation Rules

- `CSR_REQUIRED_ORG` - Required organization name in CSR (optional)
- `CSR_REQUIRED_COUNTRY` - Required country code in CSR (optional)

#### 7. Docker Configuration

- `DOCKER_HOST` - Docker daemon socket path
- `DOCKER_NETWORK` - Docker network name for client containers
- `CERTIFICATES_PATH` - Host path where certificates are stored

#### 8. Client Container Configuration

- `CLIENT_CONTAINER_IMAGE` - Docker image name for client containers
- `CLIENT_CONTAINER_IMAGE_TAG` - Docker image tag
- `CLIENT_CONTAINER_PORT` - HTTP port for client containers
- `CLIENT_CONTAINER_GRPC_PORT` - gRPC port for client containers
- `CLIENT_CONTAINER_DB_HOST` - Database host for client containers
- `CLIENT_CONTAINER_DB_PORT` - Database port for client containers
- `CLIENT_CONTAINER_DB_NAME` - Database name for client containers
- `CLIENT_CONTAINER_DB_USER` - Database user for client containers
- `CLIENT_CONTAINER_DB_PASSWORD` - Database password for client containers

#### 9. External Service URLs

- `CONTAINER_MGMT_SERVICE_URL` - This service's URL (for client containers to connect)
- `NOTIFICATION_SERVICE_GRPC` - Notification service gRPC address

### File Structure

```env
# =============================================================================
# Container Management Service Configuration
# =============================================================================
# This file contains all configuration for the container-management service.
# Copy this file to .env and update values for your environment.
# =============================================================================

# -----------------------------------------------------------------------------
# Database Configuration
# -----------------------------------------------------------------------------
CONTAINER_MGMT_DB_HOST=localhost
CONTAINER_MGMT_DB_PORT=5432
CONTAINER_MGMT_DB_NAME=container_mgmt_db
CONTAINER_MGMT_DB_USER=postgres
CONTAINER_MGMT_DB_PASSWORD=postgres

# -----------------------------------------------------------------------------
# Server Configuration
# -----------------------------------------------------------------------------
# Port on which the HTTP server listens
CONTAINER_MGMT_PORT=8005

# -----------------------------------------------------------------------------
# Certificate Service Connection (mTLS Client)
# -----------------------------------------------------------------------------
# URL of the certificate service (HTTPS required for mTLS)
CERTIFICATE_SERVICE_URL=https://certificate-service:8004

# Path to CA certificate for verifying certificate service's server certificate
# This should be the CA that signed the certificate service's server cert
CERTIFICATE_SERVICE_CA=/certs/ca.crt

# Path to client certificate for mTLS authentication with certificate service
CERTIFICATE_SERVICE_CERT=/certs/client.crt

# Path to client private key for mTLS authentication
CERTIFICATE_SERVICE_KEY=/certs/client.key

# Optional: Password for encrypted private key (leave empty if key is not encrypted)
CERTIFICATE_SERVICE_KEY_PASSWORD=

# -----------------------------------------------------------------------------
# mTLS Server Configuration
# -----------------------------------------------------------------------------
# Path to CA certificate for validating client certificates
# This CA should have signed all client certificates that connect to this service
MTLS_CA_CERT=/certs/ca.crt

# Path to server certificate for mTLS
# This certificate should be signed by the CA specified in MTLS_CA_CERT
MTLS_SERVER_CERT=/certs/server.crt

# Path to server private key
MTLS_SERVER_KEY=/certs/server.key

# Optional: Password for encrypted server private key (leave empty if key is not encrypted)
MTLS_SERVER_KEY_PASSWORD=

# -----------------------------------------------------------------------------
# Bootstrap Token Configuration
# -----------------------------------------------------------------------------
# Secret key for generating and validating bootstrap tokens
# IMPORTANT: Change this in production! Use a strong, random secret.
# Generate with: openssl rand -base64 32
BOOTSTRAP_TOKEN_SECRET=change-me-in-production-use-strong-random-secret

# -----------------------------------------------------------------------------
# CSR Validation Rules (Optional)
# -----------------------------------------------------------------------------
# If set, all CSRs must contain this organization name
# Leave empty to allow any organization
CSR_REQUIRED_ORG=

# If set, all CSRs must contain this country code (e.g., "US", "GB")
# Leave empty to allow any country
CSR_REQUIRED_COUNTRY=

# -----------------------------------------------------------------------------
# Docker Configuration
# -----------------------------------------------------------------------------
# Docker daemon socket path
# For Linux: unix:///var/run/docker.sock
# For Windows: npipe:////./pipe/docker_engine
# For Docker Desktop: unix:///var/run/docker.sock (when mounted)
DOCKER_HOST=unix:///var/run/docker.sock

# Docker network name for client containers
# This network should already exist or be created by docker-compose
DOCKER_NETWORK=microservices-network

# Host path where certificates are stored
# This path will be mounted into client containers
CERTIFICATES_PATH=/certs

# -----------------------------------------------------------------------------
# Client Container Configuration
# -----------------------------------------------------------------------------
# Docker image name for client containers
CLIENT_CONTAINER_IMAGE=client-container

# Docker image tag
CLIENT_CONTAINER_IMAGE_TAG=latest

# HTTP port for client containers
CLIENT_CONTAINER_PORT=8006

# gRPC port for client containers
CLIENT_CONTAINER_GRPC_PORT=9006

# Database configuration for client containers
# These values are passed as environment variables to client containers
CLIENT_CONTAINER_DB_HOST=postgres-itaas
CLIENT_CONTAINER_DB_PORT=5432
CLIENT_CONTAINER_DB_NAME=client_container_db
CLIENT_CONTAINER_DB_USER=postgres
CLIENT_CONTAINER_DB_PASSWORD=postgres

# -----------------------------------------------------------------------------
# External Service URLs
# -----------------------------------------------------------------------------
# This service's URL (used by client containers to connect back)
# Use HTTPS for production
CONTAINER_MGMT_SERVICE_URL=https://container-management-service:8005

# Notification service gRPC address (for sending verification codes)
NOTIFICATION_SERVICE_GRPC=notification-service:9003
```



### Phase 2: Update Configuration Loading

**File:** `services/container-management/internal/config/config.go`Ensure the config loader looks for `.env` in the service directory first, then falls back to root:

```go
// Try service-specific .env first, then root .env
_ = godotenv.Load(".env")
if _, err := os.Stat("../.env"); err == nil {
    _ = godotenv.Overload("../.env") // Root .env as fallback
}
```



### Phase 3: Create .env.example Template

Create `services/container-management/.env.example` with the structure above, including:

- Clear section headers
- Comments explaining each variable
- Default values for development
- Placeholders for production secrets
- Notes about certificate generation

### Phase 4: Documentation

Create `services/container-management/ENV_CONFIG.md` documenting:

- How to generate certificates
- Certificate file locations
- Environment-specific configurations
- Security best practices

### Phase 5: Replicate for Other Services

After container-management is complete, create similar `.env.example` files for:

1. `services/certificate/.env.example`
2. `services/client-container/.env.example`
3. `services/auth/.env.example`
4. `services/helpdesk/.env.example`
5. `services/notification/.env.example`
6. `services/gateway/.env.example`

## Certificate Path Notes

For Docker deployments:

- Certificates should be mounted from host to `/certs` in container
- Paths in `.env` should use container paths (e.g., `/certs/ca.crt`)
- Host paths are specified in `docker-compose.yml` volumes

For local development:

- Certificates can be in `./certs` relative to project root
- Update paths accordingly (e.g., `./certs/ca.crt`)

## Security Considerations

1. **Never commit `.env` files** - Add to `.gitignore`
2. **Use `.env.example`** - Template files can be committed
3. **Rotate secrets** - Bootstrap token secret should be rotated regularly
4. **Protect keys** - Private keys should have restricted permissions (600)
5. **Use strong secrets** - Bootstrap token secret should be at least 32 bytes

## Next Steps