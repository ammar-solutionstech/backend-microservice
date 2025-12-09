---
name: Container-Management Docker Orchestration and Service Management
overview: Implement Docker container orchestration in container-management service to dynamically create, configure, and manage client-container instances. Each client organization gets one client-container Docker container. Container-management can add/remove services (plugins) from client-container instances. Services are implemented as plugins that can be dynamically loaded/unloaded.
todos:
  - id: docker-client
    content: Add Docker SDK dependency and create Docker client service with container operations (create, start, stop, restart, remove, status, update, list)
    status: completed
  - id: container-orchestrator
    content: Create container orchestrator service for generating Docker container configurations and creating client-container instances
    status: completed
    dependencies:
      - docker-client
  - id: service-manager
    content: Create service manager for managing services/plugins within client-container instances (add, remove, enable, disable, list, update)
    status: completed
    dependencies:
      - docker-client
  - id: add-model-comments
    content: Add detailed comments to all existing model files (organization.go, client_container.go) - clarify each field purpose, format, and list static value options where applicable
    status: completed
  - id: update-models
    content: Update client_container model with Docker fields (docker_container_id, docker_container_name, docker_status, services, container_config) and create container_services model - both with detailed field comments
    status: completed
    dependencies:
      - add-model-comments
  - id: update-migration
    content: Update migration to add Docker fields to client_containers table and create container_services table
    status: completed
    dependencies:
      - update-models
  - id: update-registration
    content: Update RegisterClientContainer to create Docker container via orchestrator, configure certificates/data, and start container
    status: completed
    dependencies:
      - container-orchestrator
      - update-models
  - id: lifecycle-endpoints
    content: Add lifecycle management endpoints (start, stop, restart, update, remove, status) to client_container_controller
    status: completed
    dependencies:
      - container-orchestrator
  - id: service-endpoints
    content: Create service_controller with endpoints for managing services in containers (add, list, get, update, enable, disable, remove)
    status: completed
    dependencies:
      - service-manager
  - id: client-container-dockerfile
    content: Create Dockerfile for client-container service with multi-stage build
    status: completed
  - id: service-registry
    content: Implement service registry and service interface in client-container for plugin management
    status: completed
  - id: internal-service-api
    content: Add internal service management API endpoints in client-container (requires mTLS) for container-management to manage services
    status: completed
    dependencies:
      - service-registry
  - id: update-config
    content: Update container-management config with Docker settings (DockerHost, ClientContainerImage, DockerNetwork)
    status: completed
  - id: update-docker-compose
    content: Update docker-compose.yml to mount Docker socket to container-management and add client-container build configuration
    status: completed
    dependencies:
      - client-container-dockerfile
  - id: update-main-server
    content: Update container-management main server to initialize Docker client, orchestrator, and service manager, and register new routes
    status: completed
    dependencies:
      - docker-client
      - container-orchestrator
      - service-manager
      - lifecycle-endpoints
      - service-endpoints
---

# Container-Management Docker Orchestration and Service Management

## Overview

Implement Docker container orchestration in the container-management service to dynamically create, configure, and manage client-container Docker container instances. Each client organization gets one client-container instance. Container-management can add/remove services (plugins) from client-container instances. Services are implemented as plugins that can be dynamically loaded/unloaded.

## Architecture

- **Plugin Architecture**: Services are plugins loaded into the single client-container instance
- **One Container Per Organization**: Each organization gets one client-container Docker container
- **Dynamic Service Management**: Container-management can add/remove services (both internal and external) from client-container instances
- **Docker Orchestration**: Container-management uses Docker API to create, start, stop, update, and remove containers

## Key Components

### 1. Docker Client Integration

**File: `services/container-management/internal/services/docker_client.go`**

- Docker SDK client (`github.com/docker/docker/client`)
- Functions:
- `CreateContainer(config)` - Create Docker container with configuration
- `StartContainer(containerID)` - Start container
- `StopContainer(containerID)` - Stop container
- `RestartContainer(containerID)` - Restart container
- `RemoveContainer(containerID)` - Remove container
- `GetContainerStatus(containerID)` - Get container status
- `UpdateContainer(containerID, config)` - Update container configuration
- `ListContainers(filters)` - List containers with filters

### 2. Container Orchestrator Service

**File: `services/container-management/internal/services/container_orchestrator.go`**

- Generates Docker container configuration for client-container instances
- Functions:
- `GenerateContainerConfig(orgID, containerID, containerData)` - Generate Docker container config
- `CreateClientContainer(orgID, containerID, containerData)` - Create and start container
- `ConfigureContainerEnvironment(orgID, containerID)` - Generate environment variables
- `ConfigureContainerVolumes(containerID)` - Configure volumes (certificates, data)
- `ConfigureContainerNetwork(containerID)` - Configure network settings
- `GenerateContainerName(containerID)` - Generate unique container name

### 3. Service/Plugin Management

**File: `services/container-management/internal/services/service_manager.go`**

- Manages services (plugins) within client-container instances
- Functions:
- `AddServiceToContainer(containerID, serviceConfig)` - Add service/plugin to container
- `RemoveServiceFromContainer(containerID, serviceName)` - Remove service/plugin
- `ListContainerServices(containerID)` - List active services in container
- `UpdateServiceInContainer(containerID, serviceName, config)` - Update service configuration
- `EnableService(containerID, serviceName)` - Enable a service
- `DisableService(containerID, serviceName)` - Disable a service

### 4. Update Client Container Service

**File: `services/container-management/internal/services/client_container_service.go`**

- Modify `RegisterClientContainer` to:
- Generate container configuration
- Create Docker container via orchestrator
- Configure certificates and data
- Start the container
- Register container for management
- Add functions:
- `GetContainerStatus(containerID)` - Get Docker container status
- `UpdateContainerConfiguration(containerID, updates)` - Update container config

### 5. Update Database Models

**File: `services/container-management/internal/models/organization.go`**

Add detailed comments to all fields:

- `ID` - Primary key, auto-incrementing integer
- `Name` - Organization name, required, max 255 characters
- `Domain` - Organization domain (e.g., "acme.example.com"), required, unique, max 255 characters
- `AdminEmail` - Organization administrator email address, required, max 255 characters
- `AdminPhone` - Organization administrator phone number, optional, max 255 characters
- `Status` - Organization status. Options: `active`, `suspended`, `revoked`. Default: `active`
- `CreatedAt` - Timestamp when organization was created
- `UpdatedAt` - Timestamp when organization was last updated

**File: `services/container-management/internal/models/client_container.go`**

Add detailed comments to existing fields, then add new Docker fields with comments:

- `ID` - Primary key, auto-incrementing integer
- `OrganizationID` - Foreign key to organizations table, required, indexed
- `ContainerID` - Unique container identifier (e.g., "container-123"), required, unique index, max 255 characters
- `Name` - Human-readable container name, optional, max 255 characters
- `Status` - Container status. Options: `active`, `inactive`. Default: `active`
- `CertificateSerial` - Serial number of container's certificate, optional, max 255 characters
- `ContainerEndpointURL` - Full URL endpoint for the container (e.g., "https://container.example.com:8006"), optional, max 500 characters
- `AdminEmail` - Container administrator email address, required, max 255 characters
- `AdminPhone` - Container administrator phone number, optional, max 255 characters
- `CreatedAt` - Timestamp when container was created
- `UpdatedAt` - Timestamp when container was last updated

New Docker fields (with detailed comments):

- `docker_container_id` (string) - Docker container ID from Docker daemon (64-character hex string), optional, max 255 characters
- `docker_container_name` (string) - Docker container name (format: `client-container-{container_id}`), optional, max 255 characters
- `docker_status` (string) - Docker container status. Options: `running`, `stopped`, `restarting`, `paused`, `exited`, `dead`. Optional, max 50 characters
- `services` (JSONB) - Active services/plugins configuration (map structure: `{"service_name": {"type": "internal|external", "config": {...}, "enabled": true/false}}`), optional
- `container_config` (JSONB) - Container configuration snapshot including environment variables, volumes, network settings, optional

**File: `services/container-management/internal/models/container_service.go`** (new)

Model for tracking services in containers with detailed comments:

- `ID` - Primary key, auto-incrementing integer
- `ContainerID` (string, FK) - Reference to client_container.container_id, required, indexed, max 255 characters
- `ServiceName` (string) - Unique service name within container (e.g., "telemetry", "device-management"), required, max 255 characters
- `ServiceType` (string) - Service type. Options: `internal` (built-in service), `external` (custom plugin). Required, max 50 characters
- `ServiceConfig` (JSONB) - Service-specific configuration (structure varies by service type), optional
- `Status` (string) - Service status. Options: `active`, `inactive`, `starting`, `stopping`, `error`. Default: `inactive`, max 50 characters
- `Enabled` (bool) - Whether service is enabled (can be disabled without removing), default: `true`
- `CreatedAt` (timestamp) - Timestamp when service was added to container
- `UpdatedAt` (timestamp) - Timestamp when service was last updated

**Note**: All model files must include detailed comments for each field explaining its purpose, format, constraints, and if applicable, list of valid values/options.

### 6. Update Migration

**File: `services/container-management/migrations/001_initial_schema.sql`**

- Add columns to `client_containers` table:
- `docker_container_id VARCHAR(255)`
- `docker_container_name VARCHAR(255)`
- `docker_status VARCHAR(50)`
- `services JSONB`
- `container_config JSONB`
- Create `container_services` table for service management

### 7. Lifecycle Management Endpoints

**File: `services/container-management/internal/routes/client_container_controller.go`**

Add endpoints:

- `POST /api/v1/containers/:container_id/start` - Start container
- `POST /api/v1/containers/:container_id/stop` - Stop container
- `POST /api/v1/containers/:container_id/restart` - Restart container
- `PUT /api/v1/containers/:container_id/update` - Update container configuration
- `DELETE /api/v1/containers/:container_id` - Remove container
- `GET /api/v1/containers/:container_id/status` - Get container status

### 8. Service Management Endpoints

**File: `services/container-management/internal/routes/service_controller.go`** (new)

Endpoints:

- `POST /api/v1/containers/:container_id/services` - Add service to container
- `GET /api/v1/containers/:container_id/services` - List container services
- `GET /api/v1/containers/:container_id/services/:service_name` - Get service details
- `PUT /api/v1/containers/:container_id/services/:service_name` - Update service
- `POST /api/v1/containers/:container_id/services/:service_name/enable` - Enable service
- `POST /api/v1/containers/:container_id/services/:service_name/disable` - Disable service
- `DELETE /api/v1/containers/:container_id/services/:service_name` - Remove service

### 9. Client-Container Dockerfile

**File: `services/client-container/Dockerfile`**

- Multi-stage build
- Copy application code
- Build Go binary
- Create minimal runtime image
- Expose ports (REST and gRPC)
- Set entrypoint

### 10. Client-Container Service Plugin Interface

**File: `services/client-container/internal/services/service_registry.go`** (new)

- Service registry for managing plugins
- Functions:
- `RegisterService(name, service)` - Register a service
- `UnregisterService(name)` - Unregister a service
- `GetService(name)` - Get service instance
- `ListServices()` - List all registered services
- `StartService(name)` - Start a service
- `StopService(name)` - Stop a service

**File: `services/client-container/internal/services/service_interface.go`** (new)

- Define service interface that all plugins must implement:
- `Name() string`
- `Start() error`
- `Stop() error`
- `Status() string`
- `Configure(config map[string]interface{}) error`

### 11. Client-Container Service Management API

**File: `services/client-container/internal/routes/service_controller.go`** (new)

Endpoints for container-management to manage services:

- `POST /api/v1/internal/services` - Add service (internal API, mTLS required)
- `GET /api/v1/internal/services` - List services
- `PUT /api/v1/internal/services/:name` - Update service
- `POST /api/v1/internal/services/:name/start` - Start service
- `POST /api/v1/internal/services/:name/stop` - Stop service
- `DELETE /api/v1/internal/services/:name` - Remove service

### 12. Update Container-Management Configuration

**File: `services/container-management/internal/config/config.go`**

Add:

- `DockerHost` - Docker daemon socket (default: `unix:///var/run/docker.sock`)
- `ClientContainerImage` - Client-container Docker image name
- `ClientContainerImageTag` - Image tag
- `DockerNetwork` - Docker network name for client containers

### 13. Update docker-compose.yml

**File: `docker-compose.yml`**

- Add Docker socket volume mount to container-management-service
- Add client-container build configuration (for building the image)
- Ensure container-management has access to Docker daemon

### 14. Update Main Server

**File: `services/container-management/cmd/server/main.go`**

- Initialize Docker client
- Initialize container orchestrator
- Initialize service manager
- Register new routes

## Implementation Steps

1. Add Docker SDK dependency to container-management service
2. Create Docker client service with container operations
3. Create container orchestrator service for generating configurations
4. Create service manager for managing services/plugins
5. **Add detailed comments to all existing model files** (organization.go, client_container.go) - clarify each field and list static value options
6. Update client-container database model with Docker fields (with detailed comments)
7. Create container_services model (with detailed comments)
8. Create container_services table migration
9. Update client-container service registration to create Docker container
10. Add lifecycle management endpoints (start, stop, restart, update, remove)
11. Add service management endpoints (add, remove, enable, disable services)
12. Create client-container Dockerfile
13. Implement service registry and interface in client-container
14. Add internal service management API in client-container
15. Update container-management configuration
16. Update docker-compose.yml with Docker socket mount
17. Update main server initialization

## Service Types

### Internal Services

- Telemetry Service
- Device Management Service
- Plugin Manager Service
- Certificate Service

### External Services

- Custom plugins deployed as separate processes/containers
- Third-party integrations
- Custom business logic services

## Security Considerations

- Container-management must have Docker socket access (use Docker socket proxy or proper permissions)
- Internal service management API in client-container requires mTLS
- Service configurations validated before deployment
- Container isolation using Docker networks
- Certificate-based authentication for all inter-service communication