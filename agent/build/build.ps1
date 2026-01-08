# Build script for Windows PowerShell

param(
    [string]$Version = "1.0.0"
)

$ErrorActionPreference = "Stop"

$BuildTime = Get-Date -Format "yyyy-MM-ddTHH:mm:ssZ"
$GitCommit = git rev-parse --short HEAD 2>$null
if (-not $GitCommit) {
    $GitCommit = "unknown"
}

# Build flags
$LDFlags = "-X backend/agent/internal/utils.Version=$Version -X backend/agent/internal/utils.BuildTime=$BuildTime -X backend/agent/internal/utils.GitCommit=$GitCommit"

# Build directory
$BuildDir = "../../build/agent"
New-Item -ItemType Directory -Force -Path $BuildDir | Out-Null

Write-Host "Building agent binaries..."

# Build for Windows
Write-Host "Building for Windows..."
$env:GOOS = "windows"
$env:GOARCH = "amd64"
go build -ldflags $LDFlags -o "$BuildDir/update-service-windows-amd64.exe" ../cmd/update-service
go build -ldflags $LDFlags -o "$BuildDir/agent-core-windows-amd64.exe" ../cmd/agent-core

# Build for Linux
Write-Host "Building for Linux..."
$env:GOOS = "linux"
$env:GOARCH = "amd64"
go build -ldflags $LDFlags -o "$BuildDir/update-service-linux-amd64" ../cmd/update-service
go build -ldflags $LDFlags -o "$BuildDir/agent-core-linux-amd64" ../cmd/agent-core

# Build for macOS
Write-Host "Building for macOS..."
$env:GOOS = "darwin"
$env:GOARCH = "amd64"
go build -ldflags $LDFlags -o "$BuildDir/update-service-darwin-amd64" ../cmd/update-service
go build -ldflags $LDFlags -o "$BuildDir/agent-core-darwin-amd64" ../cmd/agent-core

$env:GOOS = "darwin"
$env:GOARCH = "arm64"
go build -ldflags $LDFlags -o "$BuildDir/update-service-darwin-arm64" ../cmd/update-service
go build -ldflags $LDFlags -o "$BuildDir/agent-core-darwin-arm64" ../cmd/agent-core

Write-Host "Build complete! Binaries are in $BuildDir"

