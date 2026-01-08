-- Navigation Service Database Schema

-- Menus
CREATE TABLE IF NOT EXISTS "Menu" (
    id SERIAL PRIMARY KEY,
    name CHAR NOT NULL,
    title CHAR NOT NULL,
    uri CHAR NOT NULL
);

-- Menu Roles (many-to-many, no FK to role_id - Role is in Auth service)
CREATE TABLE IF NOT EXISTS "menu_role" (
    menu_id INTEGER NOT NULL REFERENCES "Menu"(id),
    role_id INTEGER NOT NULL,
    PRIMARY KEY (menu_id, role_id)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_menu_role_menu ON "menu_role"(menu_id);
CREATE INDEX IF NOT EXISTS idx_menu_role_role ON "menu_role"(role_id);

