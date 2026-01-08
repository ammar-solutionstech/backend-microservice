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

# Generate Inventory Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/inventory/proto/inventory.proto

# Generate Geography Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/geography/proto/geography.proto

# Generate Navigation Service proto
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/navigation/proto/navigation.proto

# Generate Agent proto (for agent program)
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    agent/proto/agent.proto

# Generate Client Container Agent proto (for backend service)
protoc --go_out=. --go_opt=paths=source_relative `
    --go-grpc_out=. --go-grpc_opt=paths=source_relative `
    services/client-container/proto/agent.proto

Write-Host "Proto files generated successfully"
