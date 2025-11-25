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
    
    SELECT 'CREATE DATABASE "ITaaS"'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'ITaaS')\gexec
EOSQL

echo "Running Auth Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d auth_db -f /migrations/auth/001_initial_schema.sql

echo "Running Help Desk Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d helpdesk_db -f /migrations/helpdesk/001_initial_schema.sql

echo "Running Notification Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d notification_db -f /migrations/notification/001_initial_schema.sql

echo "Database initialization complete!"