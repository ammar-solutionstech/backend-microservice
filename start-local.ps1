# start-local.ps1
Write-Host "Starting services locally..." -ForegroundColor Green

# Start Auth Service
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd D:\backend_v1; Write-Host 'Auth Service' -ForegroundColor Cyan; go run ./services/auth/cmd/server"

Start-Sleep -Seconds 2

# Start Help Desk Service
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd D:\backend_v1; Write-Host 'Help Desk Service' -ForegroundColor Cyan; go run ./services/helpdesk/cmd/server"

Start-Sleep -Seconds 2

# Start Notification Service
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd D:\backend_v1; Write-Host 'Notification Service' -ForegroundColor Cyan; go run ./services/notification/cmd/server"

Start-Sleep -Seconds 2

# Start Gateway
Start-Process powershell -ArgumentList "-NoExit", "-Command", "cd D:\backend_v1; Write-Host 'Gateway' -ForegroundColor Cyan; go run ./services/gateway/cmd/server"

Write-Host "All services started in separate windows!" -ForegroundColor Green