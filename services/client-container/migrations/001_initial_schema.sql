-- Client Container Service Database Schema

-- Client containers table
CREATE TABLE IF NOT EXISTS client_containers (
    id SERIAL PRIMARY KEY,
    organization_id INTEGER NOT NULL,
    container_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    certificate_serial VARCHAR(255),
    container_endpoint_url VARCHAR(500),
    admin_email VARCHAR(255) NOT NULL,
    admin_phone VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_client_containers_organization_id ON client_containers(organization_id);
CREATE INDEX IF NOT EXISTS idx_client_containers_container_id ON client_containers(container_id);
CREATE INDEX IF NOT EXISTS idx_client_containers_status ON client_containers(status);

-- Devices table
CREATE TABLE IF NOT EXISTS devices (
    id SERIAL PRIMARY KEY,
    container_id INTEGER NOT NULL,
    device_id VARCHAR(255) NOT NULL,
    hostname VARCHAR(255),
    os_type VARCHAR(50),
    os_version VARCHAR(255),
    architecture VARCHAR(50),
    ip_address VARCHAR(45),
    device_serial VARCHAR(255) NOT NULL,
    last_seen TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    certificate_serial VARCHAR(255),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    registered_by INTEGER
);

CREATE INDEX IF NOT EXISTS idx_devices_container_id ON devices(container_id);
CREATE INDEX IF NOT EXISTS idx_devices_device_id ON devices(device_id);
CREATE INDEX IF NOT EXISTS idx_devices_status ON devices(status);
CREATE INDEX IF NOT EXISTS idx_devices_device_serial ON devices(device_serial);
CREATE UNIQUE INDEX IF NOT EXISTS idx_devices_container_device_id ON devices(container_id, device_id);

-- Telemetry table
CREATE TABLE IF NOT EXISTS telemetry (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL,
    metric_name VARCHAR(255) NOT NULL,
    metric_value JSONB,
    timestamp TIMESTAMP NOT NULL,
    tags JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_telemetry_device_id ON telemetry(device_id);
CREATE INDEX IF NOT EXISTS idx_telemetry_timestamp ON telemetry(timestamp);
CREATE INDEX IF NOT EXISTS idx_telemetry_metric_name ON telemetry(metric_name);
CREATE INDEX IF NOT EXISTS idx_telemetry_device_timestamp ON telemetry(device_id, timestamp);

-- Plugins table
CREATE TABLE IF NOT EXISTS plugins (
    id SERIAL PRIMARY KEY,
    container_id INTEGER NOT NULL,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    plugin_type VARCHAR(50),
    download_url VARCHAR(500),
    checksum VARCHAR(255),
    config_schema JSONB,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugins_container_id ON plugins(container_id);
CREATE INDEX IF NOT EXISTS idx_plugins_status ON plugins(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plugins_container_name_version ON plugins(container_id, name, version);

-- Plugin deployments table
CREATE TABLE IF NOT EXISTS plugin_deployments (
    id SERIAL PRIMARY KEY,
    plugin_id INTEGER NOT NULL,
    device_id INTEGER NOT NULL,
    version VARCHAR(50) NOT NULL,
    status VARCHAR(50) DEFAULT 'pending',
    installed_at TIMESTAMP,
    last_heartbeat TIMESTAMP,
    config JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_plugin_deployments_plugin_id ON plugin_deployments(plugin_id);
CREATE INDEX IF NOT EXISTS idx_plugin_deployments_device_id ON plugin_deployments(device_id);
CREATE INDEX IF NOT EXISTS idx_plugin_deployments_status ON plugin_deployments(status);
CREATE UNIQUE INDEX IF NOT EXISTS idx_plugin_deployments_plugin_device ON plugin_deployments(plugin_id, device_id);

-- Agent certificates table
CREATE TABLE IF NOT EXISTS agent_certificates (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL,
    certificate_serial VARCHAR(255) NOT NULL,
    certificate_pem TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    revoked_at TIMESTAMP,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_agent_certificates_device_id ON agent_certificates(device_id);
CREATE INDEX IF NOT EXISTS idx_agent_certificates_serial ON agent_certificates(certificate_serial);
CREATE INDEX IF NOT EXISTS idx_agent_certificates_status ON agent_certificates(status);

-- Verification codes table
CREATE TABLE IF NOT EXISTS verification_codes (
    id SERIAL PRIMARY KEY,
    device_id INTEGER NOT NULL,
    code_hash VARCHAR(255) NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_verification_codes_device_id ON verification_codes(device_id);
CREATE INDEX IF NOT EXISTS idx_verification_codes_code_hash ON verification_codes(code_hash);
CREATE INDEX IF NOT EXISTS idx_verification_codes_expires_at ON verification_codes(expires_at);

