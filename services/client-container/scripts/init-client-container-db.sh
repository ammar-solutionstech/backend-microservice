#!/bin/bash
set -e

echo "Waiting for PostgreSQL to be ready..."
until pg_isready -U postgres; do
  sleep 1
done

echo "Creating databases..."
psql -v ON_ERROR_STOP=1 -U postgres <<-EOSQL
    SELECT 'CREATE DATABASE client_container_db'
    WHERE NOT EXISTS (SELECT FROM pg_database WHERE datname = 'client_container_db')\gexec
EOSQL

echo "Running Client Container Service migrations..."
psql -v ON_ERROR_STOP=1 -U postgres -d client_container_db -f /migrations/client-container/001_initial_schema.sql

echo "Database initialization complete!"
