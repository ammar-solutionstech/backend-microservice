# Protocol Buffers Code Generation

## Critical: Generate Proto Code Before Compilation

The `.proto` files define the gRPC services, but the Go code must be generated before the services can compile.

## Quick Start

1. **Install protoc** (if not already installed):
   - Windows: Download from https://github.com/protocolbuffers/protobuf/releases
   - Or use: `choco install protoc`
   - macOS: `brew install protobuf`
   - Linux: `apt-get install protobuf-compiler`

2. **Install Go plugins**:
   ```bash
   go install google.golang.org/protobuf/cmd/protoc-gen-go@latest
   go install google.golang.org/grpc/cmd/protoc-gen-go-grpc@latest
   ```

3. **Generate code**:
   ```bash
   # Windows PowerShell
   .\generate-proto.ps1
   
   # Or manually:
   protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/auth/proto/auth.proto
   protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/helpdesk/proto/helpdesk.proto
   protoc --go_out=. --go_opt=paths=source_relative --go-grpc_out=. --go-grpc_opt=paths=source_relative services/notification/proto/notification.proto
   ```

## What Gets Generated

After running protoc, these files will be created/updated:
- `services/auth/proto/auth.pb.go` - Message types
- `services/auth/proto/auth_grpc.pb.go` - gRPC service interfaces
- `services/helpdesk/proto/helpdesk.pb.go` - Message types
- `services/helpdesk/proto/helpdesk_grpc.pb.go` - gRPC service interfaces
- `services/notification/proto/notification.pb.go` - Message types
- `services/notification/proto/notification_grpc.pb.go` - gRPC service interfaces

**Note**: The placeholder files will be overwritten with actual generated code.

## Verification

After generation, you should be able to compile:
```bash
go build ./services/auth/cmd/server
go build ./services/helpdesk/cmd/server
go build ./services/notification/cmd/server
go build ./services/gateway/cmd/server
```

