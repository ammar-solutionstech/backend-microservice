# Agent Program

Cross-platform Agent Program for managing customer devices and integrating with the Client Container system.

## Architecture

The Agent Program consists of two separate services:

1. **Update Service** - Handles secure self-updates with rollback capability
2. **Agent Core Service** - Manages device registration, plugins, and communication with backend

## Building

See `build/` directory for platform-specific build scripts.

## Configuration

Configuration is managed through `config/` package with platform-specific paths for secure credential storage.

