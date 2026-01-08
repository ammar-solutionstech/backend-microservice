-- Data Migration Script: Navigation Service
-- Export data from monolithic database (ITaaS) and import to navigation_db
-- Run this script after the navigation_db schema has been created

-- Export and import Menus
INSERT INTO navigation_db."Menu" (
    id, name, path, icon, parent_id, order_index, is_active
)
SELECT 
    id, name, path, icon, parent_id, order_index, is_active
FROM "ITaaS".public."Menu"
ON CONFLICT (id) DO NOTHING;

-- Export and import Menu Role relationships
INSERT INTO navigation_db.menu_roles (menu_id, role_id)
SELECT menu_id, role_id
FROM "ITaaS".public.menu_roles
ON CONFLICT (menu_id, role_id) DO NOTHING;

-- Reset sequences to avoid ID conflicts
SELECT setval('navigation_db."Menu_id_seq"', (SELECT MAX(id) FROM navigation_db."Menu"));

