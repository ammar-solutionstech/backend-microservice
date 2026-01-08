# Integration Testing Guide

This guide outlines how to test all endpoints through the Gateway and verify service-to-service communication after the microservices migration.

## Prerequisites

1. All services must be running (via Docker Compose or individually)
2. All databases must be initialized and migrated
3. Gateway must be accessible on the configured port (default: 8080)
4. Authentication tokens must be obtained for protected endpoints

## Testing Checklist

### 1. Service Health Checks

Test that all services are running and healthy:

```bash
# Gateway
curl http://localhost:8080/health

# Auth Service
curl http://localhost:8001/health

# Help Desk Service
curl http://localhost:8002/health

# Inventory Service
curl http://localhost:8007/health

# Geography Service
curl http://localhost:8008/health

# Navigation Service
curl http://localhost:8009/health
```

### 2. Authentication Flow

Test the authentication endpoints through the Gateway:

```bash
# Register a user
curl -X POST http://localhost:8080/api/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123","first_name":"Test","last_name":"User"}'

# Login
curl -X POST http://localhost:8080/api/auth/login \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com","password":"password123"}'

# Save the access_token and refresh_token from the response
export ACCESS_TOKEN="your_access_token_here"
export REFRESH_TOKEN="your_refresh_token_here"

# Refresh token
curl -X POST http://localhost:8080/api/auth/refresh \
  -H "Content-Type: application/json" \
  -d "{\"refresh_token\":\"$REFRESH_TOKEN\"}"
```

### 3. Inventory Service Endpoints

Test all inventory endpoints through the Gateway:

```bash
# List brands
curl -X GET http://localhost:8080/api/brands \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Create brand
curl -X POST http://localhost:8080/api/brands \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Brand"}'

# List models
curl -X GET http://localhost:8080/api/models \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List equipment types
curl -X GET http://localhost:8080/api/equipment-types \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List operating systems
curl -X GET http://localhost:8080/api/operating-systems \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List software categories
curl -X GET http://localhost:8080/api/software-categories \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List software
curl -X GET http://localhost:8080/api/software \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List equipment
curl -X GET http://localhost:8080/api/equipment \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Get equipment by ID
curl -X GET http://localhost:8080/api/equipment/1 \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List equipment software
curl -X GET http://localhost:8080/api/equipment/1/software \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List documents
curl -X GET http://localhost:8080/api/documents \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List maintenance records
curl -X GET http://localhost:8080/api/maintenance \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 4. Geography Service Endpoints

Test all geography endpoints through the Gateway:

```bash
# List countries
curl -X GET http://localhost:8080/api/countries \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Create country
curl -X POST http://localhost:8080/api/countries \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Country","code":"TC","phone_code":"+1"}'

# List cities
curl -X GET http://localhost:8080/api/cities \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List locations
curl -X GET http://localhost:8080/api/locations \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List contacts
curl -X GET http://localhost:8080/api/contacts \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 5. Navigation Service Endpoints

Test all navigation endpoints through the Gateway:

```bash
# List menus
curl -X GET http://localhost:8080/api/menus \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Create menu
curl -X POST http://localhost:8080/api/menus \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Menu","path":"/test","icon":"test-icon","order_index":1,"is_active":true}'

# Get menu roles
curl -X GET http://localhost:8080/api/menus/1/roles \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 6. Help Desk Service Endpoints

Test all help desk endpoints through the Gateway:

```bash
# List tickets
curl -X GET http://localhost:8080/api/help-desk \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Create ticket
curl -X POST http://localhost:8080/api/help-desk \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"name":"Test Ticket","description":"Test description","help_desk_type_id":1,"portal_user_id":1,"state":"open"}'

# List ratings
curl -X GET "http://localhost:8080/api/help-desk/ratings?help_desk_id=1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List transactions
curl -X GET "http://localhost:8080/api/help-desk/transactions?help_desk_id=1" \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List teams
curl -X GET http://localhost:8080/api/help-desk/teams \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List transaction types
curl -X GET http://localhost:8080/api/help-desk/transaction-types \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

### 7. Auth Service - Roles and Permissions

Test role and permission endpoints through the Gateway:

```bash
# List roles
curl -X GET http://localhost:8080/api/roles \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# List permissions
curl -X GET http://localhost:8080/api/permissions \
  -H "Authorization: Bearer $ACCESS_TOKEN"
```

## Service-to-Service Communication Testing

### 1. Verify gRPC Communication

Check Gateway logs to ensure gRPC calls are being made:

```bash
# View Gateway logs
docker logs gateway

# Look for gRPC connection messages and any errors
```

### 2. Test Cross-Service References

Test that services can validate references to other services:

```bash
# Create equipment with user_id (should validate via Auth Service)
curl -X POST http://localhost:8080/api/equipment \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name":"Test Equipment",
    "description":"Test",
    "state":true,
    "serial_number":"SN123",
    "equipment_type_id":1,
    "model_id":1,
    "production_date":"2024-01-01",
    "ip_v4":"192.168.1.1",
    "ip_v6":"::1",
    "mac_address":"00:00:00:00:00:00",
    "operating_system_id":1,
    "location_id":1,
    "proccessors":"Intel",
    "country_of_region":1,
    "user_id":1,
    "warrantly_start_date":"2024-01-01",
    "warrantly_end_date":"2025-01-01"
  }'

# Verify that invalid user_id returns an error
```

### 3. Test Error Handling

Verify that errors are properly propagated:

```bash
# Test 404 on non-existent resource
curl -X GET http://localhost:8080/api/equipment/99999 \
  -H "Authorization: Bearer $ACCESS_TOKEN"

# Test 400 on invalid data
curl -X POST http://localhost:8080/api/brands \
  -H "Authorization: Bearer $ACCESS_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{}'

# Test 401 on missing/invalid token
curl -X GET http://localhost:8080/api/brands
```

## Automated Testing Script

A basic test script is provided in `extra/integration_tests.sh` (or `.ps1` for Windows).

## Expected Results

- All health checks return `200 OK`
- Authentication flow works end-to-end
- All CRUD operations work through Gateway
- Error responses are properly formatted
- Service-to-service communication logs show successful gRPC calls
- Invalid cross-service references return appropriate errors

## Troubleshooting

1. **Connection Refused**: Ensure all services are running
2. **401 Unauthorized**: Check token validity and expiration
3. **500 Internal Server Error**: Check service logs for details
4. **gRPC Errors**: Verify mTLS certificates are configured correctly
5. **Database Errors**: Ensure migrations have been run

## Next Steps

After successful integration testing:
1. Mark integration-testing task as complete
2. Proceed with deleting monolithic code
3. Update documentation

