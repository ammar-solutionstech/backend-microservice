# setup-prerequisites.ps1
$ErrorActionPreference = "Stop"

Write-Host "Setting up prerequisites..." -ForegroundColor Green

# 1. Generate certificates
Write-Host "`n1. Generating mTLS certificates..." -ForegroundColor Cyan
if (Test-Path ".\generate-certs.ps1") {
    .\generate-certs.ps1
} else {
    Write-Host "Warning: generate-certs.ps1 not found. Please generate certificates manually." -ForegroundColor Yellow
}

# 2. Build client-container image
Write-Host "`n2. Building client-container image..." -ForegroundColor Cyan
docker build -f ./services/client-container/Dockerfile -t client-container:latest .
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Failed to build client-container image" -ForegroundColor Red
    exit 1
}

# 3. Start PostgreSQL (if not running)
Write-Host "`n3. Starting PostgreSQL..." -ForegroundColor Cyan
docker-compose up -d postgres-itaas
Start-Sleep -Seconds 5

# 4. Start certificate service
Write-Host "`n4. Starting certificate service..." -ForegroundColor Cyan
docker-compose up -d certificate-service
Start-Sleep -Seconds 3

# 5. Start container-management service
Write-Host "`n5. Starting container-management service..." -ForegroundColor Cyan
docker-compose up -d container-management-service
Start-Sleep -Seconds 3

# 6. Verify services
Write-Host "`n6. Verifying services..." -ForegroundColor Cyan
Write-Host "Certificate service:" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8004/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "  ✓ Running on port 8004" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Not responding on port 8004" -ForegroundColor Red
}

Write-Host "Container-management service:" -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8005/health" -UseBasicParsing -TimeoutSec 5
    Write-Host "  ✓ Running on port 8005" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Not responding on port 8005" -ForegroundColor Red
}

Write-Host "`nPrerequisites setup complete!" -ForegroundColor Green