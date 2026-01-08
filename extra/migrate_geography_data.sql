-- Data Migration Script: Geography Service
-- Export data from monolithic database (ITaaS) and import to geography_db
-- Run this script after the geography_db schema has been created

-- Export and import Countries
INSERT INTO geography_db."Country" (id, name, code, phone_code)
SELECT id, name, code, phone_code
FROM "ITaaS".public."Country"
ON CONFLICT (id) DO NOTHING;

-- Export and import Cities
INSERT INTO geography_db."City" (id, name, country_id)
SELECT id, name, country_id
FROM "ITaaS".public."City"
ON CONFLICT (id) DO NOTHING;

-- Export and import Locations
INSERT INTO geography_db."Location" (
    id, name, address, city_id, country_id, postal_code, latitude, longitude
)
SELECT 
    id, name, address, city_id, country_id, postal_code, latitude, longitude
FROM "ITaaS".public."Location"
ON CONFLICT (id) DO NOTHING;

-- Export and import Contacts
INSERT INTO geography_db."Contact" (
    id, first_name, last_name, email, phone, mobile, fax, location_id
)
SELECT 
    id, first_name, last_name, email, phone, mobile, fax, location_id
FROM "ITaaS".public."Contact"
ON CONFLICT (id) DO NOTHING;

-- Reset sequences to avoid ID conflicts
SELECT setval('geography_db."Country_id_seq"', (SELECT MAX(id) FROM geography_db."Country"));
SELECT setval('geography_db."City_id_seq"', (SELECT MAX(id) FROM geography_db."City"));
SELECT setval('geography_db."Location_id_seq"', (SELECT MAX(id) FROM geography_db."Location"));
SELECT setval('geography_db."Contact_id_seq"', (SELECT MAX(id) FROM geography_db."Contact"));

