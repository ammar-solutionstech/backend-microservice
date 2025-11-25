# PowerShell script to generate protobuf files
# Requires protoc to be installed

Write-Host "Generating protobuf files..."

# Generate Auth Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/auth/proto/auth.proto

# Generate Help Desk Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/helpdesk/proto/helpdesk.proto

# Generate Notification Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/notification/proto/notification.proto

Write-Host "Proto files generated successfully"

