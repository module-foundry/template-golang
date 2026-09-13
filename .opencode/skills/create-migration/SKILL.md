---
name: create-migration
description: Use when creating or changing a database migration or schema in Russian or English ("миграция", "изменить схему", "migration", "alter table", "new table"). Covers goose naming, up/down and automatic startup migrations.
---

# Create migration

## Commands

```bash
task migrate:create -- add_orders   # creates migrations/<timestamp>_add_orders.sql
task migrate:up                     # apply (goose CLI)
task migrate:down                   # roll back the last one
task migrate:status
```

At application startup pending migrations are applied automatically
(`DB_AUTO_MIGRATE=true`), guarded by a PostgreSQL session advisory lock, so
several instances are safe.

## File rules

- One logical change per migration; goose Up/Down sections:

```sql
-- +goose Up
CREATE TABLE IF NOT EXISTS orders (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id UUID NOT NULL REFERENCES users(id),
    total BIGINT NOT NULL DEFAULT 0,
    created_at TIMESTAMPTZ NOT NULL DEFAULT now()
);
CREATE INDEX IF NOT EXISTS orders_user_id_idx ON orders (user_id);

-- +goose Down
DROP TABLE IF EXISTS orders;
```

- `Down` is mandatory and must actually revert `Up`.
- Destructive changes (drop/rename column/table) require a backup note in the PR
  description and an expand/contract plan: release 1 adds the new shape, release
  2 removes the old one.
- Add indexes for foreign keys and frequent filters; name them
  `<table>_<column>_idx`.
- Never edit an applied migration: create a new one.

## After the migration

Update `docs/architecture/{en,ru}.md` if the domain model changed, and add or
update repository tests (`pgx-queries`, `tdd-tests`).
