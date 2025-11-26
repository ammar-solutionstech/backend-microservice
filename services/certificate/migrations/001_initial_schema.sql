-- Certificate Service Database Schema
-- Migration: 001_initial_schema.sql

-- Create CSR requests table
CREATE TABLE IF NOT EXISTS csr_requests (
    id SERIAL PRIMARY KEY,
    csr_pem TEXT NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'pending',
    requester_user_id INTEGER,
    requester_email VARCHAR(255),
    approved_by INTEGER,
    rejection_reason TEXT,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_csr_requests_status ON csr_requests(status);
CREATE INDEX IF NOT EXISTS idx_csr_requests_requester_user_id ON csr_requests(requester_user_id);

-- Create certificates table
CREATE TABLE IF NOT EXISTS certificates (
    id SERIAL PRIMARY KEY,
    serial_number VARCHAR(255) NOT NULL UNIQUE,
    csr_id INTEGER,
    certificate_pem TEXT NOT NULL,
    issued_at TIMESTAMP NOT NULL,
    expires_at TIMESTAMP NOT NULL,
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    step_ca_cert_id VARCHAR(255),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_certificates_serial_number ON certificates(serial_number);
CREATE INDEX IF NOT EXISTS idx_certificates_csr_id ON certificates(csr_id);
CREATE INDEX IF NOT EXISTS idx_certificates_status ON certificates(status);
CREATE INDEX IF NOT EXISTS idx_certificates_expires_at ON certificates(expires_at);

-- Create revocations table
CREATE TABLE IF NOT EXISTS revocations (
    id SERIAL PRIMARY KEY,
    certificate_id INTEGER NOT NULL,
    serial_number VARCHAR(255) NOT NULL,
    revoked_at TIMESTAMP NOT NULL,
    reason INTEGER NOT NULL DEFAULT 0,
    revoked_by INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX IF NOT EXISTS idx_revocations_certificate_id ON revocations(certificate_id);
CREATE INDEX IF NOT EXISTS idx_revocations_serial_number ON revocations(serial_number);


