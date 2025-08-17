package postgres_test

import (
    "context"
    "regexp"
    "testing"
    "time"
    "github.com/DATA-DOG/go-sqlmock"
    "github.com/hex-zero/MaxwellGoSpine/internal/storage/postgres"
    "github.com/hex-zero/MaxwellGoSpine/internal/core"
    "github.com/google/uuid"
)

func TestUserRepoCreate(t *testing.T) {
    db, mock, err := sqlmock.New()
    if err != nil { t.Fatalf("sqlmock: %v", err) }
    defer db.Close()
    repo := postgres.NewUserRepo(db)
    u := &core.User{ID: uuid.New(), Name: "John", Email: "john@example.com", CreatedAt: time.Now(), UpdatedAt: time.Now()}
    mock.ExpectExec(regexp.QuoteMeta(`INSERT INTO users (id, name, email, created_at, updated_at) VALUES ($1,$2,$3,$4,$5)`)).
        WithArgs(u.ID, u.Name, u.Email, u.CreatedAt, u.UpdatedAt).
        WillReturnResult(sqlmock.NewResult(0,1))
    if err := repo.Create(context.Background(), u); err != nil { t.Fatalf("create: %v", err) }
    if err := mock.ExpectationsWereMet(); err != nil { t.Fatalf("expect: %v", err) }
}
