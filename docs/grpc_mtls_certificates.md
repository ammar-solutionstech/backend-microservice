# gRPC mTLS Certificate Requirements

This document describes the certificate requirements for gRPC mTLS communication across all microservices.

## Overview

All gRPC services require mutual TLS (mTLS) certificates for secure communication. Certificates can be reused from HTTP/REST mTLS or generated separately for gRPC.

## Certificate Naming Convention

### Server Certificates

Each service requires a server certificate and key for its gRPC server:

| Service | Server Certificate | Server Key | CA Certificate | CN (Common Name) |
|---------|-------------------|------------|----------------|------------------|
| Auth | `auth-server.crt` | `auth-server.key` | `auth-ca.crt` | `auth-service` |
| Gateway | N/A (no gRPC server) | N/A | N/A | N/A |
| Helpdesk | `helpdesk-server.crt` | `helpdesk-server.key` | `helpdesk-ca.crt` | `helpdesk-service` |
| Notification | `notification-server.crt` | `notification-server.key` | `notification-ca.crt` | `notification-service` |
| Certificate | `certificate-server.crt` | `certificate-server.key` | `certificate-ca.crt` | `certificate-service` |
| Container-Management | `container-management-server.crt` | `container-management-server.key` | `container-management-ca.crt` | `container-management-service` |
| Client-Container | `client-container-{id}-server.crt` | `client-container-{id}-server.key` | `client-container-ca.crt` | `client-container-{id}` |

### Client Certificates

Services that make gRPC calls require client certificates:

| Service | Target Service | Client Certificate | Client Key | CA Certificate |
|---------|----------------|-------------------|------------|----------------|
| Gateway | Auth | `gateway-client.crt` | `gateway-client.key` | `auth-ca.crt` |
| Gateway | Helpdesk | `gateway-client.crt` | `gateway-client.key` | `helpdesk-ca.crt` |
| Gateway | Notification | `gateway-client.crt` | `gateway-client.key` | `notification-ca.crt` |
| Helpdesk | Auth | `helpdesk-client.crt` | `helpdesk-client.key` | `auth-ca.crt` |
| Client-Container | Notification | `client-container-client.crt` | `client-container-client.key` | `notification-ca.crt` |
| Container-Management | Certificate | `container-management-client.crt` | `container-management-client.key` | `certificate-ca.crt` |

## Certificate Reuse

**Option 1: Reuse HTTP/REST Certificates (Recommended for Development)**

For simplicity, gRPC can reuse the same certificates as HTTP/REST mTLS. In this case:
- Server certificates: Use the same server certificates as HTTP (e.g., `auth-server.crt` for both HTTP and gRPC)
- Client certificates: Use the same client certificates as HTTP (e.g., `gateway-client.crt` for both HTTP and gRPC)
- CA certificates: Use the same CA certificates

**Option 2: Separate gRPC Certificates (Recommended for Production)**

For production, it's recommended to use separate certificates for gRPC:
- Provides better security isolation
- Allows independent certificate rotation
- Easier to manage and audit

## Certificate Generation

### Using step-ca

Certificates should be generated using step-ca with the following requirements:

1. **Server Certificates**:
   - Common Name (CN) must match the service name (e.g., `auth-service`, `helpdesk-service`)
   - Subject Alternative Names (SANs) should include:
     - Service DNS name (e.g., `auth-service`)
     - Container name (e.g., `auth-service`)
     - Localhost (for local testing)
   - Extended Key Usage: Server Authentication
   - Key Usage: Digital Signature, Key Encipherment

2. **Client Certificates**:
   - Common Name (CN) should identify the client service (e.g., `gateway-client`, `helpdesk-client`)
   - Extended Key Usage: Client Authentication
   - Key Usage: Digital Signature, Key Encipherment

3. **CA Certificates**:
   - Each service should have its own CA certificate for client verification
   - Or use a shared CA for all services

### Example step-ca Certificate Request

For a server certificate:
```json
{
  "subject": {
    "commonName": "auth-service",
    "organization": ["ITaaS"],
    "organizationalUnit": ["Microservices"]
  },
  "sans": {
    "dns": ["auth-service", "localhost"],
    "ips": ["127.0.0.1"]
  },
  "keyUsage": ["digitalSignature", "keyEncipherment"],
  "extKeyUsage": ["serverAuth"]
}
```

For a client certificate:
```json
{
  "subject": {
    "commonName": "gateway-client",
    "organization": ["ITaaS"],
    "organizationalUnit": ["Microservices"]
  },
  "keyUsage": ["digitalSignature", "keyEncipherment"],
  "extKeyUsage": ["clientAuth"]
}
```

## Certificate Placement

All certificates should be placed in the `/certs` directory (mounted from `./certs` in docker-compose.yml):

```
certs/
├── auth-ca.crt
├── auth-server.crt
├── auth-server.key
├── gateway-client.crt
├── gateway-client.key
├── helpdesk-ca.crt
├── helpdesk-server.crt
├── helpdesk-server.key
├── helpdesk-client.crt
├── helpdesk-client.key
├── notification-ca.crt
├── notification-server.crt
├── notification-server.key
├── certificate-ca.crt
├── certificate-server.crt
├── certificate-server.key
├── container-management-ca.crt
├── container-management-server.crt
├── container-management-server.key
├── container-management-client.crt
├── container-management-client.key
├── client-container-ca.crt
├── client-container-server.crt
├── client-container-server.key
├── client-container-client.crt
└── client-container-client.key
```

## Environment Variables

Each service requires the following environment variables for gRPC mTLS:

### Services with gRPC Servers

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/{service}-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/{service}-server.crt
GRPC_MTLS_SERVER_KEY=/certs/{service}-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=  # Optional, if key is password-protected
```

### Services with gRPC Clients

```env
# For each target service
{TARGET}_SERVICE_GRPC_MTLS_CA=/certs/{target}-ca.crt
{TARGET}_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/{client}-client.crt
{TARGET}_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/{client}-client.key
```

## Certificate Validation

All gRPC servers are configured with:
- `ClientAuth: tls.RequireAndVerifyClientCert` - Requires and verifies client certificates
- `ClientCAs` - CA certificate pool for client certificate verification
- `MinVersion: tls.VersionTLS12` - Minimum TLS version 1.2

All gRPC clients are configured with:
- Client certificate for authentication
- `RootCAs` - CA certificate pool for server certificate verification
- `ServerName` - Expected server name for certificate validation

## Certificate Rotation

Certificates can be rotated without downtime:
1. Generate new certificates using step-ca
2. Update certificate files in `/certs` directory
3. Restart services to load new certificates
4. Services will automatically use new certificates on next connection

## Troubleshooting

### Certificate Errors

1. **"certificate signed by unknown authority"**:
   - Verify CA certificate is correctly mounted and readable
   - Check CA certificate path in environment variables

2. **"certificate has expired or is not yet valid"**:
   - Check certificate validity period
   - Regenerate certificates if expired

3. **"certificate Common Name mismatch"**:
   - Verify CN matches expected service name
   - Check ServerName configuration in client

4. **"client certificate required"**:
   - Verify client certificate is correctly configured
   - Check client certificate path in environment variables

### Testing Certificate Configuration

1. Verify certificates are readable:
   ```bash
   docker exec -it auth-service ls -la /app/certs/
   ```

2. Check certificate details:
   ```bash
   docker exec -it auth-service openssl x509 -in /app/certs/auth-server.crt -text -noout
   ```

3. Test gRPC connection:
   ```bash
   docker exec -it gateway grpcurl -cacert /app/certs/auth-ca.crt \
     -cert /app/certs/gateway-client.crt \
     -key /app/certs/gateway-client.key \
     auth-service:9001 list
   ```

## Security Best Practices

1. **Use separate certificates for production**: Don't reuse HTTP certificates in production
2. **Rotate certificates regularly**: Set up automated certificate rotation
3. **Protect private keys**: Ensure private keys are not exposed in logs or version control
4. **Use strong key sizes**: Use at least 2048-bit RSA or 256-bit ECDSA keys
5. **Set appropriate validity periods**: Balance security (shorter) with operational convenience (longer)
6. **Monitor certificate expiration**: Set up alerts for upcoming certificate expiration

