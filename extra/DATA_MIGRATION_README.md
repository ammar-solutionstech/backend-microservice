# Data Migration Guide

This directory contains SQL scripts to migrate data from the monolithic database (`ITaaS`) to the new microservices databases.

## Prerequisites

1. All new service databases must be created and migrations run:
   - `inventory_db`
   - `geography_db`
   - `navigation_db`

2. PostgreSQL client tools (`psql`) must be installed

3. Access to the monolithic database (`ITaaS`)

## Migration Scripts

### Individual Service Migrations

- `migrate_inventory_data.sql` - Migrates all inventory-related data
- `migrate_geography_data.sql` - Migrates all geography-related data
- `migrate_navigation_data.sql` - Migrates all navigation-related data

### Automated Migration Scripts

- `migrate_all_data.sh` - Bash script to run all migrations (Linux/macOS)
- `migrate_all_data.ps1` - PowerShell script to run all migrations (Windows)

## Usage

### Manual Migration (SQL)

```bash
# Connect to PostgreSQL and run each script
psql -h localhost -U postgres -d ITaaS -f extra/migrate_inventory_data.sql
psql -h localhost -U postgres -d ITaaS -f extra/migrate_geography_data.sql
psql -h localhost -U postgres -d ITaaS -f extra/migrate_navigation_data.sql
```

### Automated Migration (Bash)

```bash
chmod +x extra/migrate_all_data.sh
./extra/migrate_all_data.sh
```

### Automated Migration (PowerShell)

```powershell
.\extra\migrate_all_data.ps1
```

## Important Notes

1. **Foreign Key References**: Some columns like `user_id`, `location_id`, `supplier_id` are kept as integers. These are references to other services and should be validated through gRPC calls, not database foreign keys.

2. **ID Conflicts**: The scripts use `ON CONFLICT DO NOTHING` to handle existing records. If you need to overwrite, modify the scripts accordingly.

3. **Sequence Reset**: After migration, sequences are reset to avoid ID conflicts. Verify this worked correctly.

4. **Data Verification**: After migration, verify data integrity:
   ```sql
   -- Check record counts
   SELECT 'Brand' as table_name, COUNT(*) FROM inventory_db."Brand"
   UNION ALL
   SELECT 'Model', COUNT(*) FROM inventory_db."Model"
   -- ... etc
   ```

5. **Backup First**: Always backup your databases before running migrations:
   ```bash
   pg_dump -h localhost -U postgres ITaaS > backup_monolithic.sql
   pg_dump -h localhost -U postgres inventory_db > backup_inventory.sql
   ```

## Troubleshooting

- **Connection Issues**: Ensure PostgreSQL is running and accessible
- **Permission Errors**: Ensure the database user has SELECT permissions on the monolithic database and INSERT permissions on the new databases
- **Schema Mismatches**: If table/column names differ, update the SQL scripts to match your actual schema
- **ID Conflicts**: If sequences aren't reset properly, manually reset them:
  ```sql
  SELECT setval('inventory_db."Brand_id_seq"', (SELECT MAX(id) FROM inventory_db."Brand"));
  ```

## Post-Migration Steps

1. Verify data counts match between old and new databases
2. Test API endpoints to ensure data is accessible
3. Update any application code that references old database paths
4. Consider keeping the monolithic database as a backup for a period

