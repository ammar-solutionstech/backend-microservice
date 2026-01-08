-- Data Migration Script: Inventory Service
-- Export data from monolithic database (ITaaS) and import to inventory_db
-- Run this script after the inventory_db schema has been created

-- Note: Adjust table names and column names based on actual monolithic database structure
-- This script assumes the monolithic database is named "ITaaS" and uses the public schema

-- Export and import Brands
INSERT INTO inventory_db."Brand" (id, name)
SELECT id, name
FROM "ITaaS".public."Brand"
ON CONFLICT (id) DO NOTHING;

-- Export and import Models
INSERT INTO inventory_db."Model" (id, name, brand_id)
SELECT id, name, brand_id
FROM "ITaaS".public."Model"
ON CONFLICT (id) DO NOTHING;

-- Export and import Equipment Types
INSERT INTO inventory_db."Equipment_type" (id, name, description)
SELECT id, name, description
FROM "ITaaS".public."Equipment_type"
ON CONFLICT (id) DO NOTHING;

-- Export and import Operating Systems
INSERT INTO inventory_db."Operating_System" (id, name, description, version, architectures)
SELECT id, name, description, version, architectures
FROM "ITaaS".public."Operating_System"
ON CONFLICT (id) DO NOTHING;

-- Export and import Software Categories
INSERT INTO inventory_db."Software_category" (id, name)
SELECT id, name
FROM "ITaaS".public."Software_category"
ON CONFLICT (id) DO NOTHING;

-- Export and import Software
INSERT INTO inventory_db."Software" (id, software_name, category_id, license_exp_date)
SELECT id, software_name, category_id, license_exp_date
FROM "ITaaS".public."Software"
ON CONFLICT (id) DO NOTHING;

-- Export and import Equipment
-- Note: location_id, user_id, supplier_id are kept as integers (references to other services)
INSERT INTO inventory_db."Equipment" (
    id, name, description, state, serial_number, equipment_type_id, model_id,
    production_date, ip_v4, ip_v6, mac_address, operating_system_id,
    location_id, proccessors, country_of_region, user_id,
    warrantly_start_date, warrantly_end_date, supplier_id
)
SELECT 
    id, name, description, state, serial_number, equipment_type_id, model_id,
    production_date, ip_v4, ip_v6, mac_address, operating_system_id,
    location_id, proccessors, country_of_region, user_id,
    warrantly_start_date, warrantly_end_date, supplier_id
FROM "ITaaS".public."Equipment"
ON CONFLICT (id) DO NOTHING;

-- Export and import Equipment Software relationships
INSERT INTO inventory_db.equipment_softwares (equipment_id, software_id, software_version, software_size, license)
SELECT equipment_id, software_id, software_version, software_size, license
FROM "ITaaS".public.equipment_softwares
ON CONFLICT (equipment_id, software_id) DO NOTHING;

-- Export and import Equipment Help Desk relationships
INSERT INTO inventory_db.equipment_help_desk (equipment_id, help_desk_id)
SELECT equipment_id, help_desk_id
FROM "ITaaS".public.equipment_help_desk
ON CONFLICT (equipment_id, help_desk_id) DO NOTHING;

-- Export and import Equipment User History
INSERT INTO inventory_db.equipment_user_history (equipment_id, user_id, start_date, end_date, comment)
SELECT equipment_id, user_id, start_date, end_date, comment
FROM "ITaaS".public.equipment_user_history
ON CONFLICT (equipment_id, user_id, start_date) DO NOTHING;

-- Export and import Documents (that are related to inventory/equipment)
INSERT INTO inventory_db."Document" (
    id, name, document_size, picture, document_type, help_desk_id, equipment_id, supplier_id
)
SELECT id, name, document_size, picture, document_type, help_desk_id, equipment_id, supplier_id
FROM "ITaaS".public."Document"
WHERE equipment_id IS NOT NULL OR help_desk_id IS NOT NULL
ON CONFLICT (id) DO NOTHING;

-- Export and import Maintenance records
INSERT INTO inventory_db."Maintenance" (
    id, equipment_id, maintenance_date, maintenance_type, description, cost, technician_id
)
SELECT id, equipment_id, maintenance_date, maintenance_type, description, cost, technician_id
FROM "ITaaS".public."Maintenance"
ON CONFLICT (id) DO NOTHING;

-- Reset sequences to avoid ID conflicts
SELECT setval('inventory_db."Brand_id_seq"', (SELECT MAX(id) FROM inventory_db."Brand"));
SELECT setval('inventory_db."Model_id_seq"', (SELECT MAX(id) FROM inventory_db."Model"));
SELECT setval('inventory_db."Equipment_type_id_seq"', (SELECT MAX(id) FROM inventory_db."Equipment_type"));
SELECT setval('inventory_db."Operating_System_id_seq"', (SELECT MAX(id) FROM inventory_db."Operating_System"));
SELECT setval('inventory_db."Software_category_id_seq"', (SELECT MAX(id) FROM inventory_db."Software_category"));
SELECT setval('inventory_db."Software_id_seq"', (SELECT MAX(id) FROM inventory_db."Software"));
SELECT setval('inventory_db."Equipment_id_seq"', (SELECT MAX(id) FROM inventory_db."Equipment"));
SELECT setval('inventory_db."Document_id_seq"', (SELECT MAX(id) FROM inventory_db."Document"));
SELECT setval('inventory_db."Maintenance_id_seq"', (SELECT MAX(id) FROM inventory_db."Maintenance"));

