#!/bin/bash

# Build script for cross-platform agent binaries

set -e

VERSION=${1:-"1.0.0"}
BUILD_TIME=$(date -u +"%Y-%m-%dT%H:%M:%SZ")
GIT_COMMIT=$(git rev-parse --short HEAD 2>/dev/null || echo "unknown")

# Build flags
LDFLAGS="-X backend/agent/internal/utils.Version=${VERSION} -X backend/agent/internal/utils.BuildTime=${BUILD_TIME} -X backend/agent/internal/utils.GitCommit=${GIT_COMMIT}"

# Build directories
BUILD_DIR="../../build/agent"
mkdir -p ${BUILD_DIR}

echo "Building agent binaries..."

# Build for Linux
echo "Building for Linux..."
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/update-service-linux-amd64 ./cmd/update-service
GOOS=linux GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/agent-core-linux-amd64 ./cmd/agent-core

# Build for Windows
echo "Building for Windows..."
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/update-service-windows-amd64.exe ./cmd/update-service
GOOS=windows GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/agent-core-windows-amd64.exe ./cmd/agent-core

# Build for macOS
echo "Building for macOS..."
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/update-service-darwin-amd64 ./cmd/update-service
GOOS=darwin GOARCH=amd64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/agent-core-darwin-amd64 ./cmd/agent-core

GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/update-service-darwin-arm64 ./cmd/update-service
GOOS=darwin GOARCH=arm64 go build -ldflags "${LDFLAGS}" -o ${BUILD_DIR}/agent-core-darwin-arm64 ./cmd/agent-core

echo "Build complete! Binaries are in ${BUILD_DIR}"

