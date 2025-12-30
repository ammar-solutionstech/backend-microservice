# Certificate Distribution Guide

This document describes how to distribute certificates from the central `./certs` directory to service-specific `certs/` directories.

## Overview

Each microservice now has its own `certs/` directory containing only the certificates it needs. This provides better isolation and security.

## Certificate Distribution

### Auth Service (`services/auth/certs/`)

Required certificates:
- `auth-ca.crt` - CA for client verification
- `auth-server.crt` - Server certificate
- `auth-server.key` - Server private key

### Gateway Service (`services/gateway/certs/`)

Required certificates:
- `auth-ca.crt` - CA for Auth service
- `helpdesk-ca.crt` - CA for Helpdesk service
- `notification-ca.crt` - CA for Notification service
- `gateway-client.crt` - Client certificate
- `gateway-client.key` - Client private key

### Helpdesk Service (`services/helpdesk/certs/`)

Required certificates:
- `helpdesk-ca.crt` - CA for client verification
- `helpdesk-server.crt` - Server certificate
- `helpdesk-server.key` - Server private key
- `auth-ca.crt` - CA for Auth service
- `helpdesk-client.crt` - Client certificate for Auth
- `helpdesk-client.key` - Client private key

### Notification Service (`services/notification/certs/`)

Required certificates:
- `notification-ca.crt` - CA for client verification
- `notification-server.crt` - Server certificate
- `notification-server.key` - Server private key

### Certificate Service (`services/certificate/certs/`)

Required certificates:
- `certificate-ca.crt` - CA for client verification
- `certificate-server.crt` - Server certificate
- `certificate-server.key` - Server private key

### Container-Management Service (`services/container-management/certs/`)

Required certificates:
- `container-management-ca.crt` - CA for client verification
- `container-management-server.crt` - Server certificate
- `container-management-server.key` - Server private key
- `certificate-ca.crt` - CA for Certificate service
- `container-management-client.crt` - Client certificate for Certificate
- `container-management-client.key` - Client private key

### Client-Container Service (`services/client-container/certs/`)

Required certificates:
- `client-container-ca.crt` - CA for client verification
- `client-container-server.crt` - Server certificate
- `client-container-server.key` - Server private key
- `notification-ca.crt` - CA for Notification service
- `client-container-client.crt` - Client certificate for Notification
- `client-container-client.key` - Client private key
- `container.crt` - Container-specific certificate (per container)
- `container.key` - Container-specific private key (per container)
- `agent-ca.crt` - Agent CA certificate

## Distribution Script

You can use the following PowerShell script to copy certificates from the central `./certs` directory to service-specific directories:

```powershell
# Certificate Distribution Script
$rootCerts = "D:\backend_v1\certs"
$services = @(
    @{
        Name = "auth"
        Certs = @("auth-ca.crt", "auth-server.crt", "auth-server.key")
    },
    @{
        Name = "gateway"
        Certs = @("auth-ca.crt", "helpdesk-ca.crt", "notification-ca.crt", "gateway-client.crt", "gateway-client.key")
    },
    @{
        Name = "helpdesk"
        Certs = @("helpdesk-ca.crt", "helpdesk-server.crt", "helpdesk-server.key", "auth-ca.crt", "helpdesk-client.crt", "helpdesk-client.key")
    },
    @{
        Name = "notification"
        Certs = @("notification-ca.crt", "notification-server.crt", "notification-server.key")
    },
    @{
        Name = "certificate"
        Certs = @("certificate-ca.crt", "certificate-server.crt", "certificate-server.key")
    },
    @{
        Name = "container-management"
        Certs = @("container-management-ca.crt", "container-management-server.crt", "container-management-server.key", "certificate-ca.crt", "container-management-client.crt", "container-management-client.key")
    },
    @{
        Name = "client-container"
        Certs = @("client-container-ca.crt", "client-container-server.crt", "client-container-server.key", "notification-ca.crt", "client-container-client.crt", "client-container-client.key")
    }
)

foreach ($service in $services) {
    $targetDir = "D:\backend_v1\services\$($service.Name)\certs"
    if (-not (Test-Path $targetDir)) {
        New-Item -ItemType Directory -Path $targetDir -Force | Out-Null
    }
    
    foreach ($cert in $service.Certs) {
        $source = Join-Path $rootCerts $cert
        $target = Join-Path $targetDir $cert
        
        if (Test-Path $source) {
            Copy-Item $source $target -Force
            Write-Host "Copied $cert to $targetDir"
        } else {
            Write-Warning "Certificate $cert not found in $rootCerts"
        }
    }
}

Write-Host "Certificate distribution complete!"
```

## Manual Distribution

If you prefer to distribute certificates manually:

1. Navigate to the central `./certs` directory
2. Copy the required certificates for each service to its `services/{service-name}/certs/` directory
3. Ensure file permissions are correct (read-only for certificates, private keys should have restricted permissions)

## Certificate Generation

If certificates don't exist yet, you'll need to generate them using your Certificate Authority (step-ca or other). Refer to the gRPC mTLS certificate documentation for naming conventions and requirements.

## Verification

After distribution, verify that:
1. Each service's `certs/` directory contains only the required certificates
2. Certificate paths in `.env` files match the actual file locations
3. Docker containers can access the certificates (check volume mounts in `docker-compose.yml`)

