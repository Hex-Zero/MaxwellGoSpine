package core

import (
	"context"
	"fmt"
	"github.com/google/uuid"
	"net/mail"
	"strings"
	"time"
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
	WithTx(ctx context.Context, fn func(context.Context, UnitOfWork) error) error
}

func NewUserService(r UserRepository) UserService { return &userService{repo: r} }

type userService struct{ repo UserRepository }

func normalizeEmail(e string) (string, error) {
	e = strings.TrimSpace(strings.ToLower(e))
	if e == "" {
		return "", fmt.Errorf("email empty: %w", ErrValidation)
	}
	if _, err := mail.ParseAddress(e); err != nil {
		return "", fmt.Errorf("invalid email: %w", ErrValidation)
	}
	return e, nil
}

func (s *userService) Create(ctx context.Context, name, email string) (*User, error) {
	if strings.TrimSpace(name) == "" {
		return nil, fmt.Errorf("name empty: %w", ErrValidation)
	}
	ne, err := normalizeEmail(email)
	if err != nil {
		return nil, err
	}
	now := time.Now().UTC()
	u := &User{ID: uuid.New(), Name: strings.TrimSpace(name), Email: ne, CreatedAt: now, UpdatedAt: now}
	if err := s.repo.Create(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	return s.repo.Get(ctx, id)
}

func (s *userService) Update(ctx context.Context, id uuid.UUID, name *string, email *string) (*User, error) {
	u, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	if name != nil {
		if strings.TrimSpace(*name) == "" {
			return nil, fmt.Errorf("name empty: %w", ErrValidation)
		}
		u.Name = strings.TrimSpace(*name)
	}
	if email != nil {
		ne, err := normalizeEmail(*email)
		if err != nil {
			return nil, err
		}
		u.Email = ne
	}
	u.UpdatedAt = time.Now().UTC()
	if err := s.repo.Update(ctx, u); err != nil {
		return nil, err
	}
	return u, nil
}

func (s *userService) Delete(ctx context.Context, id uuid.UUID) error { return s.repo.Delete(ctx, id) }

func (s *userService) List(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	return s.repo.List(ctx, page, pageSize)
}

func (s *userService) WithTx(ctx context.Context, fn func(context.Context, UnitOfWork) error) error {
	txStarter, ok := s.repo.(TxStarter)
	if !ok {
		return fmt.Errorf("transactions not supported")
	}
	uow, err := txStarter.BeginTx(ctx)
	if err != nil {
		return err
	}
	committed := false
	defer func() {
		if !committed {
			_ = uow.Rollback()
		}
	}()
	if err := fn(ctx, uow); err != nil {
		return err
	}
	if err := uow.Commit(); err != nil {
		return err
	}
	committed = true
	return nil
}

type UnitOfWork interface {
	UserRepo() UserRepository
	Commit() error
	Rollback() error
}

type TxStarter interface {
	BeginTx(ctx context.Context) (UnitOfWork, error)
}
