#!/bin/bash
# Data Migration Script - Run all data migrations
# This script exports data from the monolithic database (ITaaS) and imports to new service databases

set -e

echo "Starting data migration from monolithic database to microservices databases..."

# Set database connection parameters
MONOLITHIC_DB="${MONOLITHIC_DB:-ITaaS}"
MONOLITHIC_HOST="${MONOLITHIC_HOST:-localhost}"
MONOLITHIC_PORT="${MONOLITHIC_PORT:-5432}"
MONOLITHIC_USER="${MONOLITHIC_USER:-postgres}"

# Check if psql is available
if ! command -v psql &> /dev/null; then
    echo "Error: psql is not installed or not in PATH"
    exit 1
fi

echo "Migrating Inventory Service data..."
psql -h "$MONOLITHIC_HOST" -p "$MONOLITHIC_PORT" -U "$MONOLITHIC_USER" -d "$MONOLITHIC_DB" -f extra/migrate_inventory_data.sql

echo "Migrating Geography Service data..."
psql -h "$MONOLITHIC_HOST" -p "$MONOLITHIC_PORT" -U "$MONOLITHIC_USER" -d "$MONOLITHIC_DB" -f extra/migrate_geography_data.sql

echo "Migrating Navigation Service data..."
psql -h "$MONOLITHIC_HOST" -p "$MONOLITHIC_PORT" -U "$MONOLITHIC_USER" -d "$MONOLITHIC_DB" -f extra/migrate_navigation_data.sql

echo "Data migration completed successfully!"
echo ""
echo "Note: Verify data integrity by checking record counts in each database."
echo "Note: Some foreign key references (like user_id, location_id) are kept as integers"
echo "      and will need to be validated through gRPC calls to respective services."

