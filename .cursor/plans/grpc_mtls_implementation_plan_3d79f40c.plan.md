---
name: gRPC mTLS Implementation Plan
overview: Comprehensive plan to implement mutual TLS (mTLS) for all gRPC communication across all microservices, including server and client configurations, certificate management, and updates to internal and external services.
todos:
  - id: grpc-config-auth
    content: Add gRPC mTLS configuration to Auth service config (server certs, CA cert)
    status: completed
  - id: grpc-server-auth
    content: Update Auth service gRPC server to use mTLS credentials with client verification
    status: completed
    dependencies:
      - grpc-config-auth
  - id: grpc-config-gateway
    content: Add gRPC mTLS client configuration to Gateway service config (client certs for Auth, Helpdesk, Notification)
    status: completed
  - id: grpc-clients-gateway
    content: Update Gateway gRPC clients (Auth, Helpdesk, Notification) to use mTLS credentials
    status: completed
    dependencies:
      - grpc-config-gateway
  - id: grpc-config-helpdesk
    content: Add gRPC mTLS configuration to Helpdesk service config (server certs, client certs for Auth)
    status: completed
  - id: grpc-server-helpdesk
    content: Update Helpdesk service gRPC server to use mTLS credentials
    status: completed
    dependencies:
      - grpc-config-helpdesk
  - id: grpc-client-helpdesk
    content: Update Helpdesk Auth gRPC client to use mTLS credentials
    status: completed
    dependencies:
      - grpc-config-helpdesk
  - id: grpc-config-notification
    content: Add gRPC mTLS server configuration to Notification service config
    status: completed
  - id: grpc-server-notification
    content: Update Notification service gRPC server to use mTLS credentials with client verification
    status: completed
    dependencies:
      - grpc-config-notification
  - id: grpc-config-certificate
    content: Add gRPC mTLS server configuration to Certificate service config
    status: completed
  - id: grpc-server-certificate
    content: Add gRPC server to Certificate service with mTLS credentials
    status: completed
    dependencies:
      - grpc-config-certificate
  - id: grpc-config-client-container
    content: Complete gRPC mTLS configuration in Client-Container service config (server and client certs)
    status: completed
  - id: grpc-server-client-container
    content: Complete Client-Container gRPC server mTLS configuration with client verification
    status: completed
    dependencies:
      - grpc-config-client-container
  - id: grpc-client-client-container
    content: Update Client-Container Notification gRPC client to use mTLS credentials
    status: completed
    dependencies:
      - grpc-config-client-container
  - id: grpc-config-container-management
    content: Add gRPC mTLS configuration to Container-Management service config (server certs, optional client certs for Certificate service)
    status: completed
  - id: grpc-server-container-management
    content: Add gRPC server to Container-Management service with mTLS credentials
    status: completed
    dependencies:
      - grpc-config-container-management
  - id: grpc-client-container-management
    content: Optionally add gRPC client to Container-Management for Certificate service communication (currently uses HTTP)
    status: completed
    dependencies:
      - grpc-config-container-management
  - id: grpc-tls-utils
    content: Create reusable gRPC TLS utility functions for loading server and client credentials
    status: completed
  - id: grpc-env-files
    content: Update all service .env files with gRPC mTLS certificate paths
    status: completed
  - id: grpc-docker-compose
    content: Verify Docker Compose has certificate volumes mounted for all services
    status: completed
  - id: grpc-certificates
    content: Generate gRPC mTLS certificates for all services (or document reuse of HTTP certificates)
    status: completed
  - id: grpc-testing
    content: Test all gRPC mTLS connections between services and verify certificate validation
    status: completed
    dependencies:
      - grpc-server-auth
      - grpc-clients-gateway
      - grpc-server-helpdesk
      - grpc-client-helpdesk
      - grpc-server-notification
      - grpc-server-certificate
      - grpc-server-client-container
      - grpc-client-client-container
      - grpc-server-container-management
      - grpc-client-container-management
---

# gRPC

mTLS Implementation Plan

## Overview

Implement mutual TLS (mTLS) for all gRPC communication across all microservices. Internal services (certificate, client-container, notification) will use gRPC with mTLS only, while other services (auth, helpdesk, gateway) will support both REST and gRPC with mTLS.

## Current State Analysis

### gRPC Servers (All Currently Insecure)

- **Auth Service**: gRPC server on port 9001 (no TLS)
- **Helpdesk Service**: gRPC server on port 9002 (no TLS)
- **Notification Service**: gRPC server on port 9003 (no TLS)
- **Certificate Service**: gRPC server on port 9004 (no TLS) - internal
- **Container-Management Service**: No gRPC server (needs to be added) - port 9005
- **Client-Container Service**: gRPC server on port 9006 (partial TLS, incomplete) - internal

### gRPC Clients (All Currently Insecure)

- **Gateway → Auth**: Uses `insecure.NewCredentials()`
- **Gateway → Helpdesk**: Uses `insecure.NewCredentials()`
- **Gateway → Notification**: Uses `insecure.NewCredentials()`
- **Helpdesk → Auth**: Uses `insecure.NewCredentials()`
- **Client-Container → Notification**: Uses `insecure.NewCredentials()`
- **Container-Management → Certificate**: Currently uses HTTP mTLS (optional gRPC client)

## Certificate Naming Convention

All gRPC certificates follow the same naming as HTTP certificates:| Service | gRPC Server Cert | gRPC Client Cert | CN (Server) | CN (Client) ||---------|------------------|------------------|-------------|-------------|| Auth | `auth-server.crt/key` | `auth-client.crt/key` | `auth-service` | `auth-client` || Gateway | `gateway-server.crt/key` | `gateway-client.crt/key` | `gateway` | `gateway-client` || Helpdesk | `helpdesk-server.crt/key` | `helpdesk-client.crt/key` | `helpdesk-service` | `helpdesk-client` || Notification | `notification-server.crt/key` | `notification-client.crt/key` | `notification-service` | `notification-client` || Certificate | `certificate-server.crt/key` | `certificate-client.crt/key` | `certificate-service` | `certificate-client` || Container-Management | `container-management-server.crt/key` | `container-management-client.crt/key` | `container-management-service` | `container-management-client` || Client-Container | `client-container-{id}-server.crt/key` | `client-container-client.crt/key` | `client-container-{id}` | `client-container-client` |**Note**: gRPC can reuse the same certificates as HTTP/REST, or use separate certificates. For simplicity, we'll use the same certificates.

## Implementation Plan

### Phase 1: Configuration Updates

#### 1.1 Update All Service Configs

Add gRPC mTLS configuration fields to all service configs:**Files to update:**

- `services/auth/internal/config/config.go`
- `services/gateway/internal/config/config.go`
- `services/helpdesk/internal/config/config.go`
- `services/notification/internal/config/config.go`
- `services/certificate/internal/config/config.go` (already has HTTP mTLS, add gRPC)
- `services/container-management/internal/config/config.go` (already has HTTP mTLS, add gRPC)
- `services/client-container/internal/config/config.go` (already has partial config)

**Configuration fields needed:**

```go
// gRPC mTLS Server Configuration
GRPCMTLSCACert        string
GRPCMTLSServerCert    string
GRPCMTLSServerKey     string
GRPCMTLSServerKeyPass string
GRPCMTLSTLSConfig     *tls.Config

// gRPC mTLS Client Configuration (for services that make gRPC calls)
GRPCMTLSClientCA      string
GRPCMTLSClientCert    string
GRPCMTLSClientKey     string
GRPCMTLSClientKeyPass string
GRPCMTLSClientConfig  *tls.Config
```

**Note**: For simplicity, gRPC can reuse HTTP mTLS certificates, or use separate ones. We'll support both approaches.

### Phase 2: gRPC Server mTLS Implementation

#### 2.1 Auth Service gRPC Server

**File**: `services/auth/cmd/server/main.go`

- Add function `initGRPCTLS(cfg *config.Config) credentials.TransportCredentials`
- Update `startGRPCServer` to use TLS credentials
- Load server certificate and CA cert for client verification
- Use `credentials.NewServerTLSFromFile` or custom TLS config

#### 2.2 Helpdesk Service gRPC Server

**File**: `services/helpdesk/cmd/server/main.go`

- Add gRPC TLS initialization
- Configure server with mTLS credentials
- Require and verify client certificates

#### 2.3 Notification Service gRPC Server

**File**: `services/notification/cmd/server/main.go`

- Add gRPC TLS initialization (internal service - mTLS required)
- Configure server with mTLS credentials
- Require and verify client certificates

#### 2.4 Certificate Service gRPC Server

**File**: `services/certificate/cmd/server/main.go`

- Add gRPC server (currently missing)
- Configure with mTLS credentials (internal service - mTLS required)
- Require and verify client certificates

#### 2.5 Container-Management Service gRPC Server

**File**: `services/container-management/cmd/server/main.go`

- Add gRPC server (currently missing)
- Configure with mTLS credentials
- Require and verify client certificates
- Use port 9005 for gRPC server

#### 2.6 Client-Container Service gRPC Server

**File**: `services/client-container/cmd/server/main.go`

- Complete existing partial gRPC TLS configuration
- Add client certificate verification (currently missing)
- Use `credentials.NewTLS` with proper `tls.Config` for mTLS

### Phase 3: gRPC Client mTLS Implementation

#### 3.1 Gateway Service gRPC Clients

**Files to update:**

- `services/gateway/internal/clients/auth_client.go`
- `services/gateway/internal/clients/helpdesk_client.go`
- `services/gateway/internal/clients/notification_client.go`

**Changes:**

- Replace `insecure.NewCredentials()` with TLS credentials
- Load client certificate and CA cert
- Use `credentials.NewTLS` with client TLS config
- Update `NewAuthClient`, `NewHelpDeskClient`, `NewNotificationClient` to accept TLS config

#### 3.2 Helpdesk Service gRPC Client

**File**: `services/helpdesk/internal/services/auth_client.go`

- Replace `insecure.NewCredentials()` with TLS credentials
- Load client certificate and CA cert for Auth service
- Update `newAuthClient` to use mTLS

#### 3.3 Container-Management Service gRPC Client (Optional)

**File**: `services/container-management/internal/services/certificate_client.go` (or new gRPC client)

- Optionally add gRPC client for Certificate service
- Currently uses HTTP mTLS, but can add gRPC client for consistency
- Load client certificate and CA cert for Certificate service
- Replace HTTP calls with gRPC calls (if implemented)

#### 3.4 Client-Container Service gRPC Client

**File**: `services/client-container/internal/services/notification_client.go`

- Replace `insecure.NewCredentials()` with TLS credentials
- Load client certificate and CA cert for Notification service
- Update `NewNotificationClient` to use mTLS

### Phase 4: Helper Functions and Utilities

#### 4.1 Create gRPC TLS Utilities

**New file**: `services/{service}/internal/utils/grpc_tls.go` (or add to existing utils)Create reusable functions:

- `LoadGRPCServerCredentials(certPath, keyPath, caPath string) (credentials.TransportCredentials, error)`
- `LoadGRPCClientCredentials(certPath, keyPath, caPath, serverName string) (credentials.TransportCredentials, error)`

These functions will:

- Load server/client certificates
- Load CA certificate for verification
- Create proper `tls.Config` with `RequireAndVerifyClientCert` for servers
- Create proper `tls.Config` with client cert and server verification for clients
- Return `credentials.TransportCredentials` ready for gRPC

### Phase 5: Environment Configuration

#### 5.1 Update Environment Files

Update `.env` files for each service with gRPC mTLS certificate paths:**For services with gRPC servers:**

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/{service}-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/{service}-server.crt
GRPC_MTLS_SERVER_KEY=/certs/{service}-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=
```

**For services with gRPC clients:**

```env
# gRPC mTLS Client Configuration (for each target service)
AUTH_SERVICE_GRPC_MTLS_CA=/certs/auth-ca.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/gateway-client.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/gateway-client.key

NOTIFICATION_SERVICE_GRPC_MTLS_CA=/certs/notification-ca.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/gateway-client.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/gateway-client.key
```



### Phase 6: Docker Compose Updates

#### 6.1 Certificate Volume Mounts

Ensure all services have certificate volumes mounted:

```yaml
volumes:
    - ./certs:/certs:ro
```



### Phase 7: Testing and Validation

#### 7.1 Test gRPC mTLS Connections

- Test Gateway → Auth gRPC with mTLS
- Test Gateway → Helpdesk gRPC with mTLS
- Test Gateway → Notification gRPC with mTLS
- Test Helpdesk → Auth gRPC with mTLS
- Test Client-Container → Notification gRPC with mTLS
- Test Container-Management gRPC server (if clients connect to it)
- Test Container-Management → Certificate gRPC with mTLS (if implemented)

#### 7.2 Test Certificate Validation

- Verify invalid certificates are rejected
- Verify expired certificates are rejected
- Verify certificate CN validation
- Test certificate rotation

## Detailed Implementation Steps

### Step 1: Create gRPC TLS Helper Functions

Create utility functions in each service (or shared package) for loading gRPC TLS credentials:

```go
// LoadGRPCServerCredentials loads TLS credentials for gRPC server with mTLS
func LoadGRPCServerCredentials(serverCert, serverKey, caCert string) (credentials.TransportCredentials, error) {
    // Load server certificate
    // Load CA cert for client verification
    // Create tls.Config with RequireAndVerifyClientCert
    // Return credentials.NewTLS(config)
}

// LoadGRPCClientCredentials loads TLS credentials for gRPC client with mTLS
func LoadGRPCClientCredentials(clientCert, clientKey, caCert, serverName string) (credentials.TransportCredentials, error) {
    // Load client certificate
    // Load CA cert for server verification
    // Create tls.Config with client cert and server verification
    // Return credentials.NewTLS(config)
}
```



### Step 2: Update Auth Service

**Files:**

- `services/auth/internal/config/config.go`: Add gRPC mTLS config fields
- `services/auth/cmd/server/main.go`: Update `startGRPCServer` to use TLS

**Changes:**

- Add gRPC mTLS config to Config struct
- Initialize gRPC TLS config in `Load()`
- Update `startGRPCServer` to use `grpc.Creds(credentials)` instead of `grpc.NewServer()`

### Step 3: Update Gateway Service

**Files:**

- `services/gateway/internal/config/config.go`: Add gRPC client mTLS configs
- `services/gateway/internal/clients/auth_client.go`: Use TLS credentials
- `services/gateway/internal/clients/helpdesk_client.go`: Use TLS credentials
- `services/gateway/internal/clients/notification_client.go`: Use TLS credentials

**Changes:**

- Add gRPC client TLS configs for Auth, Helpdesk, Notification
- Update all client constructors to accept and use TLS credentials
- Replace `insecure.NewCredentials()` with TLS credentials

### Step 4: Update Helpdesk Service

**Files:**

- `services/helpdesk/internal/config/config.go`: Add gRPC server and client mTLS configs
- `services/helpdesk/cmd/server/main.go`: Update gRPC server to use TLS
- `services/helpdesk/internal/services/auth_client.go`: Update client to use TLS

**Changes:**

- Add gRPC server mTLS config
- Add gRPC client mTLS config for Auth service
- Update server and client to use TLS credentials

### Step 5: Update Notification Service

**Files:**

- `services/notification/internal/config/config.go`: Add gRPC server mTLS config
- `services/notification/cmd/server/main.go`: Update gRPC server to use TLS

**Changes:**

- Add gRPC server mTLS config (internal service - mTLS required)
- Update server to use TLS credentials with client verification

### Step 6: Update Certificate Service

**Files:**

- `services/certificate/internal/config/config.go`: Add gRPC server mTLS config
- `services/certificate/cmd/server/main.go`: Add gRPC server with TLS

**Changes:**

- Add gRPC server mTLS config (internal service - mTLS required)
- Add gRPC server startup (currently missing)
- Configure with mTLS credentials

### Step 7: Update Container-Management Service

**Files:**

- `services/container-management/internal/config/config.go`: Add gRPC mTLS config
- `services/container-management/cmd/server/main.go`: Add gRPC server with TLS
- `services/container-management/internal/services/certificate_client.go`: Optionally add gRPC client

**Changes:**

- Add gRPC server mTLS configuration (server certs, CA cert)
- Add gRPC server startup on port 9005
- Optionally add gRPC client for Certificate service (currently uses HTTP)
- Configure with mTLS credentials and client verification

### Step 8: Complete Client-Container Service

**Files:**

- `services/client-container/internal/config/config.go`: Complete gRPC mTLS config
- `services/client-container/cmd/server/main.go`: Complete gRPC server TLS
- `services/client-container/internal/services/notification_client.go`: Update client to use TLS

**Changes:**

- Complete gRPC server mTLS configuration (add client verification)
- Update notification client to use TLS credentials
- Ensure proper CA cert pool for client verification

### Step 9: Update Docker Compose

**File**: `docker-compose.yml`

- Ensure all services have certificate volume mounts
- Verify certificate paths in environment variables
- Add gRPC port (9005) to container-management service if not already present

### Step 10: Generate Certificates

Generate gRPC certificates for all services (or reuse HTTP certificates):

- Server certificates for each service
- Client certificates for services that make gRPC calls
- CA certificates for verification

### Step 11: Update Environment Files

Generate/update `.env` files with gRPC mTLS certificate paths for all services, including container-management.

## Architecture

```javascript
┌─────────────┐         ┌─────────────┐
│   Gateway   │────────▶│    Auth     │
│  (gRPC mTLS)│         │ (gRPC mTLS) │
└─────────────┘         └─────────────┘
       │
       ├─────────────┐
       │             │
       ▼             ▼
┌─────────────┐  ┌─────────────┐
│  Helpdesk   │  │Notification │
│(gRPC mTLS)  │  │ (gRPC mTLS) │
└─────────────┘  └─────────────┘
       │
       ▼
┌─────────────┐
│    Auth     │
│ (gRPC mTLS) │
└─────────────┘

Internal Services (gRPC mTLS only):
- Certificate Service
- Container-Management Service
- Client-Container Service
- Notification Service (internal calls)

Container-Management → Certificate (HTTP mTLS or optional gRPC mTLS)
Client-Container → Notification (gRPC mTLS)
```



## Security Considerations

1. **Strict Certificate Validation**: Use `RequireAndVerifyClientCert` for all gRPC servers
2. **No Insecure Fallbacks**: Remove all `insecure.NewCredentials()` usage
3. **Certificate CN Validation**: Verify client certificate CN matches expected service
4. **Certificate Rotation**: Support certificate rotation without downtime
5. **Separate Certificates**: Optionally use separate certificates for gRPC vs HTTP (recommended for production)

## Testing Strategy

1. **Unit Tests**: Test TLS credential loading functions
2. **Integration Tests**: Test gRPC mTLS connections between services
3. **Negative Tests**: Test rejection of invalid/expired certificates
4. **Certificate Rotation Tests**: Test certificate updates without downtime
5. **Performance Tests**: Measure mTLS overhead on gRPC calls