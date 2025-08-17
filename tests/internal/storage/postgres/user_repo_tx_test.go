package postgres_test

import (
	"context"
	"github.com/DATA-DOG/go-sqlmock"
	"github.com/google/uuid"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
	"github.com/hex-zero/MaxwellGoSpine/internal/storage/postgres"
	"regexp"
	"testing"
	"time"
)

func TestUserRepoTxCreate(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := postgres.NewUserRepo(db)
	ctx := context.Background()
	mock.ExpectBegin()
	uow, err := repo.BeginTx(ctx)
	if err != nil {
		t.Fatalf("begin: %v", err)
	}
	txRepo := uow.UserRepo()
	u := &core.User{ID: uuid.New(), Name: "Tx", Email: "tx@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
	mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`)).
		WithArgs(u.ID, u.Name, u.Email, u.CreatedAt, u.UpdatedAt).WillReturnResult(sqlmock.NewResult(0, 1))
	mock.ExpectCommit()
	if err := txRepo.Create(ctx, u); err != nil {
		t.Fatalf("create tx: %v", err)
	}
	if err := uow.Commit(); err != nil {
		t.Fatalf("commit: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}

func TestUserRepoSoftDelete(t *testing.T) {
	db, mock, err := sqlmock.New()
	if err != nil {
		t.Fatalf("sqlmock: %v", err)
	}
	defer db.Close()
	repo := postgres.NewUserRepo(db)
	id := uuid.New()
	mock.ExpectExec(regexp.QuoteMeta(`UPDATE users SET deleted_at=$2 WHERE id=$1 AND deleted_at IS NULL`)).
		WithArgs(id, sqlmock.AnyArg()).WillReturnResult(sqlmock.NewResult(0, 1))
	if err := repo.Delete(context.Background(), id); err != nil {
		t.Fatalf("soft delete: %v", err)
	}
	if err := mock.ExpectationsWereMet(); err != nil {
		t.Fatalf("expect: %v", err)
	}
}
