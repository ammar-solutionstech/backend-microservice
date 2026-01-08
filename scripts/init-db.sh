#!/bin/bash
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -U postgres; do
  sleep 1
done

echo "Creating databases..."
psql -v ON_ERROR_STOP=1 -U postgres <<-EOSQL
    SELECT 'CREATE DATABASE auth_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'auth_db')\gexec
    
    SELECT 'CREATE DATABASE helpdesk_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'helpdesk_db')\gexec
    
    SELECT 'CREATE DATABASE notification_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'notification_db')\gexec
    
    SELECT 'CREATE DATABASE certificate_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'certificate_db')\gexec
    
    SELECT 'CREATE DATABASE container_mgmt_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'container_mgmt_db')\gexec
    
    SELECT 'CREATE DATABASE inventory_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'inventory_db')\gexec
    
    SELECT 'CREATE DATABASE geography_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'geography_db')\gexec
    
    SELECT 'CREATE DATABASE navigation_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'navigation_db')\gexec
    
    SELECT 'CREATE DATABASE "ITaaS"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ITaaS')\gexec
EOSQL

echo "Running Auth Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d auth_db -f /migrations/auth/001_initial_schema.sql

echo "Running Help Desk Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d helpdesk_db -f /migrations/helpdesk/001_initial_schema.sql

echo "Running Notification Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d notification_db -f /migrations/notification/001_initial_schema.sql

echo "Running Certificate Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d certificate_db -f /migrations/certificate/001_initial_schema.sql

echo "Running Container Management Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d container_mgmt_db -f /migrations/container-management/001_initial_schema.sql

echo "Running Inventory Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d inventory_db -f /migrations/inventory/001_initial_schema.sql

echo "Running Geography Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d geography_db -f /migrations/geography/001_initial_schema.sql

echo "Running Navigation Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d navigation_db -f /migrations/navigation/001_initial_schema.sql

echo "Database initialization complete!"