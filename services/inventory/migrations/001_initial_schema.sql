-- Inventory Service Database Schema
-- This schema contains all inventory-related tables

-- Brands
CREATE TABLE IF NOT EXISTS "Brand" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL
);

-- Models
CREATE TABLE IF NOT EXISTS "Model" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    brand_id INTEGER NOT NULL REFERENCES "Brand"(id)
);

-- Equipment Types
CREATE TABLE IF NOT EXISTS "Equipment_type" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    description CHAR NOT NULL
);

-- Operating Systems
CREATE TABLE IF NOT EXISTS "Operating_System" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    description CHAR NOT NULL,
    version CHAR NOT NULL,
    architectures CHAR NOT NULL
);

-- Software Categories
CREATE TABLE IF NOT EXISTS "Software_category" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL
);

-- Software
CREATE TABLE IF NOT EXISTS "Software" (
    id SERIAL PRIMARY KEY,
    software_name CHAR NOT NULL,
    category_id INTEGER NOT NULL REFERENCES "Software_category"(id),
    license_exp_date BIGINT NOT NULL
);

-- Equipment
CREATE TABLE IF NOT EXISTS "Equipment" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    description TEXT NOT NULL,
    state BOOLEAN NOT NULL,
    serial_number CHAR NOT NULL,
    equipment_type_id INTEGER NOT NULL REFERENCES "Equipment_type"(id),
    model_id INTEGER NOT NULL REFERENCES "Model"(id),
    production_date DATE NOT NULL,
    ip_v4 CHAR NOT NULL,
    ip_v6 CHAR NOT NULL,
    mac_address CHAR NOT NULL,
    operating_system_id INTEGER NOT NULL REFERENCES "Operating_System"(id),
    location_id INTEGER NOT NULL,
    proccessors CHAR NOT NULL,
    country_of_region INTEGER NOT NULL,
    user_id INTEGER NOT NULL,
    warrantly_start_date DATE NOT NULL,
    warrantly_end_date DATE NOT NULL,
    supplier_id INTEGER
);

-- Equipment Software (many-to-many)
CREATE TABLE IF NOT EXISTS "equipment_softwares" (
    software_id INTEGER NOT NULL REFERENCES "Software"(id),
    equipment_id INTEGER NOT NULL REFERENCES "Equipment"(id),
    software_version CHAR NOT NULL,
    software_size INTEGER NOT NULL,
    license CHAR NOT NULL,
    PRIMARY KEY (software_id, equipment_id)
);

-- Equipment Help Desk Links (many-to-many, no FK to help_desk_id - it's in another service)
CREATE TABLE IF NOT EXISTS "equipment_help_desk" (
    equipment_id INTEGER NOT NULL REFERENCES "Equipment"(id),
    help_desk_id INTEGER NOT NULL,
    PRIMARY KEY (equipment_id, help_desk_id)
);

-- Equipment User History
CREATE TABLE IF NOT EXISTS "Equipment_User_History" (
    equipment_id INTEGER NOT NULL REFERENCES "Equipment"(id),
    user_id INTEGER NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    comment CHAR NOT NULL,
    PRIMARY KEY (equipment_id, user_id, start_date)
);

-- Documents
CREATE TABLE IF NOT EXISTS "Documents" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    document_size CHAR NOT NULL,
    picture BYTEA NOT NULL,
    document_type CHAR NOT NULL,
    equipment_id INTEGER NOT NULL REFERENCES "Equipment"(id),
    supplier_id INTEGER NOT NULL,
    help_desk_id INTEGER NOT NULL
);

-- Maintenance
CREATE TABLE IF NOT EXISTS "Maintenance" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    start_date DATE NOT NULL,
    end_date DATE NOT NULL,
    contact_id INTEGER NOT NULL
);

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_equipment_type ON "Equipment"(equipment_type_id);
CREATE INDEX IF NOT EXISTS idx_equipment_model ON "Equipment"(model_id);
CREATE INDEX IF NOT EXISTS idx_equipment_os ON "Equipment"(operating_system_id);
CREATE INDEX IF NOT EXISTS idx_equipment_software_equipment ON "equipment_softwares"(equipment_id);
CREATE INDEX IF NOT EXISTS idx_equipment_software_software ON "equipment_softwares"(software_id);
CREATE INDEX IF NOT EXISTS idx_equipment_help_desk_equipment ON "equipment_help_desk"(equipment_id);
CREATE INDEX IF NOT EXISTS idx_equipment_user_history_equipment ON "Equipment_User_History"(equipment_id);
CREATE INDEX IF NOT EXISTS idx_documents_equipment ON "Documents"(equipment_id);

