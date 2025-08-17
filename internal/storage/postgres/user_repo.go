package postgres

import (
    "context"
    "database/sql"
    "errors"
    "fmt"
    "github.com/google/uuid"
    "github.com/hex-zero/MaxwellGoSpine/internal/core"
)

type UserRepo struct { db *sql.DB }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

func (r *UserRepo) Create(ctx context.Context, u *core.User) error {
    const q = `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`
    _, err := r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.CreatedAt, u.UpdatedAt)
    if err != nil { return fmt.Errorf("insert user: %w", err) }
    return nil
}

func (r *UserRepo) Get(ctx context.Context, id uuid.UUID) (*core.User, error) {
    const q = `SELECT id, name, email, created_at, updated_at FROM users WHERE id=$1`
    row := r.db.QueryRowContext(ctx, q, id)
    u := &core.User{}
    if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil {
        if errors.Is(err, sql.ErrNoRows) { return nil, core.ErrNotFound }
        return nil, fmt.Errorf("get user: %w", err)
    }
    return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *core.User) error {
    const q = `UPDATE users SET name=$2, email=$3, updated_at=$4 WHERE id=$1`
    res, err := r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.UpdatedAt)
    if err != nil { return fmt.Errorf("update user: %w", err) }
    n, _ := res.RowsAffected(); if n == 0 { return core.ErrNotFound }
    return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
    const q = `DELETE FROM users WHERE id=$1`
    res, err := r.db.ExecContext(ctx, q, id)
    if err != nil { return fmt.Errorf("delete user: %w", err) }
    n, _ := res.RowsAffected(); if n == 0 { return core.ErrNotFound }
    return nil
}

func (r *UserRepo) List(ctx context.Context, page, pageSize int) ([]*core.User, int, error) {
    if page < 1 { page = 1 }
    if pageSize <= 0 || pageSize > 100 { pageSize = 20 }
    offset := (page - 1) * pageSize
    const q = `SELECT id, name, email, created_at, updated_at FROM users ORDER BY created_at DESC LIMIT $1 OFFSET $2`
    rows, err := r.db.QueryContext(ctx, q, pageSize, offset)
    if err != nil { return nil, 0, fmt.Errorf("list users: %w", err) }
    defer rows.Close()
    var out []*core.User
    for rows.Next() {
        u := &core.User{}
        if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt); err != nil { return nil, 0, fmt.Errorf("scan user: %w", err) }
        out = append(out, u)
    }
    if err := rows.Err(); err != nil { return nil, 0, err }
    var total int
    if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM users`).Scan(&total); err != nil { return nil, 0, err }
    return out, total, nil
}
