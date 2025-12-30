-- Agent-related schema updates

-- Agent versions table for tracking agent versions and update manifests
CREATE TABLE IF NOT EXISTS agent_versions (
    id SERIAL PRIMARY KEY,
    version VARCHAR(50) NOT NULL,
    os_type VARCHAR(50) NOT NULL,
    architecture VARCHAR(50) NOT NULL,
    manifest JSONB,
    download_url VARCHAR(500),
    checksum VARCHAR(255),
    signature TEXT,
    release_date TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(version, os_type, architecture)
);

CREATE INDEX IF NOT EXISTS idx_agent_versions_version ON agent_versions(version);
CREATE INDEX IF NOT EXISTS idx_agent_versions_os_arch ON agent_versions(os_type, architecture);

-- Plugin registry table for tracking available plugins
CREATE TABLE IF NOT EXISTS plugin_registry (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    version VARCHAR(50) NOT NULL,
    description TEXT,
    manifest JSONB,
    download_url VARCHAR(500),
    checksum VARCHAR(255),
    signature TEXT,
    dependencies JSONB,
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    UNIQUE(name, version)
);

CREATE INDEX IF NOT EXISTS idx_plugin_registry_name ON plugin_registry(name);
CREATE INDEX IF NOT EXISTS idx_plugin_registry_status ON plugin_registry(status);

-- Extend devices table with agent metadata
ALTER TABLE devices ADD COLUMN IF NOT EXISTS agent_version VARCHAR(50);
ALTER TABLE devices ADD COLUMN IF NOT EXISTS last_health_report TIMESTAMP;
ALTER TABLE devices ADD COLUMN IF NOT EXISTS health_status VARCHAR(50) DEFAULT 'unknown';

CREATE INDEX IF NOT EXISTS idx_devices_agent_version ON devices(agent_version);
CREATE INDEX IF NOT EXISTS idx_devices_health_status ON devices(health_status);

