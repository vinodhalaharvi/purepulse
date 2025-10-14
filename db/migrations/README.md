# Database Migrations

This directory contains SQL migration files managed by [golang-migrate](https://github.com/golang-migrate/migrate).

## Naming Convention

Migrations follow the pattern: `{version}_{description}.{up|down}.sql`

Example:
- `000001_initial_schema.up.sql` - Creates initial schema
- `000001_initial_schema.down.sql` - Rolls back initial schema

## Creating New Migrations

### Using Makefile:
```bash
make migration-create name=add_users_table
```

### Manually:
```bash
migrate create -ext sql -dir db/migrations -seq add_users_table
```

This creates:
- `000002_add_users_table.up.sql`
- `000002_add_users_table.down.sql`

## Running Migrations

### Up (apply all pending migrations):
```bash
make migrate-up
```

### Down (rollback last migration):
```bash
make migrate-down
```

### Check version:
```bash
make migrate-version
```

## Migration Guidelines

1. **Always create both up and down migrations**
2. **Test migrations thoroughly before committing**
3. **Keep migrations idempotent** (use `IF NOT EXISTS`, `IF EXISTS`)
4. **Never modify existing migrations** after they've been deployed
5. **Add comments** explaining complex schema changes
6. **Break large migrations** into smaller, incremental changes

## Migration Order

Migrations are applied in numerical order based on the version number (left-padded with zeros).


