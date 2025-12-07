# Container Management Microservice

This microservice acts as a secure front-facing API layer in front of the existing certificate service. All communication uses mutual TLS (mTLS), and the service enforces CSR validation, container-level certificate management, and certificate-based authorization using custom extensions.

## Features

- **mTLS Communication**: All requests to/from certificate service use mTLS
- **CSR Validation**: Centralized, extendable validation (Organization, CN, Country required)
- **Container Service Support**: Endpoints for container-level certificate requests
- **Authorization**: Custom certificate extensions for roles/permissions
- **Bootstrap**: Pre-shared token/API key for initial container registration

## Architecture

The service acts as a proxy/validation layer - it does not issue certificates directly. All certificate operations are forwarded to the existing certificate service via mTLS.

- **Front-facing API**: Clean REST API for certificate requests and management
- **mTLS Required**: All inter-service communication uses mTLS
- **CSR Validation**: Enforced before forwarding to certificate service
- **Bootstrap Security**: Tokens are hashed, single-use, time-limited
- **Authorization**: Role/permission checks via custom certificate extensions

## Prerequisites

1. **PostgreSQL Database**: A database named `container_mgmt_db` must exist
2. **Certificate Service**: The existing certificate service must be running and accessible
3. **mTLS Certificates**: 
   - CA certificate for validating client certificates
   - Server certificate and key for mTLS server
   - Client certificate and key for communicating with certificate service
   - CA certificate for validating certificate service server certificate

## Configuration

Add these environment variables to your `.env` file:

```env
# Container Management Service Database
CONTAINER_MGMT_DB_HOST=localhost
CONTAINER_MGMT_DB_PORT=5432
CONTAINER_MGMT_DB_NAME=container_mgmt_db
CONTAINER_MGMT_DB_USER=postgres
CONTAINER_MGMT_DB_PASSWORD=postgres

# Container Management Service Port
CONTAINER_MGMT_PORT=8005

# Certificate Service Connection (mTLS)
CERTIFICATE_SERVICE_URL=https://certificate-service:8004
CERTIFICATE_SERVICE_CA=/certs/ca.crt
CERTIFICATE_SERVICE_CERT=/certs/client.crt
CERTIFICATE_SERVICE_KEY=/certs/client.key
CERTIFICATE_SERVICE_KEY_PASSWORD=

# mTLS Server Configuration
MTLS_CA_CERT=/certs/ca.crt
MTLS_SERVER_CERT=/certs/server.crt
MTLS_SERVER_KEY=/certs/server.key
MTLS_SERVER_KEY_PASSWORD=

# Bootstrap Token
BOOTSTRAP_TOKEN_SECRET=your-secret-key-for-token-generation

# CSR Validation Rules (optional)
CSR_REQUIRED_ORG=
CSR_REQUIRED_COUNTRY=
```

## mTLS Setup

### Generating Certificates

You'll need to generate certificates for mTLS. Here's a basic example using OpenSSL:

```bash
# Create CA
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 365 -key ca.key -out ca.crt -subj "/CN=ContainerManagementCA"

# Create server certificate
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -subj "/CN=container-management-service"
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Create client certificate (for certificate service communication)
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj "/CN=container-management-client"
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt
```

Place all certificates in a `certs/` directory and mount it to `/certs` in the container.

### Certificate Extensions

The service uses custom certificate extensions for authorization:

- **OID for Roles**: `1.3.6.1.4.1.12345.1.1`
- **OID for Permissions**: `1.3.6.1.4.1.12345.1.2`

The extension values should be JSON arrays of strings, e.g.:
```json
["admin", "operator"]
```

## API Endpoints

### Bootstrap (No mTLS Required)

- `POST /api/v1/bootstrap/register` - Register container with bootstrap token
- `GET /api/v1/bootstrap/status` - Check bootstrap status

### Certificates (mTLS Required)

- `POST /api/v1/certificates/request` - Submit CSR (validates, forwards to certificate service)
- `GET /api/v1/certificates/:serial` - Get certificate details
- `POST /api/v1/certificates/:serial/revoke` - Revoke certificate
- `GET /api/v1/certificates` - List certificates

### Container Certificates (mTLS Required)

- `POST /api/v1/containers/:container_id/certificates/request` - Request cert for container
- `POST /api/v1/containers/:container_id/applications/:app_name/certificates/request` - Request cert for app
- `GET /api/v1/containers/:container_id/certificates` - List container's certificates
- `POST /api/v1/containers/:container_id/certificates/:serial/revoke` - Revoke container cert

### Health Check

- `GET /health` - Service health check

## Bootstrap Process

1. **Generate Bootstrap Token**: Admin generates a bootstrap token (via separate admin endpoint or CLI)
2. **Container Registration**: Container uses token in `POST /api/v1/bootstrap/register` to get initial certificate
3. **Token Marked as Used**: Token is marked as used, container receives certificate
4. **Future Requests**: Container uses mTLS with that certificate for all future requests

### Example Bootstrap Request

```bash
curl -X POST https://container-management:8005/api/v1/bootstrap/register \
  -H "Content-Type: application/json" \
  -d '{
    "bootstrap_token": "your-bootstrap-token",
    "container_id": "container-123",
    "name": "My Container",
    "csr_pem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----"
  }'
```

## CSR Validation

The service validates CSRs before forwarding them to the certificate service. Required fields:

- **Organization**: Required (must match `CSR_REQUIRED_ORG` if configured)
- **Common Name (CN)**: Required
- **Country**: Required (must match `CSR_REQUIRED_COUNTRY` if configured)

Validation is centralized and extendable - you can add new rules by implementing the `CSRValidationRule` interface.

## Certificate-Based Authorization

The service uses custom certificate extensions for authorization:

- **Roles**: Extracted from OID `1.3.6.1.4.1.12345.1.1`
- **Permissions**: Extracted from OID `1.3.6.1.4.1.12345.1.2`

Middleware can check for required roles or permissions before allowing access to endpoints.

## Example Usage

### Request Certificate for Container

```bash
curl -X POST https://container-management:8005/api/v1/containers/container-123/certificates/request \
  --cert client.crt \
  --key client.key \
  --cacert ca.crt \
  -H "Content-Type: application/json" \
  -d '{
    "csr_pem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----"
  }'
```

### Request Certificate for Application

```bash
curl -X POST https://container-management:8005/api/v1/containers/container-123/applications/myapp/certificates/request \
  --cert client.crt \
  --key client.key \
  --cacert ca.crt \
  -H "Content-Type: application/json" \
  -d '{
    "csr_pem": "-----BEGIN CERTIFICATE REQUEST-----\n...\n-----END CERTIFICATE REQUEST-----"
  }'
```

## Docker Integration

The service is included in `docker-compose.yml`. Make sure to:

1. Place mTLS certificates in a `certs/` directory
2. Set `BOOTSTRAP_TOKEN_SECRET` in your environment
3. Configure certificate paths in environment variables

## Security Considerations

- **mTLS Required**: All inter-service communication uses mTLS
- **Certificate Validation**: Client certificates validated against CA
- **CSR Validation**: Enforced before forwarding to certificate service
- **Bootstrap Security**: Tokens are hashed, single-use, time-limited
- **Authorization**: Role/permission checks via custom certificate extensions
- **No Direct step-ca Access**: Service only communicates with certificate service

## Troubleshooting

### mTLS Connection Issues

1. Verify certificates are correctly mounted in `/certs`
2. Check certificate paths in environment variables
3. Ensure CA certificates are valid and trusted
4. Verify certificate service is accessible and using HTTPS

### Bootstrap Token Issues

1. Check `BOOTSTRAP_TOKEN_SECRET` is set
2. Verify token hasn't expired
3. Ensure token hasn't been used already
4. Check container_id matches token (if specified)

### CSR Validation Failures

1. Ensure CSR contains required fields (Organization, CN, Country)
2. Check if `CSR_REQUIRED_ORG` or `CSR_REQUIRED_COUNTRY` are set and match
3. Verify CSR is in valid PEM format

## Notes

- Service acts as a proxy/validation layer - does not issue certificates directly
- All certificate operations forwarded to existing certificate service via mTLS
- CSR validation is centralized and extendable (add new rules easily)
- Container service can request certificates for itself and applications under it
- Bootstrap process allows containers to get initial certificate without mTLS
- After bootstrap, all communication requires mTLS with valid certificate

