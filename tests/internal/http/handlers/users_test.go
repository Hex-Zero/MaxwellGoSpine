package handlers_test

import (
    "context"
    "net/http"
    "net/http/httptest"
    "strings"
    "testing"
    "github.com/go-chi/chi/v5"
    "github.com/hex-zero/MaxwellGoSpine/internal/core"
    "github.com/hex-zero/MaxwellGoSpine/internal/http/handlers"
    "github.com/google/uuid"
)

type mockUserSvc struct { users map[string]*core.User }

func (m *mockUserSvc) Create(ctx context.Context, name, email string) (*core.User, error) { u := &core.User{ID: uuid.New(), Name: name, Email: email}; if m.users==nil {m.users=map[string]*core.User{}}; m.users[email]=u; return u, nil }
func (m *mockUserSvc) Get(ctx context.Context, id uuid.UUID) (*core.User, error) { return nil, core.ErrNotFound }
func (m *mockUserSvc) Update(ctx context.Context, id uuid.UUID, name *string, email *string) (*core.User, error) { return nil, core.ErrNotFound }
func (m *mockUserSvc) Delete(ctx context.Context, id uuid.UUID) error { return core.ErrNotFound }
func (m *mockUserSvc) List(ctx context.Context, page, pageSize int) ([]*core.User, int, error) { return nil,0,nil }

func TestCreateUser(t *testing.T) {
    svc := &mockUserSvc{}
    h := handlers.NewUserHandler(svc)
    r := chi.NewRouter()
    h.Register(r)
    req := httptest.NewRequest(http.MethodPost, "/users", strings.NewReader(`{"name":"Jane","email":"jane@example.com"}`))
    req.Header.Set("Content-Type", "application/json")
    w := httptest.NewRecorder()
    r.ServeHTTP(w, req)
    if w.Code != http.StatusCreated { t.Fatalf("expected 201 got %d", w.Code) }
}
