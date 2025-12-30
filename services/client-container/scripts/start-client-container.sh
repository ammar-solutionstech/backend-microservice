#!/bin/bash
set -e

# Start PostgreSQL in the background using the postgres entrypoint
echo "Starting PostgreSQL..."
docker-entrypoint.sh postgres &
POSTGRES_PID=$!

# Wait for PostgreSQL to be ready
echo "Waiting for PostgreSQL to be ready..."
until pg_isready -h localhost -U postgres; do
  sleep 1
done

echo "PostgreSQL is ready!"

# Wait a bit more to ensure PostgreSQL is fully initialized
sleep 2

# Start the client-container application
echo "Starting client-container application..."
cd /app
exec /app/client-container