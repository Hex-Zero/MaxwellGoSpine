package core

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/google/uuid"
	"github.com/hex-zero/MaxwellGoSpine/internal/cache"
	"sync/atomic"
	"time"
)

// cachedUserService decorates UserService with layered cache and invalidation.
type cachedUserService struct {
	base    UserService
	cache   *cache.Layered
	listVer atomic.Uint64
	ttl     time.Duration
}

func NewCachedUserService(base UserService, c *cache.Layered) UserService {
	if c == nil {
		return base
	}
	s := &cachedUserService{base: base, cache: c, ttl: 5 * time.Minute}
	return s
}

func (s *cachedUserService) cacheKeyUser(id uuid.UUID) string { return "user:get:" + id.String() }
func (s *cachedUserService) cacheKeyList(page, size int) string {
	return fmt.Sprintf("user:list:v%d:%d:%d", s.listVer.Load(), page, size)
}

// Create invalidates list caches by version bump and caches new user.
func (s *cachedUserService) Create(ctx context.Context, name, email string) (*User, error) {
	u, err := s.base.Create(ctx, name, email)
	if err != nil {
		return nil, err
	}
	s.listVer.Add(1)
	s.setUser(ctx, u)
	return u, nil
}

func (s *cachedUserService) Get(ctx context.Context, id uuid.UUID) (*User, error) {
	if u := s.getUser(ctx, id); u != nil {
		return u, nil
	}
	u, err := s.base.Get(ctx, id)
	if err != nil {
		return nil, err
	}
	s.setUser(ctx, u)
	return u, nil
}

func (s *cachedUserService) Update(ctx context.Context, id uuid.UUID, name *string, email *string) (*User, error) {
	u, err := s.base.Update(ctx, id, name, email)
	if err != nil {
		return nil, err
	}
	s.delUser(ctx, id)
	s.listVer.Add(1)
	s.setUser(ctx, u)
	return u, nil
}

func (s *cachedUserService) Delete(ctx context.Context, id uuid.UUID) error {
	if err := s.base.Delete(ctx, id); err != nil {
		return err
	}
	s.delUser(ctx, id)
	s.listVer.Add(1)
	return nil
}

func (s *cachedUserService) List(ctx context.Context, page, pageSize int) ([]*User, int, error) {
	key := s.cacheKeyList(page, pageSize)
	if b, ok, _ := s.cache.Get(ctx, key); ok {
		var wrap struct {
			Users []*User `json:"u"`
			Total int     `json:"t"`
		}
		if err := json.Unmarshal(b, &wrap); err == nil {
			return wrap.Users, wrap.Total, nil
		}
	}
	users, total, err := s.base.List(ctx, page, pageSize)
	if err != nil {
		return nil, 0, err
	}
	payload, _ := json.Marshal(struct {
		Users []*User `json:"u"`
		Total int     `json:"t"`
	}{users, total})
	s.cache.Set(ctx, key, payload)
	return users, total, nil
}

func (s *cachedUserService) WithTx(ctx context.Context, fn func(context.Context, UnitOfWork) error) error {
	return s.base.WithTx(ctx, fn)
}

// helpers
func (s *cachedUserService) getUser(ctx context.Context, id uuid.UUID) *User {
	key := s.cacheKeyUser(id)
	if b, ok, _ := s.cache.Get(ctx, key); ok {
		var u User
		if json.Unmarshal(b, &u) == nil {
			return &u
		}
	}
	return nil
}
func (s *cachedUserService) setUser(ctx context.Context, u *User) {
	b, _ := json.Marshal(u)
	s.cache.Set(ctx, s.cacheKeyUser(u.ID), b)
}
func (s *cachedUserService) delUser(ctx context.Context, id uuid.UUID) {
	s.cache.Delete(ctx, s.cacheKeyUser(id))
}
