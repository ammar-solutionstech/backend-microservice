# generate-certs.ps1
$ErrorActionPreference = "Stop"

Write-Host "Generating mTLS certificates..." -ForegroundColor Green

# Create certs directory
New-Item -ItemType Directory -Force -Path .\certs | Out-Null
cd .\certs

# Check if OpenSSL is available
$openssl = Get-Command openssl -ErrorAction SilentlyContinue
if (-not $openssl) {
    Write-Host "Error: OpenSSL is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install OpenSSL or use an alternative method to generate certificates" -ForegroundColor Yellow
    exit 1
}

# Generate CA
Write-Host "Generating CA..." -ForegroundColor Cyan
openssl genrsa -out ca.key 4096
openssl req -new -x509 -days 3650 -key ca.key -out ca.crt -subj "/CN=ITaaS CA/O=ITaaS"

# Generate Server Certificate
Write-Host "Generating server certificate..." -ForegroundColor Cyan
openssl genrsa -out server.key 4096
openssl req -new -key server.key -out server.csr -subj "/CN=container-management-service/O=ITaaS"
openssl x509 -req -days 365 -in server.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out server.crt

# Generate Client Certificate
Write-Host "Generating client certificate..." -ForegroundColor Cyan
openssl genrsa -out client.key 4096
openssl req -new -key client.key -out client.csr -subj "/CN=container-management-client/O=ITaaS"
openssl x509 -req -days 365 -in client.csr -CA ca.crt -CAkey ca.key -CAcreateserial -out client.crt

# Cleanup
Remove-Item *.csr -ErrorAction SilentlyContinue
Remove-Item *.srl -ErrorAction SilentlyContinue

Write-Host "Certificates generated successfully in .\certs directory!" -ForegroundColor Green
Write-Host "Files created:" -ForegroundColor Cyan
Get-ChildItem -File | Select-Object Name, Length, LastWriteTime