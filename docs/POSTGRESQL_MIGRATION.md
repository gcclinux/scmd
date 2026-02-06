# PostgreSQL Migration Guide

## Overview
This project has been migrated from using SQLite (tardigrade.db) to PostgreSQL for better scalability and performance.

## Changes Made

### 1. New Dependencies
- `github.com/lib/pq` - PostgreSQL driver for Go
- `github.com/joho/godotenv` - Environment variable loader

### 2. New Files
- `database.go` - PostgreSQL database connection and query functions

### 3. Modified Files
- `go.mod` - Added new dependencies
- `search.go` - Updated to use PostgreSQL instead of tardigrade-mod
- `savecmd.go` - Updated to use PostgreSQL instead of tardigrade-mod
- `server.go` - Updated web interface to use PostgreSQL

## Configuration

The application reads database configuration from the `.env` file:

```env
# PostgreSQL Database Configuration
DB_HOST=192.168.1.4
DB_PORT=5432
DB_USER=user_name
DB_PASS=password
DB_NAME=database_name
TB_NAME=scmd

# Embedding Configuration (for future use)
EMBEDDING_MODEL=all-MiniLM-L6-v2
EMBEDDING_DIM=1536
```

## Database Schema

Your PostgreSQL table should have the following structure:

```sql
CREATE TABLE scmd (
    id SERIAL PRIMARY KEY,
    key TEXT NOT NULL,
    data TEXT NOT NULL
);

-- Optional: Add indexes for better search performance
CREATE INDEX idx_scmd_key ON scmd USING gin(to_tsvector('english', key));
CREATE INDEX idx_scmd_data ON scmd USING gin(to_tsvector('english', data));
```

## Usage

All existing commands work the same way:

### CLI Search
```bash
scmd.exe --search "docker,kubernetes"
```

### CLI Save
```bash
scmd.exe --save "docker ps -a" "List all containers"
```

### Web Interface
```bash
scmd.exe --web
scmd.exe --web -port 8080
scmd.exe --web -port 8080 -block
```

## Migration Notes

1. **Database Connection**: The application now connects to PostgreSQL on startup
2. **Search Functionality**: Uses PostgreSQL's ILIKE for case-insensitive pattern matching
3. **Duplicate Detection**: Checks for exact command matches before inserting
4. **Error Handling**: Improved error handling with proper logging

## Backward Compatibility

The old `tardigrade.db` file is no longer used. All data should be imported into PostgreSQL using the CLI tools in the `cli/` directory.

## Testing

1. Ensure PostgreSQL is running and accessible
2. Verify `.env` file has correct database credentials
3. Test CLI search: `scmd.exe --search "test"`
4. Test CLI save: `scmd.exe --save "test command" "test description"`
5. Test web interface: `scmd.exe --web`

## Troubleshooting

### Connection Issues
- Verify PostgreSQL is running
- Check firewall settings
- Verify database credentials in `.env`
- Test connection: `psql -h DB_HOST -p DB_PORT -U DB_USER -d DB_NAME`

### Missing Data
- Ensure data was imported from tardigrade.db to PostgreSQL
- Check table name matches TB_NAME in `.env`
- Verify table structure matches expected schema
