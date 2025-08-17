package postgres

import (
    "context"
    "database/sql"
    _ "github.com/jackc/pgx/v5/stdlib"
    "time"
)

func Open(ctx context.Context, dsn string) (*sql.DB, error) {
    db, err := sql.Open("pgx", dsn)
    if err != nil { return nil, err }
    db.SetMaxOpenConns(25)
    db.SetMaxIdleConns(25)
    db.SetConnMaxLifetime(30 * time.Minute)
    ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
    defer cancel()
    if err := db.PingContext(ctx); err != nil { return nil, err }
    return db, nil
}
