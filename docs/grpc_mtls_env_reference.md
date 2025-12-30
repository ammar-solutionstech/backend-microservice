# gRPC mTLS Environment Variables Reference

This document provides a complete reference of all gRPC mTLS environment variables required for each service.

## Auth Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/auth-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/auth-server.crt
GRPC_MTLS_SERVER_KEY=/certs/auth-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=
```

## Gateway Service

```env
# Auth Service gRPC mTLS Client Configuration
AUTH_SERVICE_GRPC_MTLS_CA=/certs/auth-ca.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/gateway-client.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/gateway-client.key

# Helpdesk Service gRPC mTLS Client Configuration
HELPDESK_SERVICE_GRPC_MTLS_CA=/certs/helpdesk-ca.crt
HELPDESK_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/gateway-client.crt
HELPDESK_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/gateway-client.key

# Notification Service gRPC mTLS Client Configuration
NOTIFICATION_SERVICE_GRPC_MTLS_CA=/certs/notification-ca.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/gateway-client.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/gateway-client.key
```

## Helpdesk Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/helpdesk-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/helpdesk-server.crt
GRPC_MTLS_SERVER_KEY=/certs/helpdesk-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=

# Auth Service gRPC mTLS Client Configuration
AUTH_SERVICE_GRPC_MTLS_CA=/certs/auth-ca.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/helpdesk-client.crt
AUTH_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/helpdesk-client.key
```

## Notification Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/notification-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/notification-server.crt
GRPC_MTLS_SERVER_KEY=/certs/notification-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=
```

## Certificate Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/certificate-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/certificate-server.crt
GRPC_MTLS_SERVER_KEY=/certs/certificate-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=
```

## Container-Management Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/container-management-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/container-management-server.crt
GRPC_MTLS_SERVER_KEY=/certs/container-management-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=

# Certificate Service gRPC mTLS Client Configuration (optional)
CERTIFICATE_SERVICE_GRPC_MTLS_CA=/certs/certificate-ca.crt
CERTIFICATE_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/container-management-client.crt
CERTIFICATE_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/container-management-client.key
```

## Client-Container Service

```env
# gRPC mTLS Server Configuration
GRPC_MTLS_CA_CERT=/certs/client-container-ca.crt
GRPC_MTLS_SERVER_CERT=/certs/client-container-server.crt
GRPC_MTLS_SERVER_KEY=/certs/client-container-server.key
GRPC_MTLS_SERVER_KEY_PASSWORD=

# Notification Service gRPC mTLS Client Configuration
NOTIFICATION_SERVICE_GRPC_MTLS_CA=/certs/notification-ca.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_CERT=/certs/client-container-client.crt
NOTIFICATION_SERVICE_GRPC_MTLS_CLIENT_KEY=/certs/client-container-client.key
```

## Notes

1. **Certificate Reuse**: For development, you can reuse HTTP/REST certificates by pointing gRPC variables to the same certificate files.

2. **Optional Variables**: Variables marked as optional can be omitted if the service doesn't need that functionality.

3. **Password-Protected Keys**: If private keys are password-protected, set the `*_KEY_PASSWORD` variable. Otherwise, leave it empty.

4. **Default Paths**: All certificates are expected to be in `/certs` directory (mounted from `./certs` in docker-compose.yml).

5. **Fallback Behavior**: If gRPC mTLS certificates are not configured, services will fall back to insecure connections with a warning log message.

