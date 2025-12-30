FROM postgres:15-alpine

# Copy init script
COPY scripts/init-db.sh /docker-entrypoint-initdb.d/
RUN chmod +x /docker-entrypoint-initdb.d/init-db.sh

# Copy migration files
COPY services/auth/migrations /migrations/auth
COPY services/helpdesk/migrations /migrations/helpdesk
COPY services/notification/migrations /migrations/notification
COPY services/certificate/migrations /migrations/certificate
COPY services/container-management/migrations /migrations/container-management

ENTRYPOINT ["docker-entrypoint.sh", "postgres"]