package resolver

import (
	"context"
	"time"

	"github.com/google/uuid"
	"github.com/hex-zero/MaxwellGoSpine/graph/model"
	"github.com/hex-zero/MaxwellGoSpine/internal/core"
)

// Query resolvers
func (q *Query) Users(ctx context.Context, page *int, pageSize *int) ([]*model.User, error) {
	p := 1
	ps := 20
	if page != nil {
		p = *page
	}
	if pageSize != nil {
		ps = *pageSize
	}
	users, _, err := q.UserService.List(ctx, p, ps)
	if err != nil {
		return nil, err
	}
	out := make([]*model.User, 0, len(users))
	for _, u := range users {
		out = append(out, convertUser(u))
	}
	return out, nil
}

func (q *Query) User(ctx context.Context, id string) (*model.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	u, err := q.UserService.Get(ctx, uid)
	if err != nil {
		return nil, err
	}
	return convertUser(u), nil
}

// Mutation resolvers
func (m *Mutation) CreateUser(ctx context.Context, name string, email string) (*model.User, error) {
	u, err := m.UserService.Create(ctx, name, email)
	if err != nil {
		return nil, err
	}
	return convertUser(u), nil
}

func (m *Mutation) UpdateUser(ctx context.Context, id string, name *string, email *string) (*model.User, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return nil, err
	}
	u, err := m.UserService.Update(ctx, uid, name, email)
	if err != nil {
		return nil, err
	}
	return convertUser(u), nil
}

func (m *Mutation) DeleteUser(ctx context.Context, id string) (bool, error) {
	uid, err := uuid.Parse(id)
	if err != nil {
		return false, err
	}
	if err := m.UserService.Delete(ctx, uid); err != nil {
		return false, err
	}
	return true, nil
}

// Helpers
func convertUser(u *core.User) *model.User {
	if u == nil {
		return nil
	}
	var del *string
	if u.DeletedAt != nil {
		s := u.DeletedAt.Format(time.RFC3339)
		del = &s
	}
	return &model.User{
		ID:        u.ID.String(),
		Name:      u.Name,
		Email:     u.Email,
		CreatedAt: u.CreatedAt.Format(time.RFC3339),
		UpdatedAt: u.UpdatedAt.Format(time.RFC3339),
		DeletedAt: del,
	}
}
