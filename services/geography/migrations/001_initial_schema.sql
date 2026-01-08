-- Geography Service Database Schema

-- Countries
CREATE TABLE IF NOT EXISTS "Country" (
    id SERIAL PRIMARY KEY,
    name VARCHAR(255) NOT NULL,
    code VARCHAR(50) NOT NULL
);

-- Cities
CREATE TABLE IF NOT EXISTS "City" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    country_id INTEGER NOT NULL REFERENCES "Country"(id)
);

-- Locations
CREATE TABLE IF NOT EXISTS "Location" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    zip_code CHAR,
    state CHAR,
    building_number INTEGER,
    room_number INTEGER,
    latitude NUMERIC,
    longitude NUMERIC,
    country_id INTEGER REFERENCES "Country"(id),
    city_id INTEGER REFERENCES "City"(id),
    location_map TEXT
);

-- Contacts
CREATE TABLE IF NOT EXISTS "Contact" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    country_id INTEGER NOT NULL REFERENCES "Country"(id),
    mobile_number CHAR NOT NULL,
    phone_number CHAR NOT NULL,
    website CHAR NOT NULL,
    email CHAR NOT NULL
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_city_country ON "City"(country_id);
CREATE INDEX IF NOT EXISTS idx_location_country ON "Location"(country_id);
CREATE INDEX IF NOT EXISTS idx_location_city ON "Location"(city_id);
CREATE INDEX IF NOT EXISTS idx_contact_country ON "Contact"(country_id);

