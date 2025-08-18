# Database Migrations

This directory contains database migration files for the Ice Cream Store application.

## Migration Rules

- **Migration files MUST be consecutive**: Migration files must be numbered sequentially (001, 002, 003, etc.) without gaps or duplicates.
- Each migration should have both `.up.sql` and `.down.sql` files.
- Migration files should be descriptive and indicate what they do.
- Never reuse migration numbers or create gaps in the sequence.

## Migration Structure

Each migration consists of two files:
- `{number}_{description}.up.sql` - Migration to apply changes
- `{number}_{description}.down.sql` - Migration to rollback changes

## Migration Files

### 001_add_items_per_unit_to_invoice_details
- **Up**: Adds `items_per_unit` column to `invoice_details` table
- **Down**: Removes `items_per_unit` column from `invoice_details` table

### 002_update_existences_pricing_fields
- **Up**: Renames `calculated_price` to `minimum_price`, `final_price` to `maximum_price`, and adds new `final_price` field with range validation
- **Down**: Reverts pricing field changes back to original structure

## Running Migrations

### Manual Execution
```bash
# Apply migration
psql -d icecream_store -f migrations/001_add_items_per_unit_to_invoice_details.up.sql

# Rollback migration
psql -d icecream_store -f migrations/001_add_items_per_unit_to_invoice_details.down.sql
```

### Using Migration Tool (Future)
When a proper migration tool is implemented, migrations can be run with:
```bash
# Apply all pending migrations
migrate up

# Rollback last migration
migrate down

# Rollback to specific version
migrate down 001
```

## Migration Naming Convention

- Use sequential numbers (001, 002, 003, etc.)
- Use descriptive names in snake_case
- Include both .up.sql and .down.sql files
- Always test both up and down migrations

## Best Practices

1. **Always include down migration** - Every change must be reversible
2. **Test migrations** - Test both up and down on a copy of production data
3. **Use transactions** - Wrap migrations in transactions when possible
4. **Add comments** - Document what the migration does
5. **Version control** - Keep migrations in version control 