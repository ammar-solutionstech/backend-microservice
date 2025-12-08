-- Container Management Service Database Schema

-- Containers table
CREATE TABLE IF NOT EXISTS containers (
    id SERIAL PRIMARY KEY,
    container_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255),
    registered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    certificate_serial VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    metadata JSONB,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_containers_container_id ON containers(container_id);
CREATE INDEX IF NOT EXISTS idx_containers_status ON containers(status);

-- Bootstrap tokens table
CREATE TABLE IF NOT EXISTS bootstrap_tokens (
    id SERIAL PRIMARY KEY,
    token_hash VARCHAR(255) UNIQUE NOT NULL,
    container_id VARCHAR(255),
    expires_at TIMESTAMP NOT NULL,
    used_at TIMESTAMP,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_token_hash ON bootstrap_tokens(token_hash);
CREATE INDEX IF NOT EXISTS idx_bootstrap_tokens_container_id ON bootstrap_tokens(container_id);

-- Certificate requests table
CREATE TABLE IF NOT EXISTS certificate_requests (
    id SERIAL PRIMARY KEY,
    container_id VARCHAR(255) NOT NULL,
    csr_id INTEGER,
    request_type VARCHAR(50) NOT NULL,
    app_name VARCHAR(255),
    status VARCHAR(50) DEFAULT 'pending',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_certificate_requests_container_id ON certificate_requests(container_id);
CREATE INDEX IF NOT EXISTS idx_certificate_requests_status ON certificate_requests(status);
CREATE INDEX IF NOT EXISTS idx_certificate_requests_csr_id ON certificate_requests(csr_id);

-- Organizations table
CREATE TABLE IF NOT EXISTS organizations (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    domain VARCHAR(255) UNIQUE NOT NULL,
    admin_email VARCHAR(255) NOT NULL,
    admin_phone VARCHAR(255),
    status VARCHAR(50) DEFAULT 'active',
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_organizations_domain ON organizations(domain);
CREATE INDEX IF NOT EXISTS idx_organizations_status ON organizations(status);

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

