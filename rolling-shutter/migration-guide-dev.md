# Database Migration Guide for Keyper Implementations

This guide explains how to create and manage database migrations for any keyper
implementation in the Rolling Shutter project.

## Overview

Database migrations are automatically executed when running the `initdb` command
for all keyper implementations. This ensures that each keyper's database schema
is properly initialized and updated to the latest version.

## Migration Requirements

All migrations must adhere to the following prerequisites:

### 1. File Location

Migrations must be stored in the following directory structure:

```
keyperimpl/{keyper_name}/database/sql/migrations/
```

Where `{keyper_name}` represents the specific keyper implementation (e.g.,
`gnosis`, `optimism`, `shutterservice`, etc.).

### 2. File Format and Naming Convention

- **File Format**: All migrations must be written in SQL
- **Naming Convention**: Files must follow a strict versioning pattern:
  ```
  V{version_number}_{migration_description}.sql
  ```

#### Examples:

- `V2_validatorRegistrations.sql`
- `V3_addUserTable.sql`
- `V4_updateIndexes.sql`

#### Version Number Rules:

- **Minimum Version**: The lowest allowed version number is **2**
- **Version 1**: Represents the initial database schema (automatically created)
- **Sequential Ordering**: All subsequent migrations must have version numbers
  greater than any previously defined migration

### 3. Version Tracking

- Migration version records are maintained in the `meta_inf` table for each
  keyper
- This tracking system ensures that only new migrations are executed when the
  `initdb` command is run again
- The system automatically detects which migrations have already been applied

## Important Limitations

**No Rollback Support**: The migration system does not support rolling back to
previous versions. Once a migration is applied, it cannot be undone through the
automated migration system.
