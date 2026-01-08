# PowerShell integration test script
# Tests key endpoints through the Gateway

param(
    [string]$GatewayUrl = "http://localhost:8080",
    [string]$TestEmail = "",
    [string]$TestPassword = ""
)

$ErrorActionPreference = "Stop"

Write-Host "=== Integration Tests ===" -ForegroundColor Cyan
Write-Host ""

# Test health checks
Write-Host "1. Testing health checks..." -ForegroundColor Yellow
try {
    $response = Invoke-WebRequest -Uri "$GatewayUrl/health" -Method Get -UseBasicParsing
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Gateway health check passed" -ForegroundColor Green
    } else {
        Write-Host "✗ Gateway health check failed (Status: $($response.StatusCode))" -ForegroundColor Red
    }
} catch {
    Write-Host "✗ Gateway health check failed: $_" -ForegroundColor Red
}

# Test authentication (if credentials are provided)
if ($TestEmail -and $TestPassword) {
    Write-Host ""
    Write-Host "2. Testing authentication..." -ForegroundColor Yellow
    try {
        $body = @{
            email = $TestEmail
            password = $TestPassword
        } | ConvertTo-Json
        
        $response = Invoke-RestMethod -Uri "$GatewayUrl/api/auth/login" -Method Post -Body $body -ContentType "application/json"
        
        if ($response.access_token) {
            $script:AccessToken = $response.access_token
            Write-Host "✓ Authentication successful" -ForegroundColor Green
            Write-Host "  Token: $($AccessToken.Substring(0, [Math]::Min(20, $AccessToken.Length)))..." -ForegroundColor Gray
        } else {
            Write-Host "✗ Authentication failed: No token in response" -ForegroundColor Red
        }
    } catch {
        Write-Host "✗ Authentication failed: $_" -ForegroundColor Red
    }
}

# Test protected endpoints (if token is available)
if ($script:AccessToken) {
    Write-Host ""
    Write-Host "3. Testing protected endpoints..." -ForegroundColor Yellow
    
    $headers = @{
        "Authorization" = "Bearer $($script:AccessToken)"
    }
    
    # Test inventory endpoints
    try {
        $response = Invoke-WebRequest -Uri "$GatewayUrl/api/brands" -Method Get -Headers $headers -UseBasicParsing
        Write-Host "✓ Inventory Service (brands) accessible" -ForegroundColor Green
    } catch {
        Write-Host "✗ Inventory Service (brands) failed: $_" -ForegroundColor Red
    }
    
    # Test geography endpoints
    try {
        $response = Invoke-WebRequest -Uri "$GatewayUrl/api/countries" -Method Get -Headers $headers -UseBasicParsing
        Write-Host "✓ Geography Service (countries) accessible" -ForegroundColor Green
    } catch {
        Write-Host "✗ Geography Service (countries) failed: $_" -ForegroundColor Red
    }
    
    # Test navigation endpoints
    try {
        $response = Invoke-WebRequest -Uri "$GatewayUrl/api/menus" -Method Get -Headers $headers -UseBasicParsing
        Write-Host "✓ Navigation Service (menus) accessible" -ForegroundColor Green
    } catch {
        Write-Host "✗ Navigation Service (menus) failed: $_" -ForegroundColor Red
    }
    
    # Test help desk endpoints
    try {
        $response = Invoke-WebRequest -Uri "$GatewayUrl/api/help-desk" -Method Get -Headers $headers -UseBasicParsing
        Write-Host "✓ Help Desk Service (tickets) accessible" -ForegroundColor Green
    } catch {
        Write-Host "✗ Help Desk Service (tickets) failed: $_" -ForegroundColor Red
    }
}

Write-Host ""
Write-Host "=== Integration Tests Complete ===" -ForegroundColor Cyan
Write-Host ""
Write-Host "Note: For comprehensive testing, see extra/integration_test_guide.md" -ForegroundColor Gray

