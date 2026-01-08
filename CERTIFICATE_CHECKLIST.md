# Certificate Checklist

This document lists all required certificates for all services (server & client) and their expected locations in both Docker and Development environments.

## Certificate Path Conventions

- **Development**: All paths are relative from the service root directory (e.g., `./certs/auth-ca.crt`)
- **Docker**: All paths are absolute container paths (e.g., `/app/certs/auth-ca.crt`)

## Server Services (gRPC mTLS Server Certificates)

Each service requires three files for its gRPC server:
- CA Certificate: `{service}-ca.crt`
- Server Certificate: `{service}-server.crt`
- Server Key: `{service}-server.key`

### 1. Auth Service
- **Development**: `services/auth/certs/`
  - `auth-ca.crt`
  - `auth-server.crt`
  - `auth-server.key`
- **Docker**: `/app/certs/` (mounted from `services/auth/certs/`)
  - `auth-ca.crt`
  - `auth-server.crt`
  - `auth-server.key`

### 2. Helpdesk Service
- **Development**: `services/helpdesk/certs/`
  - `helpdesk-ca.crt`
  - `helpdesk-server.crt`
  - `helpdesk-server.key`
- **Docker**: `/app/certs/` (mounted from `services/helpdesk/certs/`)
  - `helpdesk-ca.crt`
  - `helpdesk-server.crt`
  - `helpdesk-server.key`

### 3. Notification Service
- **Development**: `services/notification/certs/`
  - `notification-ca.crt`
  - `notification-server.crt`
  - `notification-server.key`
- **Docker**: `/app/certs/` (mounted from `services/notification/certs/`)
  - `notification-ca.crt`
  - `notification-server.crt`
  - `notification-server.key`

### 4. Certificate Service
- **Development**: `services/certificate/certs/`
  - `certificate-ca.crt`
  - `certificate-server.crt`
  - `certificate-server.key`
- **Docker**: `/app/certs/` (mounted from `services/certificate/certs/`)
  - `certificate-ca.crt`
  - `certificate-server.crt`
  - `certificate-server.key`

### 5. Container Management Service
- **Development**: `services/container-management/certs/`
  - `container-management-ca.crt`
  - `container-management-server.crt`
  - `container-management-server.key`
- **Docker**: `/app/certs/` (mounted from `services/container-management/certs/`)
  - `container-management-ca.crt`
  - `container-management-server.crt`
  - `container-management-server.key`

### 6. Client Container Service
- **Development**: `services/client-container/certs/`
  - `client-container-ca.crt`
  - `client-container-server.crt`
  - `client-container-server.key`
- **Docker**: `/app/certs/` (mounted from `services/client-container/certs/`)
  - `client-container-ca.crt`
  - `client-container-server.crt`
  - `client-container-server.key`

### 7. Inventory Service
- **Development**: `services/inventory/certs/`
  - `inventory-ca.crt`
  - `inventory-server.crt`
  - `inventory-server.key`
- **Docker**: `/app/certs/` (mounted from `services/inventory/certs/`)
  - `inventory-ca.crt`
  - `inventory-server.crt`
  - `inventory-server.key`

### 8. Geography Service
- **Development**: `services/geography/certs/`
  - `geography-ca.crt`
  - `geography-server.crt`
  - `geography-server.key`
- **Docker**: `/app/certs/` (mounted from `services/geography/certs/`)
  - `geography-ca.crt`
  - `geography-server.crt`
  - `geography-server.key`

### 9. Navigation Service
- **Development**: `services/navigation/certs/`
  - `navigation-ca.crt`
  - `navigation-server.crt`
  - `navigation-server.key`
- **Docker**: `/app/certs/` (mounted from `services/navigation/certs/`)
  - `navigation-ca.crt`
  - `navigation-server.crt`
  - `navigation-server.key`

## Client Services (gRPC mTLS Client Certificates)

### 1. Gateway Service
The gateway acts as a client to all backend services and uses the same client certificate for all services.

- **Development**: `services/gateway/certs/`
  - `gateway-client.crt`
  - `gateway-client.key`
  - `auth-ca.crt` (CA for Auth service)
  - `helpdesk-ca.crt` (CA for Helpdesk service)
  - `notification-ca.crt` (CA for Notification service)
  - `inventory-ca.crt` (CA for Inventory service)
  - `geography-ca.crt` (CA for Geography service)
  - `navigation-ca.crt` (CA for Navigation service)
- **Docker**: `/app/certs/` (mounted from `services/gateway/certs/`)
  - Same files as development

### 2. Helpdesk Service (Client to Auth)
- **Development**: `services/helpdesk/certs/`
  - `helpdesk-client.crt`
  - `helpdesk-client.key`
  - `auth-ca.crt` (CA for Auth service)
- **Docker**: `/app/certs/` (mounted from `services/helpdesk/certs/`)
  - Same files as development

### 3. Client Container Service (Client to Container Management & Notification)
- **Development**: `services/client-container/certs/`
  - `client-container-client.crt`
  - `client-container-client.key`
  - `container-management-ca.crt` (CA for Container Management service)
  - `notification-ca.crt` (CA for Notification service)
- **Docker**: `/app/certs/` (mounted from `services/client-container/certs/`)
  - Same files as development

### 4. Container Management Service (Client to Certificate Service)
- **Development**: `services/container-management/certs/`
  - `container-management-client.crt`
  - `container-management-client.key`
  - `certificate-ca.crt` (CA for Certificate service)
- **Docker**: `/app/certs/` (mounted from `services/container-management/certs/`)
  - Same files as development

## HTTP mTLS Certificates

### 1. Certificate Service (HTTP mTLS)
Uses the same certificates as gRPC:
- **Development**: `services/certificate/certs/`
  - `certificate-ca.crt`
  - `certificate-server.crt`
  - `certificate-server.key`
- **Docker**: `/app/certs/` (mounted from `services/certificate/certs/`)
  - Same files as development

### 2. Container Management Service (HTTP mTLS)
Uses the same certificates as gRPC:
- **Development**: `services/container-management/certs/`
  - `container-management-ca.crt`
  - `container-management-server.crt`
  - `container-management-server.key`
- **Docker**: `/app/certs/` (mounted from `services/container-management/certs/`)
  - Same files as development

## Client Container Specific Certificates

### Container Certificates
These are generated per container instance:
- **Development**: `services/client-container/certs/`
  - `container.crt` (per-container certificate)
  - `container.key` (per-container key)
  - `agent-ca.crt` (CA for validating agent certificates)
- **Docker**: `/app/certs/` (mounted from `services/client-container/certs/`)
  - Same files as development

## Agent Certificates

The agent uses platform-specific paths but should be configured via environment variables.

### Agent Client Certificates
- **Development**: Platform-specific (configure via `AGENT_CLIENT_CERT` and `AGENT_CLIENT_KEY` env vars)
  - Recommended: `./certs/agent-client.crt`
  - Recommended: `./certs/agent-client.key`
  - CA: `./certs/client-container-ca.crt` (configured via `AGENT_BACKEND_CA_CERT`)
- **Docker**: Same as development (certificates should be mounted from host)

## Certificate Summary by Service

### Auth Service
- Server: `auth-ca.crt`, `auth-server.crt`, `auth-server.key`
- Location: `services/auth/certs/`

### Helpdesk Service
- Server: `helpdesk-ca.crt`, `helpdesk-server.crt`, `helpdesk-server.key`
- Client: `helpdesk-client.crt`, `helpdesk-client.key`, `auth-ca.crt`
- Location: `services/helpdesk/certs/`

### Notification Service
- Server: `notification-ca.crt`, `notification-server.crt`, `notification-server.key`
- Location: `services/notification/certs/`

### Certificate Service
- Server (HTTP & gRPC): `certificate-ca.crt`, `certificate-server.crt`, `certificate-server.key`
- Location: `services/certificate/certs/`

### Container Management Service
- Server (HTTP & gRPC): `container-management-ca.crt`, `container-management-server.crt`, `container-management-server.key`
- Client: `container-management-client.crt`, `container-management-client.key`, `certificate-ca.crt`
- Location: `services/container-management/certs/`

### Client Container Service
- Server: `client-container-ca.crt`, `client-container-server.crt`, `client-container-server.key`
- Client: `client-container-client.crt`, `client-container-client.key`, `container-management-ca.crt`, `notification-ca.crt`
- Container: `container.crt`, `container.key`, `agent-ca.crt`
- Location: `services/client-container/certs/`

### Inventory Service
- Server: `inventory-ca.crt`, `inventory-server.crt`, `inventory-server.key`
- Location: `services/inventory/certs/`

### Geography Service
- Server: `geography-ca.crt`, `geography-server.crt`, `geography-server.key`
- Location: `services/geography/certs/`

### Navigation Service
- Server: `navigation-ca.crt`, `navigation-server.crt`, `navigation-server.key`
- Location: `services/navigation/certs/`

### Gateway Service
- Client: `gateway-client.crt`, `gateway-client.key`
- CA certs: `auth-ca.crt`, `helpdesk-ca.crt`, `notification-ca.crt`, `inventory-ca.crt`, `geography-ca.crt`, `navigation-ca.crt`
- Location: `services/gateway/certs/`

## Verification Checklist

Use this checklist to verify all certificates are in place:

### Server Certificates
- [ ] `services/auth/certs/auth-ca.crt`
- [ ] `services/auth/certs/auth-server.crt`
- [ ] `services/auth/certs/auth-server.key`
- [ ] `services/helpdesk/certs/helpdesk-ca.crt`
- [ ] `services/helpdesk/certs/helpdesk-server.crt`
- [ ] `services/helpdesk/certs/helpdesk-server.key`
- [ ] `services/notification/certs/notification-ca.crt`
- [ ] `services/notification/certs/notification-server.crt`
- [ ] `services/notification/certs/notification-server.key`
- [ ] `services/certificate/certs/certificate-ca.crt`
- [ ] `services/certificate/certs/certificate-server.crt`
- [ ] `services/certificate/certs/certificate-server.key`
- [ ] `services/container-management/certs/container-management-ca.crt`
- [ ] `services/container-management/certs/container-management-server.crt`
- [ ] `services/container-management/certs/container-management-server.key`
- [ ] `services/client-container/certs/client-container-ca.crt`
- [ ] `services/client-container/certs/client-container-server.crt`
- [ ] `services/client-container/certs/client-container-server.key`
- [ ] `services/inventory/certs/inventory-ca.crt`
- [ ] `services/inventory/certs/inventory-server.crt`
- [ ] `services/inventory/certs/inventory-server.key`
- [ ] `services/geography/certs/geography-ca.crt`
- [ ] `services/geography/certs/geography-server.crt`
- [ ] `services/geography/certs/geography-server.key`
- [ ] `services/navigation/certs/navigation-ca.crt`
- [ ] `services/navigation/certs/navigation-server.crt`
- [ ] `services/navigation/certs/navigation-server.key`

### Client Certificates
- [ ] `services/gateway/certs/gateway-client.crt`
- [ ] `services/gateway/certs/gateway-client.key`
- [ ] `services/gateway/certs/auth-ca.crt`
- [ ] `services/gateway/certs/helpdesk-ca.crt`
- [ ] `services/gateway/certs/notification-ca.crt`
- [ ] `services/gateway/certs/inventory-ca.crt`
- [ ] `services/gateway/certs/geography-ca.crt`
- [ ] `services/gateway/certs/navigation-ca.crt`
- [ ] `services/helpdesk/certs/helpdesk-client.crt`
- [ ] `services/helpdesk/certs/helpdesk-client.key`
- [ ] `services/helpdesk/certs/auth-ca.crt`
- [ ] `services/client-container/certs/client-container-client.crt`
- [ ] `services/client-container/certs/client-container-client.key`
- [ ] `services/client-container/certs/container-management-ca.crt`
- [ ] `services/client-container/certs/notification-ca.crt`
- [ ] `services/container-management/certs/container-management-client.crt`
- [ ] `services/container-management/certs/container-management-client.key`
- [ ] `services/container-management/certs/certificate-ca.crt`

### Container-Specific Certificates
- [ ] `services/client-container/certs/container.crt` (generated per container)
- [ ] `services/client-container/certs/container.key` (generated per container)
- [ ] `services/client-container/certs/agent-ca.crt`

### Agent Certificates
- [ ] Agent client certificate (configured via `AGENT_CLIENT_CERT`)
- [ ] Agent client key (configured via `AGENT_CLIENT_KEY`)
- [ ] Backend CA certificate (configured via `AGENT_BACKEND_CA_CERT`)

## Notes

1. **All paths are relative** in development mode (e.g., `./certs/...`)
2. **All paths are absolute** in Docker mode (e.g., `/app/certs/...`)
3. **Certificate directories must exist** before services start
4. **Empty certificate directories** are acceptable - services will log warnings if certificates are missing
5. **Certificate generation** should be done using the Certificate Service or Step-CA
6. **Client certificates** can be shared across services if they're signed by the same CA (e.g., gateway-client.crt is used for all services)
7. **Container certificates** (`container.crt`, `container.key`) are generated dynamically per container instance

## Environment File Usage

- **Docker**: Use `.env.docker` - all paths use `/app/certs/` and service names
- **Development**: Use `.env.development` - all paths use `./certs/` and `localhost`

To use these files:
- **Docker**: `docker-compose` automatically uses environment variables from `.env.docker` if present
- **Development**: Load `.env.development` before running services (e.g., `source .env.development` or use `godotenv`)

