package core

import (
    "context"
    "fmt"
    "time"
    "github.com/google/uuid"
)

type UserRepository interface {
    Create(ctx context.Context, u *User) error
    Get(ctx context.Context, id uuid.UUID) (*User, error)
    Update(ctx context.Context, u *User) error
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, page, pageSize int) ([]*User, int, error)
}

type UserService interface {
    Create(ctx context.Context, name, email string) (*User, error)
    Get(ctx context.Context, id uuid.UUID) (*User, error)
    Update(ctx context.Context, id uuid.UUID, name *string, email *string) (*User, error)
    Delete(ctx context.Context, id uuid.UUID) error
    List(ctx context.Context, page, pageSize int) ([]*User, int, error)
}

func NewUserService(r UserRepository) UserService { return &userService{repo: r} }

type userService struct { repo UserRepository }

func (s *userService) Create(ctx context.Context, name, email string) (*User, error) {
    if name == "" || email == "" { return nil, fmt.Errorf("empty field: %w", ErrValidation) }
    u := &User{ID: uuid.New(), Name: name, Email: email, CreatedAt: time.Now().UTC(), UpdatedAt: time.Now().UTC()}
    if err := s.repo.Create(ctx, u); err != nil { return nil, err }
    return u, nil
}

func (s *userService) Get(ctx context.Context, id uuid.UUID) (*User, error) { return s.repo.Get(ctx, id) }

func (s *userService) Update(ctx context.Context, id uuid.UUID, name *string, email *string) (*User, error) {
    u, err := s.repo.Get(ctx, id)
    if err != nil { return nil, err }
    if name != nil { u.Name = *name }
    if email != nil { u.Email = *email }
    u.UpdatedAt = time.Now().UTC()
    if err := s.repo.Update(ctx, u); err != nil { return nil, err }
    return u, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error { return s.repo.Delete(ctx, id) }

func (s *userService) List(ctx context.Context, page, pageSize int) ([]*User, int, error) { return s.repo.List(ctx, page, pageSize) }
