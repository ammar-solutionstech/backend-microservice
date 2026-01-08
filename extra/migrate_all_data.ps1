# PowerShell script to migrate data from monolithic database to microservices databases
# This script exports data from the monolithic database (ITaaS) and imports to new service databases

param(
    [string]$MonolithicDB = "ITaaS",
    [string]$MonolithicHost = "localhost",
    [string]$MonolithicPort = "5432",
    [string]$MonolithicUser = "postgres"
)

Write-Host "Starting data migration from monolithic database to microservices databases..." -ForegroundColor Green

# Check if psql is available
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "Error: psql is not installed or not in PATH" -ForegroundColor Red
    Write-Host "Please install PostgreSQL client tools" -ForegroundColor Yellow
    exit 1
}

$env:PGPASSWORD = Read-Host "Enter PostgreSQL password for user $MonolithicUser" -AsSecureString | ConvertFrom-SecureString -AsPlainText

Write-Host "Migrating Inventory Service data..." -ForegroundColor Cyan
& psql -h $MonolithicHost -p $MonolithicPort -U $MonolithicUser -d $MonolithicDB -f extra/migrate_inventory_data.sql
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error migrating Inventory Service data" -ForegroundColor Red
    exit 1
}

Write-Host "Migrating Geography Service data..." -ForegroundColor Cyan
& psql -h $MonolithicHost -p $MonolithicPort -U $MonolithicUser -d $MonolithicDB -f extra/migrate_geography_data.sql
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error migrating Geography Service data" -ForegroundColor Red
    exit 1
}

Write-Host "Migrating Navigation Service data..." -ForegroundColor Cyan
& psql -h $MonolithicHost -p $MonolithicPort -U $MonolithicUser -d $MonolithicDB -f extra/migrate_navigation_data.sql
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error migrating Navigation Service data" -ForegroundColor Red
    exit 1
}

Write-Host "Data migration completed successfully!" -ForegroundColor Green
Write-Host ""
Write-Host "Note: Verify data integrity by checking record counts in each database." -ForegroundColor Yellow
Write-Host "Note: Some foreign key references (like user_id, location_id) are kept as integers" -ForegroundColor Yellow
Write-Host "      and will need to be validated through gRPC calls to respective services." -ForegroundColor Yellow

