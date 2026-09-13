---
name: pgx-queries
description: Use when writing or changing a database query, repository method or row mapping in Russian or English ("запрос к БД", "repository", "метод репозитория", "SQL", "pgx"). Enforces db-tag mapping via pgx.RowToStructByName and transaction rules.
---

# pgx queries

## Mapping contract

Struct tags map columns; manual scanning is forbidden:

```go
type Order struct {
    ID        string    `db:"id"`
    UserID    string    `db:"user_id"`
    Total     int64     `db:"total"`
    CreatedAt time.Time `db:"created_at"`
}

func (r *OrderRepository) List(ctx context.Context, userID string) ([]Order, error) {
    rows, err := r.pool.Query(ctx,
        `select id, user_id, total, created_at from orders where user_id = $1`,
        userID,
    )
    if err != nil {
        return nil, fmt.Errorf("list orders: %w", err)
    }
    orders, err := pgx.CollectRows(rows, pgx.RowToStructByName[Order])
    if err != nil {
        return nil, fmt.Errorf("collect orders: %w", err)
    }
    return orders, nil
}
```

- Optional columns (may be NULL or absent) -> `pgx.RowToStructByNameLax[T]`.
- Aggregate/scalar queries may use `QueryRow(ctx, sql, args...).Scan(&x)`.
- Never `select *`: list columns explicitly in the same order-independent way,
  the mapper is name-based.

## Rules

- Parameterized queries only (`$1`, `$2`); string concatenation is forbidden.
- `context.Context` is the first argument; wrap errors with `%w`.
- Transactions use `postgres.WithTx(ctx, pool, func(tx pgx.Tx) error {...})`.
- Repository methods return domain values and wrap infrastructure failures with
  `apperror.Runtime(apperror.CodeDBQueryFailed, err)` (or `CodeNotFound` for
  `pgx.ErrNoRows` when it is an expected case).
- Batch loops: prepare once, reuse structures, avoid `map[string]any`.
- Every new query gets a unit test for logic and an `integration`-tagged test
  for the SQL itself.

## Schema changes

Schema changes go through `create-migration`; repository code never creates or
alters tables.
