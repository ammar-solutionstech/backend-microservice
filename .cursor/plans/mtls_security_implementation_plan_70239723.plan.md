---
name: mTLS Security Implementation Plan
overview: Comprehensive security assessment and implementation plan to enable mTLS for all inter-service communication across all microservices, including certificate naming conventions, current status analysis, and step-by-step improvements.
todos: []
---

# mTLS

Security Implementation Plan

## Executive Summary

This plan provides a comprehensive security assessment of the current microservices architecture and outlines the implementation of mutual TLS (mTLS) for all inter-service communication. Currently, most services use insecure connections (plain HTTP/gRPC), with only partial mTLS implementation in container-management and certificate services.

## Current Security Status

### Service-by-Service Analysis

#### 1. **Auth Service** (`auth-service`)

- **Ports**: HTTP 8001, gRPC 9001
- **Current TLS/mTLS**: ❌ None
- **Server**: Plain HTTP, plain gRPC
- **Client Connections**: None (receives requests only)
- **Certificate Name**: Not configured
- **Common Name (CN)**: `auth-service`
- **Issues**: 
- No TLS on HTTP server
- No mTLS on gRPC server
- No certificate configuration

#### 2. **Gateway Service** (`gateway`)

- **Ports**: HTTP 8080
- **Current TLS/mTLS**: ❌ None
- **Server**: Plain HTTP
- **Client Connections**: 
- gRPC to Auth (insecure)
- gRPC to Helpdesk (insecure)
- gRPC to Notification (insecure)
- **Certificate Name**: Not configured
- **Common Name (CN)**: `gateway`
- **Issues**:
- No TLS on HTTP server
- All gRPC clients use `insecure.NewCredentials()`
- No mTLS configuration

#### 3. **Helpdesk Service** (`helpdesk-service`)

- **Ports**: HTTP 8002, gRPC 9002
- **Current TLS/mTLS**: ❌ None
- **Server**: Plain HTTP, plain gRPC
- **Client Connections**: 
- gRPC to Auth (insecure)
- **Certificate Name**: Not configured
- **Common Name (CN)**: `helpdesk-service`
- **Issues**:
- No TLS on HTTP/gRPC servers
- gRPC client uses `insecure.NewCredentials()`
- No certificate configuration

#### 4. **Notification Service** (`notification-service`)

- **Ports**: HTTP 8003, gRPC 9003
- **Current TLS/mTLS**: ❌ None
- **Server**: Plain HTTP, plain gRPC
- **Client Connections**: None (receives requests only)
- **Certificate Name**: Not configured
- **Common Name (CN)**: `notification-service`
- **Issues**:
- No TLS on HTTP/gRPC servers
- No certificate configuration

#### 5. **Certificate Service** (`certificate-service`)

- **Ports**: HTTP 8004, gRPC 9004
- **Current TLS/mTLS**: ⚠️ Partial
- **Server**: 
- HTTP: Optional mTLS (configured but may fallback to HTTP)
- gRPC: Plain (no TLS)
- **Client Connections**: 
- HTTP to step-ca (JWT auth, optional mTLS)
- **Certificate Name**: 
- Server: `CERTIFICATE_SERVER_CERT`, `CERTIFICATE_SERVER_KEY`
- CA: `MTLS_CA_CERT`
- **Common Name (CN)**: `certificate-service`
- **Issues**:
- gRPC server has no TLS
- mTLS config uses `RequireAnyClientCert` (should be `RequireAndVerifyClientCert`)
- `InsecureSkipVerify: true` in some configs
- No certificate validation for step-ca connection

#### 6. **Container Management Service** (`container-management-service`)

- **Ports**: HTTP 8005
- **Current TLS/mTLS**: ⚠️ Partial
- **Server**: 
- HTTP: Optional mTLS (configured)
- **Client Connections**: 
- HTTP to Certificate Service (mTLS configured)
- HTTP to Client Containers (TODO - no mTLS)
- **Certificate Name**: 
- Server: `MTLS_SERVER_CERT`, `MTLS_SERVER_KEY`
- Client (to cert service): `CERTIFICATE_SERVICE_CERT`, `CERTIFICATE_SERVICE_KEY`
- CA: `MTLS_CA_CERT`, `CERTIFICATE_SERVICE_CA`
- **Common Name (CN)**: 
- Server: `container-management-service`
- Client: `container-management-client`
- **Issues**:
- Uses `RequireAnyClientCert` instead of `RequireAndVerifyClientCert`
- `InsecureSkipVerify: true` in client config
- Service manager calls to client-containers use plain HTTP (TODO comment)

#### 7. **Client Container Service** (`client-container`)

- **Ports**: HTTP 8006, gRPC 9006
- **Current TLS/mTLS**: ⚠️ Partial
- **Server**: 
- HTTP: Partial mTLS (certificate loaded but incomplete)
- gRPC: Partial TLS (certificate loaded but incomplete)
- **Client Connections**: 
- HTTP to Container Management (mTLS configured)
- gRPC to Notification (insecure)
- **Certificate Name**: 
- Server: `CONTAINER_CERT_PATH`, `CONTAINER_KEY_PATH`
- Client: `CONTAINER_MGMT_SERVICE_CERT`, `CONTAINER_MGMT_SERVICE_KEY`
- Agent CA: `AGENT_CA_CERT_PATH`
- **Common Name (CN)**: 
- Server: `client-container-{container-id}` (dynamic)
- Client: `client-container-client`
- **Issues**:
- gRPC client to notification uses `insecure.NewCredentials()`
- HTTP server TLS config incomplete
- gRPC server TLS config incomplete

## Certificate Naming Convention

### Proposed Standard

All certificates should follow this naming pattern:

- **CA Certificate**: `{service-name}-ca.crt` / `{service-name}-ca.key`
- **Server Certificate**: `{service-name}-server.crt` / `{service-name}-server.key`
- **Client Certificate**: `{service-name}-client.crt` / `{service-name}-client.key`

### Service-Specific Certificate Names

| Service | CA Cert | Server Cert | Client Cert | CN (Server) | CN (Client) ||---------|---------|-------------|-------------|-------------|-------------|| Auth | `auth-ca.crt` | `auth-server.crt` | `auth-client.crt` | `auth-service` | `auth-client` || Gateway | `gateway-ca.crt` | `gateway-server.crt` | `gateway-client.crt` | `gateway` | `gateway-client` || Helpdesk | `helpdesk-ca.crt` | `helpdesk-server.crt` | `helpdesk-client.crt` | `helpdesk-service` | `helpdesk-client` || Notification | `notification-ca.crt` | `notification-server.crt` | `notification-client.crt` | `notification-service` | `notification-client` || Certificate | `certificate-ca.crt` | `certificate-server.crt` | `certificate-client.crt` | `certificate-service` | `certificate-client` || Container Management | `container-mgmt-ca.crt` | `container-mgmt-server.crt` | `container-mgmt-client.crt` | `container-management-service` | `container-management-client` || Client Container | `client-container-ca.crt` | `client-container-{id}-server.crt` | `client-container-client.crt` | `client-container-{container-id}` | `client-container-client` |

## Implementation Plan

### Phase 1: Certificate Authority Setup

1. Create root CA for microservices
2. Generate intermediate CAs for each service (optional, for better isolation)
3. Establish certificate lifecycle management

### Phase 2: Service Server mTLS (Incoming)

1. **Auth Service**: Add mTLS to HTTP and gRPC servers
2. **Gateway Service**: Add mTLS to HTTP server
3. **Helpdesk Service**: Add mTLS to HTTP and gRPC servers
4. **Notification Service**: Add mTLS to HTTP and gRPC servers
5. **Certificate Service**: Fix gRPC server mTLS, enforce strict validation
6. **Container Management**: Fix mTLS validation (RequireAndVerifyClientCert)
7. **Client Container**: Complete HTTP and gRPC server mTLS

### Phase 3: Service Client mTLS (Outgoing)

1. **Gateway**: Update all gRPC clients to use mTLS
2. **Helpdesk**: Update Auth gRPC client to use mTLS
3. **Client Container**: Update Notification gRPC client to use mTLS
4. **Container Management**: Complete client-container HTTP client mTLS

### Phase 4: Configuration and Environment

1. Update all service configs to load certificates
2. Add certificate paths to docker-compose.yml
3. Generate .env files with certificate paths
4. Update documentation

### Phase 5: Validation and Testing

1. Test all inter-service communication with mTLS
2. Verify certificate validation
3. Test certificate revocation
4. Performance testing

## Detailed Implementation Steps

### Step 1: Create Root CA and Service Certificates

Generate certificates for all services using a common CA or service-specific CAs.

### Step 2: Update Auth Service

**Files to modify**:

- `services/auth/internal/config/config.go`: Add mTLS config fields
- `services/auth/cmd/server/main.go`: 
- Add TLS config to HTTP server
- Add TLS credentials to gRPC server
- Use `tls.RequireAndVerifyClientCert`

**Certificate files**:

- `/certs/auth-ca.crt`
- `/certs/auth-server.crt`, `/certs/auth-server.key`
- `/certs/auth-client.crt`, `/certs/auth-client.key` (for clients connecting to auth)

### Step 3: Update Gateway Service

**Files to modify**:

- `services/gateway/internal/config/config.go`: Add mTLS config
- `services/gateway/cmd/server/main.go`: Add TLS to HTTP server
- `services/gateway/internal/clients/auth_client.go`: Use TLS credentials
- `services/gateway/internal/clients/helpdesk_client.go`: Use TLS credentials
- `services/gateway/internal/clients/notification_client.go`: Use TLS credentials

**Certificate files**:

- `/certs/gateway-ca.crt`
- `/certs/gateway-server.crt`, `/certs/gateway-server.key`
- `/certs/gateway-client.crt`, `/certs/gateway-client.key`

### Step 4: Update Helpdesk Service

**Files to modify**:

- `services/helpdesk/internal/config/config.go`: Add mTLS config
- `services/helpdesk/cmd/server/main.go`: Add TLS to HTTP and gRPC servers
- `services/helpdesk/internal/services/auth_client.go`: Use TLS credentials

**Certificate files**:

- `/certs/helpdesk-ca.crt`
- `/certs/helpdesk-server.crt`, `/certs/helpdesk-server.key`
- `/certs/helpdesk-client.crt`, `/certs/helpdesk-client.key`

### Step 5: Update Notification Service

**Files to modify**:

- `services/notification/internal/config/config.go`: Add mTLS config
- `services/notification/cmd/server/main.go`: Add TLS to HTTP and gRPC servers

**Certificate files**:

- `/certs/notification-ca.crt`
- `/certs/notification-server.crt`, `/certs/notification-server.key`
- `/certs/notification-client.crt`, `/certs/notification-client.key`

### Step 6: Fix Certificate Service

**Files to modify**:

- `services/certificate/internal/config/config.go`: 
- Fix `initServerTLS` to use `RequireAndVerifyClientCert`
- Remove `InsecureSkipVerify`
- `services/certificate/cmd/server/main.go`: 
- Add TLS to gRPC server
- Enforce mTLS (no HTTP fallback)

### Step 7: Fix Container Management Service

**Files to modify**:

- `services/container-management/internal/config/config.go`: 
- Fix `initServerTLS` to use `RequireAndVerifyClientCert`
- Fix `initCertificateServiceTLS` to remove `InsecureSkipVerify`
- `services/container-management/internal/services/service_manager.go`: 
- Add mTLS client config to `callContainerServiceAPI`

### Step 8: Complete Client Container Service

**Files to modify**:

- `services/client-container/internal/config/config.go`: 
- Complete TLS config initialization
- `services/client-container/cmd/server/main.go`: 
- Complete HTTP server TLS config
- Complete gRPC server TLS config with client verification
- `services/client-container/internal/services/notification_client.go`: 
- Use TLS credentials instead of insecure

### Step 9: Update Docker Compose

**File**: `docker-compose.yml`Add certificate volume mounts to all services:

```yaml
volumes:
    - ./certs:/certs:ro
```



### Step 10: Update Environment Files

Generate/update `.env` files for each service with certificate paths following the naming convention.

## Security Improvements Summary

### Current State

- ❌ 0/7 services with complete mTLS
- ⚠️ 3/7 services with partial mTLS
- ❌ All gRPC clients use insecure connections
- ⚠️ Weak certificate validation (`RequireAnyClientCert`, `InsecureSkipVerify`)

### Target State

- ✅ 7/7 services with complete mTLS
- ✅ All gRPC clients use TLS with client certificates
- ✅ Strict certificate validation (`RequireAndVerifyClientCert`)
- ✅ No insecure fallbacks
- ✅ Standardized certificate naming

## Risk Assessment

### Current Risks

1. **Man-in-the-Middle Attacks**: All plain HTTP/gRPC connections vulnerable