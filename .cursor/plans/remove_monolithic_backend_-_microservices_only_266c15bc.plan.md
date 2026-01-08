---
name: Remove Monolithic Backend - Microservices Only
overview: Comprehensive plan to migrate all functionality from the monolithic backend (main.go) to microservices architecture and remove all monolithic code, keeping only the microservices implementation.
todos:
  - id: create-inventory-service
    content: Create Inventory Service with all inventory endpoints, models, and database migration
    status: completed
  - id: create-geography-service
    content: Create Geography Service with countries, cities, locations, contacts endpoints
    status: completed
  - id: create-navigation-service
    content: Create Navigation Service with menu and menu-role management endpoints
    status: completed
  - id: enhance-auth-service
    content: Verify and add missing Role/Permission management endpoints to Auth Service
    status: completed
  - id: verify-helpdesk-service
    content: Verify Help Desk Service has all endpoints from monolithic routes/helpdesk.go
    status: completed
  - id: create-proto-definitions
    content: Create proto definitions for Inventory, Geography, and Navigation services
    status: completed
    dependencies:
      - create-inventory-service
      - create-geography-service
      - create-navigation-service
  - id: generate-proto-code
    content: Generate Go code from proto files using protoc
    status: completed
    dependencies:
      - create-proto-definitions
  - id: update-gateway-clients
    content: Create gRPC clients in Gateway for Inventory, Geography, and Navigation services
    status: completed
    dependencies:
      - generate-proto-code
  - id: update-gateway-routes
    content: Replace legacy handlers in Gateway with proper routing to new services
    status: completed
    dependencies:
      - update-gateway-clients
  - id: remove-legacy-db
    content: Remove legacy database connection from Gateway config
    status: completed
    dependencies:
      - update-gateway-routes
  - id: update-docker-compose
    content: Add new services to docker-compose.yml and update Gateway configuration
    status: completed
    dependencies:
      - create-inventory-service
      - create-geography-service
      - create-navigation-service
  - id: migrate-data
    content: Export data from monolithic database and import to new service databases
    status: completed
    dependencies:
      - update-docker-compose
  - id: integration-testing
    content: Test all endpoints through Gateway and verify service-to-service communication
    status: completed
    dependencies:
      - migrate-data
      - remove-legacy-db
  - id: delete-monolithic-code
    content: Delete main.go, controllers/, routes/, monolithic services/, and config/config.go
    status: completed
    dependencies:
      - integration-testing
  - id: update-documentation
    content: Update all API documentation, setup guides, and implementation status
    status: completed
    dependencies:
      - delete-monolithic-code
---

# Remove Mo

nolithic Backend - Microservices Migration Plan

## Overview

This plan migrates all functionality from the monolithic backend (`main.go`) to dedicated microservices and removes all monolithic code. The goal is to have a pure microservices architecture where the Gateway routes all requests to appropriate services.

## Current State Analysis

### Monolithic Backend Components (to be migrated/removed)

- **Entry Point**: `main.go`
- **Routes**: `routes/` (8 route files)
- **Controllers**: `controllers/` (12 controller files)
- **Services**: `services/` (monolithic services: generic_service.go, user_service.go, equipment_service.go, helpdesk_service.go, navigation_service.go, help_desk_type_service.go)
- **Config**: `config/config.go`
- **Middleware**: `middleware/` (may be reusable by Gateway)
- **Models**: `models/` (28 models - some may be needed by microservices)

### Microservices Already Implemented

- ✅ Auth Service (`services/auth/`)
- ✅ Help Desk Service (`services/helpdesk/`)
- ✅ Notification Service (`services/notification/`)
- ✅ Certificate Service (`services/certificate/`)
- ✅ Client Container Service (`services/client-container/`)
- ✅ Container Management Service (`services/container-management/`)
- ✅ Gateway (`services/gateway/`) - has placeholder legacy handlers

## Migration Strategy

### Phase 1: Create New Microservices

#### 1.1 Inventory Service

**Purpose**: Handle all inventory-related operations**Endpoints to Migrate** (from `routes/inventory.go`):

- `/api/brands` - CRUD
- `/api/models` - CRUD
- `/api/equipment-types` - CRUD
- `/api/operating-systems` - CRUD
- `/api/software-categories` - CRUD
- `/api/software` - CRUD
- `/api/equipment` - CRUD
- `/api/equipment/{id}/software` - Link/unlink software
- `/api/equipment/{id}/help-desk` - Link/unlink help desk tickets
- `/api/equipment/{id}/user-history` - Track equipment assignments
- `/api/documents` - CRUD (with base64 binary support)
- `/api/maintenance` - CRUD

**Models to Migrate**:

- Brand, Model, EquipmentType, OperatingSystem, SoftwareCategory, Software
- Equipment, EquipmentSoftware, EquipmentHelpDesk, EquipmentUserHistory
- Document, Maintenance

**Service Structure**:

```javascript
services/inventory/
├── cmd/server/main.go
├── Dockerfile
├── migrations/001_initial_schema.sql
├── proto/inventory.proto
├── internal/
│   ├── config/config.go
│   ├── models/ (all inventory models)
│   ├── services/
│   │   ├── equipment_service.go (from services/equipment_service.go)
│   │   ├── generic_service.go (copy pattern)
│   │   └── inventory_service.go
│   ├── routes/
│   │   └── inventory_controller.go
│   ├── middleware/
│   │   └── cors_middleware.go
│   └── utils/
```

**Database**: New `inventory_db` database

#### 1.2 Geography Service

**Purpose**: Handle geographic reference data**Endpoints to Migrate** (from `routes/geography.go`):

- `/api/countries` - CRUD
- `/api/cities` - CRUD
- `/api/locations` - CRUD
- `/api/contacts` - CRUD

**Models to Migrate**:

- Country, City, Location, Contact, Nationality

**Service Structure**:

```javascript
services/geography/
├── cmd/server/main.go
├── Dockerfile
├── migrations/001_initial_schema.sql
├── proto/geography.proto
├── internal/
│   ├── config/config.go
│   ├── models/ (geography models)
│   ├── services/
│   │   └── geography_service.go
│   ├── routes/
│   │   └── geography_controller.go
│   └── middleware/
```

**Database**: New `geography_db` database**Alternative**: Could merge with Inventory Service if geographic data is tightly coupled with inventory.z

#### 1.3 Navigation Service

**Purpose**: Handle menu and navigation management**Endpoints to Migrate** (from `routes/navigation.go`):

- `/api/menus` - CRUD
- `/api/menus/{id}/roles` - Menu-role relationships

**Models to Migrate**:

- Menu, MenuRole

**Service Structure**:

```javascript
services/navigation/
├── cmd/server/main.go
├── Dockerfile
├── migrations/001_initial_schema.sql
├── proto/navigation.proto
├── internal/
│   ├── config/config.go
│   ├── models/ (menu models)
│   ├── services/
│   │   └── navigation_service.go (from services/navigation_service.go)
│   └── routes/
│       └── navigation_controller.go
```

**Database**: New `navigation_db` database**Alternative**: Could merge with Auth Service since menus are permission/role-based.

### Phase 2: Enhance Existing Services

#### 2.1 Auth Service - Add Roles & Permissions Management

**Current**: Has User, Role, Permission models but may need REST endpoints**Add Endpoints** (from `routes/role.go`, `routes/permission.go`):

- `/api/roles` - CRUD + role-permission management
- `/api/permissions` - CRUD
- `/api/users/{id}/roles` - User role assignment
- `/api/users/{id}/permissions` - User permission assignment

**Verify**: Check if these endpoints already exist in `services/auth/internal/routes/auth_controller.go`

#### 2.2 Help Desk Service - Verify Completeness

**Verify**: All endpoints from `routes/helpdesk.go` are implemented:

- `/api/help-desk` - CRUD
- `/api/help-desk/ratings` - CRUD
- `/api/help-desk/transactions` - CRUD
- `/api/help-desk/teams` - CRUD
- `/api/help-desk/teams/{id}/members` - Team member management
- `/api/help-desk/{id}/participants` - Participant management
- `/api/help-desk/transactions/{id}/users` - Transaction user assignment
- `/api/help-desk/transaction-types` - CRUD
- `/api/help-desk/types` - CRUD (from `routes/help_desk_type.go`)

### Phase 3: Update Gateway

#### 3.1 Replace Legacy Handlers

**File**: `services/gateway/internal/handlers/legacy_handler.go`**Replace with**:

- Inventory Service gRPC client
- Geography Service gRPC client
- Navigation Service gRPC client
- Update Auth Service routes (if needed)
- Update Help Desk Service routes (verify completeness)

#### 3.2 Update Gateway Routes

**File**: `services/gateway/internal/routes/router.go`**Changes**:

- Remove legacy handler routes
- Add Inventory Service routes
- Add Geography Service routes
- Add Navigation Service routes
- Ensure all routes use Auth middleware for JWT validation

#### 3.3 Remove Legacy Database Connection

**File**: `services/gateway/internal/config/config.go`**Remove**:

- `LegacyDBHost`, `LegacyDBPort`, `LegacyDBName`, `LegacyDBUser`, `LegacyDBPassword`, `LegacyDBSchema`
- `LegacyDB` field
- `initLegacyDB()` function

### Phase 4: Update Docker Compose

**File**: `docker-compose.yml`**Add Services**:

- `inventory-service` (port 8007, gRPC 9007)
- `geography-service` (port 8008, gRPC 9008)
- `navigation-service` (port 8009, gRPC 9009)

**Add Databases**:

- `inventory_db`
- `geography_db`
- `navigation_db`

**Update Gateway**:

- Add environment variables for new service gRPC addresses
- Add mTLS certificate paths for new services
- Remove legacy database environment variables

### Phase 5: Code Migration

#### 5.1 Migrate Controllers

**From**: `controllers/resource_controller.go`, `controllers/resource_factories.go`**To**: Each new service's `internal/routes/` directory**Pattern**: Follow the pattern used in `services/helpdesk/internal/routes/ticket_controller.go`

#### 5.2 Migrate Services

**From**:

- `services/generic_service.go` → Copy pattern to each service
- `services/equipment_service.go` → `services/inventory/internal/services/equipment_service.go`
- `services/navigation_service.go` → `services/navigation/internal/services/navigation_service.go`
- `services/help_desk_type_service.go` → Verify if in Help Desk Service

#### 5.3 Migrate Models

**From**: `models/` directory**To**: Each service's `internal/models/` directory**Important**: Remove foreign key relationships between services. Services communicate via gRPC only.

#### 5.4 Create Proto Definitions

**Files to Create**:

- `services/inventory/proto/inventory.proto`
- `services/geography/proto/geography.proto`
- `services/navigation/proto/navigation.proto`

**Generate**: Run `protoc` to generate Go code

### Phase 6: Remove Monolithic Code

#### 6.1 Files to Delete

- `main.go` (monolithic entry point)
- `config/config.go` (monolithic config)
- `controllers/` directory (all 12 files)
- `routes/` directory (all 8 files)
- `services/generic_service.go`
- `services/user_service.go` (replaced by Auth Service)
- `services/equipment_service.go` (migrated to Inventory Service)
- `services/helpdesk_service.go` (migrated to Help Desk Service)
- `services/navigation_service.go` (migrated to Navigation Service)
- `services/help_desk_type_service.go` (verify if in Help Desk Service)
- `services/generic_service_test.go`

#### 6.2 Models Directory

**Decision Required**:

- Option A: Keep `models/` for reference/documentation, mark as deprecated
- Option B: Delete `models/` entirely (models are in each service)
- **Recommendation**: Delete after migration is verified

#### 6.3 Middleware Directory

**Decision Required**:

- Check if Gateway uses any middleware from `middleware/`
- If Gateway has its own middleware, delete `middleware/`
- If Gateway imports from `middleware/`, keep only what's needed

### Phase 7: Update Documentation

#### 7.1 Update API Documentation

**Files**: `docs/API*.md`

- Update base URLs to point to Gateway
- Remove references to monolithic endpoints
- Add new service endpoints

#### 7.2 Update Setup Documentation

**Files**: `README_SETUP.md`, `README_MICROSERVICES.md`

- Remove monolithic setup instructions
- Add new service setup instructions
- Update docker-compose instructions

#### 7.3 Update Implementation Status

**File**: `IMPLEMENTATION_STATUS.md`

- Mark monolithic backend as removed
- Update service status

## Implementation Steps

### Step 1: Create Inventory Service

1. Create service directory structure
2. Create database migration
3. Copy and adapt models from `models/`
4. Copy and adapt `services/equipment_service.go`
5. Create proto definition
6. Generate proto code
7. Implement REST controllers
8. Implement gRPC server
9. Create main.go
10. Add to docker-compose.yml
11. Test service independently

### Step 2: Create Geography Service

1. Create service directory structure
2. Create database migration
3. Copy and adapt models
4. Create proto definition
5. Generate proto code
6. Implement REST controllers
7. Implement gRPC server
8. Create main.go
9. Add to docker-compose.yml
10. Test service independently

### Step 3: Create Navigation Service

1. Create service directory structure
2. Create database migration
3. Copy and adapt models
4. Copy and adapt `services/navigation_service.go`
5. Create proto definition
6. Generate proto code
7. Implement REST controllers
8. Implement gRPC server
9. Create main.go
10. Add to docker-compose.yml
11. Test service independently

### Step 4: Enhance Auth Service

1. Verify existing endpoints
2. Add missing Role/Permission management endpoints if needed
3. Test endpoints

### Step 5: Verify Help Desk Service

1. Compare with monolithic routes
2. Add missing endpoints if any
3. Test all endpoints

### Step 6: Update Gateway

1. Create gRPC clients for new services
2. Replace legacy handlers with service clients
3. Update router.go
4. Remove legacy database config
5. Test routing

### Step 7: Integration Testing

1. Test all endpoints through Gateway
2. Test service-to-service communication
3. Test authentication flow
4. Test error handling

### Step 8: Data Migration

1. Export data from monolithic database
2. Import to new service databases
3. Verify data integrity
4. Test with real data

### Step 9: Remove Monolithic Code

1. Stop monolithic server
2. Delete files (Step 6.1)
3. Update go.mod (remove unused dependencies)
4. Clean up imports

### Step 10: Final Verification

1. Run all services via docker-compose
2. Test all API endpoints
3. Verify no references to monolithic code
4. Update documentation

## Dependencies & Considerations

### Database Migration Strategy

- **Option A**: Keep monolithic database as "legacy" and migrate data gradually
- **Option B**: Migrate all data at once to new databases
- **Recommendation**: Option B for clean separation

### Backward Compatibility

- Gateway should maintain same API paths (`/api/inventory/*`, etc.)
- Response formats should remain the same
- Authentication flow unchanged

### Service Communication

- All services communicate via gRPC with mTLS
- No direct database access between services
- Use gRPC clients in Gateway to route requests

### Error Handling

- Standardize error responses across services
- Gateway should normalize errors from services
- Maintain consistent HTTP status codes

## Risk Mitigation

1. **Data Loss Risk**: 

- Export all data before migration
- Test data migration scripts
- Keep backup of monolithic database

2. **Service Downtime**:

- Deploy new services alongside monolithic
- Route traffic gradually
- Keep monolithic running until verification complete

3. **Breaking Changes**:

- Maintain API compatibility in Gateway
- Version API if needed
- Document all changes

## Success Criteria

- ✅ All monolithic endpoints accessible through Gateway
- ✅ All services running independently
- ✅ No references to monolithic code
- ✅ All tests passing
- ✅ Documentation updated
- ✅ Docker compose starts all services successfully