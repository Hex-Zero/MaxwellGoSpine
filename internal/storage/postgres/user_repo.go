package postgres

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"github.com/google/uuid"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"time"
)

type UserRepo struct{ db *sql.DB }
type userTxRepo struct{ tx *sql.Tx }

func NewUserRepo(db *sql.DB) *UserRepo { return &UserRepo{db: db} }

// BeginTx implements core.TxStarter when asserted by service layer.
func (r *UserRepo) BeginTx(ctx context.Context) (core.UnitOfWork, error) {
	tx, err := r.db.BeginTx(ctx, nil)
	if err != nil {
		return nil, fmt.Errorf("begin tx: %w", err)
	}
	return &userTxRepo{tx: tx}, nil
}

// UnitOfWork interface implementation
func (t *userTxRepo) UserRepo() core.UserRepository { return &txUserRepo{tx: t.tx} }
func (t *userTxRepo) Commit() error                 { return t.tx.Commit() }
func (t *userTxRepo) Rollback() error               { return t.tx.Rollback() }

// txUserRepo delegates methods to transactional *sql.Tx
type txUserRepo struct{ tx *sql.Tx }

func (r *txUserRepo) Create(ctx context.Context, u *core.User) error {
	const q = `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`
	if _, err := r.tx.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.CreatedAt, u.UpdatedAt); err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}
func (r *txUserRepo) Get(ctx context.Context, id uuid.UUID) (*core.User, error) {
	const q = `SELECT id, name, email, created_at, updated_at, deleted_at FROM users WHERE id=$1 AND deleted_at IS NULL`
	row := r.tx.QueryRowContext(ctx, q, id)
	u := &core.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}
func (r *txUserRepo) Update(ctx context.Context, u *core.User) error {
	const q = `UPDATE users SET name=$2, email=$3, updated_at=$4 WHERE id=$1 AND deleted_at IS NULL`
	res, err := r.tx.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}
func (r *txUserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at=$2 WHERE id=$1 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := r.tx.ExecContext(ctx, q, id, now)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}
func (r *txUserRepo) List(ctx context.Context, page, pageSize int) ([]*core.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	const q = `SELECT id, name, email, created_at, updated_at, deleted_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.tx.QueryContext(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var out []*core.User
	for rows.Next() {
		u := &core.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.tx.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

func (r *UserRepo) Create(ctx context.Context, u *core.User) error {
	const q = `INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`
	_, err := r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.CreatedAt, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("insert user: %w", err)
	}
	return nil
}

func (r *UserRepo) Get(ctx context.Context, id uuid.UUID) (*core.User, error) {
	const q = `SELECT id, name, email, created_at, updated_at, deleted_at FROM users WHERE id=$1 AND deleted_at IS NULL`
	row := r.db.QueryRowContext(ctx, q, id)
	u := &core.User{}
	if err := row.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
		if errors.Is(err, sql.ErrNoRows) {
			return nil, core.ErrNotFound
		}
		return nil, fmt.Errorf("get user: %w", err)
	}
	return u, nil
}

func (r *UserRepo) Update(ctx context.Context, u *core.User) error {
	const q = `UPDATE users SET name=$2, email=$3, updated_at=$4 WHERE id=$1 AND deleted_at IS NULL`
	res, err := r.db.ExecContext(ctx, q, u.ID, u.Name, u.Email, u.UpdatedAt)
	if err != nil {
		return fmt.Errorf("update user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *UserRepo) Delete(ctx context.Context, id uuid.UUID) error {
	const q = `UPDATE users SET deleted_at=$2 WHERE id=$1 AND deleted_at IS NULL`
	now := time.Now().UTC()
	res, err := r.db.ExecContext(ctx, q, id, now)
	if err != nil {
		return fmt.Errorf("soft delete user: %w", err)
	}
	n, _ := res.RowsAffected()
	if n == 0 {
		return core.ErrNotFound
	}
	return nil
}

func (r *UserRepo) List(ctx context.Context, page, pageSize int) ([]*core.User, int, error) {
	if page < 1 {
		page = 1
	}
	if pageSize <= 0 || pageSize > 100 {
		pageSize = 20
	}
	offset := (page - 1) * pageSize
	const q = `SELECT id, name, email, created_at, updated_at, deleted_at FROM users WHERE deleted_at IS NULL ORDER BY created_at DESC LIMIT $1 OFFSET $2`
	rows, err := r.db.QueryContext(ctx, q, pageSize, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("list users: %w", err)
	}
	defer rows.Close()
	var out []*core.User
	for rows.Next() {
		u := &core.User{}
		if err := rows.Scan(&u.ID, &u.Name, &u.Email, &u.CreatedAt, &u.UpdatedAt, &u.DeletedAt); err != nil {
			return nil, 0, fmt.Errorf("scan user: %w", err)
		}
		out = append(out, u)
	}
	if err := rows.Err(); err != nil {
		return nil, 0, err
	}
	var total int
	if err := r.db.QueryRowContext(ctx, `SELECT count(*) FROM users WHERE deleted_at IS NULL`).Scan(&total); err != nil {
		return nil, 0, err
	}
	return out, total, nil
}

// Existing root-level repository methods remain (non-transactional) below.
