# Quick Start Guide - PostgreSQL Version

## Prerequisites

1. PostgreSQL server running and accessible
2. Database and table created with imported data
3. `.env` file configured with your database credentials

## Step 1: Verify Configuration

Check your `.env` file:

```bash
type .env
```

Should contain:
```env
DB_HOST=192.168.1.4
DB_PORT=5432
DB_USER=user_name
DB_PASS=password
DB_NAME=database_name
TB_NAME=scmd
```

## Step 2: Test Connection (Optional)

```bash
go run test_connection.go database.go
```

Expected output:
```
Testing PostgreSQL connection...

Successfully connected to PostgreSQL!

Search query executed successfully (found X bytes of results)

All tests passed!
```

## Step 3: Try CLI Commands

### Search for commands:
```bash
scmd.exe --search "docker"
```

### Add a new command:
```bash
scmd.exe --save "docker ps -a" "List all Docker containers"
```

## Step 4: Start Web Interface

### Basic web server:
```bash
scmd.exe --web
```

### Custom port:
```bash
scmd.exe --web -port 8080
```

### Read-only mode (disable adding commands):
```bash
scmd.exe --web -block
```

### Background service (no browser):
```bash
scmd.exe --web -port 3333 -service
```

## Common Issues

### Connection Failed
```
Failed to connect to database: dial tcp: connect: connection refused
```

**Solution:** Check if PostgreSQL is running and accessible:
```bash
psql -h 192.168.1.4 -p 5432 -U user_name -d database_name
```

### Table Not Found
```
Error querying database: relation "scmd" does not exist
```

**Solution:** Verify table name in `.env` matches your PostgreSQL table:
```sql
\dt
```

### No Results
```
Database empty
```

**Solution:** Import data from tardigrade.db using the CLI tools:
```bash
cd cli
python import_to_postgres.py
```

## Web Interface URLs

After starting the web server, access it at:

- Local: `http://localhost:3333`
- Network: `http://YOUR_IP:3333`

The application will automatically open your browser to the correct URL.

## Next Steps

1. ✓ Test CLI search
2. ✓ Test CLI save
3. ✓ Test web interface
4. ✓ Add your own commands
5. ✓ Share with your team!

## Tips

- Use comma-separated patterns for multiple searches: `--search "docker,kubernetes"`
- The web interface supports the same search patterns
- All commands are stored in PostgreSQL for easy backup and sharing
- Use `-block` flag to prevent users from adding commands via web UI

## Need Help?

- See `POSTGRESQL_MIGRATION.md` for detailed migration info
- See `UPGRADE_SUMMARY.md` for complete change list
- Check PostgreSQL logs for database errors
- Review `.env.example` for configuration reference
