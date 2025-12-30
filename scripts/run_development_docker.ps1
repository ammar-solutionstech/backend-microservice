# Navigate to project root
cd D:\backend_v1

# Set environment variables (or use .env file)
# $env:CERTIFICATE_DB_HOST="localhost"
# $env:CERTIFICATE_DB_PORT="5432"
# $env:CERTIFICATE_DB_NAME="certificate_db"
# $env:CERTIFICATE_DB_USER="postgres"
# $env:CERTIFICATE_DB_PASSWORD="postgres"
# $env:CERTIFICATE_PORT="8004"
$env:STEP_CA_URL="https://localhost:9000"

# Run step-ca container
# docker run -d --name stepca  -p 9000:9000 --rm -it -e "DOCKER_STEPCA_INIT_NAME=ITaaS CA" \
#  -e "DOCKER_STEPCA_INIT_DNS_NAMES=localhost" -e "DOCKER_STEPCA_INIT_PASSWORD=password" \
#  -v ${PWD}/../stepca/stepca:/home/step smallstep/step-ca
docker compose -f docker-compose.yml up -d --build 'step-ca'
Start-Sleep -Seconds 5

# Run DB container
docker compose -f docker-compose.yml up -d --build 'postgres-itaas'
Start-Sleep -Seconds 5

# Run # RabbitMQ
docker compose -f docker-compose.yml up -d --build 'rabbitmq'
Start-Sleep -Seconds 5

# Run Notification service
docker compose -f docker-compose.yml up -d --build 'notification-service'
Start-Sleep -Seconds 5

# Build client-container image
docker compose -f docker-compose.yml build --no-cache 'client-container'
# Verify image was created
#docker images | Select-String "client-container"

# Docker Daemon Accessible to Container-Management
# Ensure Docker Desktop is running
# The service will use: npipe:////./pipe/docker_engine
# Or set explicitly:
$env:DOCKER_HOST="npipe:////./pipe/docker_engine"

# mTLS Certificates Configured
# Create certs directory
# New-Item -ItemType Directory -Force -Path .\certs
# Verify certificates
# openssl x509 -in ./certs/ca.crt -text -noout
# openssl x509 -in ./certs/server.crt -text -noout
# openssl x509 -in ./certs/client.crt -text -noout