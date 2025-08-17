package core

import (
    "context"
    "errors"
    "sort"
    "sync"
    "time"
    "github.com/google/uuid"
)

// InMemoryUserRepo is a concurrency-safe in-memory implementation of UserRepository for local dev/testing without Postgres.
type InMemoryUserRepo struct {
    mu    sync.RWMutex
    users map[uuid.UUID]*User
}

func NewInMemoryUserRepo() *InMemoryUserRepo {
    return &InMemoryUserRepo{users: make(map[uuid.UUID]*User)}
}

func (r *InMemoryUserRepo) Create(_ context.Context, u *User) error {
    r.mu.Lock(); defer r.mu.Unlock()
    if _, exists := r.users[u.ID]; exists {
        return errors.New("duplicate id")
    }
    // copy to avoid external mutation
    cpy := *u
    r.users[u.ID] = &cpy
    return nil
}

func (r *InMemoryUserRepo) Get(_ context.Context, id uuid.UUID) (*User, error) {
    r.mu.RLock(); defer r.mu.RUnlock()
    u, ok := r.users[id]
    if !ok || u.DeletedAt != nil {
        return nil, ErrNotFound
    }
    cpy := *u
    return &cpy, nil
}

func (r *InMemoryUserRepo) Update(_ context.Context, u *User) error {
    r.mu.Lock(); defer r.mu.Unlock()
    existing, ok := r.users[u.ID]
    if !ok || existing.DeletedAt != nil {
        return ErrNotFound
    }
    cpy := *u
    r.users[u.ID] = &cpy
    return nil
}

func (r *InMemoryUserRepo) Delete(_ context.Context, id uuid.UUID) error {
    r.mu.Lock(); defer r.mu.Unlock()
    u, ok := r.users[id]
    if !ok || u.DeletedAt != nil {
        return ErrNotFound
    }
    now := time.Now().UTC()
    u.DeletedAt = &now
    return nil
}

func (r *InMemoryUserRepo) List(_ context.Context, page, pageSize int) ([]*User, int, error) {
    r.mu.RLock(); defer r.mu.RUnlock()
    var list []*User
    for _, u := range r.users {
        if u.DeletedAt == nil {
            cpy := *u
            list = append(list, &cpy)
        }
    }
    sort.Slice(list, func(i, j int) bool { return list[i].CreatedAt.After(list[j].CreatedAt) })
    total := len(list)
    if pageSize <= 0 { pageSize = 50 }
    if page <= 0 { page = 1 }
    start := (page - 1) * pageSize
    if start >= total { return []*User{}, total, nil }
    end := start + pageSize
    if end > total { end = total }
    return list[start:end], total, nil
}

// Ensure interface compliance
var _ UserRepository = (*InMemoryUserRepo)(nil)
