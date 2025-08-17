package cache

import (
    "context"
    "time"
    rcache "github.com/dgraph-io/ristretto"
    "github.com/redis/go-redis/v9"
)

type Layered struct {
    local *rcache.Cache
    redis *redis.Client
    ttl   time.Duration
}

type Options struct {
    MaxCost int64
    NumCounters int64
    BufferItems int64
    TTL time.Duration
    RedisClient *redis.Client
}

func New(opts Options) (*Layered, error) {
    c, err := rcache.NewCache(&rcache.Config{
        NumCounters: opts.NumCounters,
        MaxCost:     opts.MaxCost,
        BufferItems: int64(opts.BufferItems),
    })
    if err != nil { return nil, err }
    return &Layered{local: c, redis: opts.RedisClient, ttl: opts.TTL}, nil
}

func (l *Layered) Get(ctx context.Context, key string) (value []byte, ok bool, err error) {
    if v, ok := l.local.Get(key); ok {
        if b, _ := v.([]byte); b != nil { return b, true, nil }
    }
    if l.redis != nil {
        res, err := l.redis.Get(ctx, key).Bytes()
        if err == nil {
            l.local.Set(key, res, int64(len(res)))
            return res, true, nil
        }
    }
    return nil, false, nil
}

func (l *Layered) Set(ctx context.Context, key string, val []byte) {
    l.local.Set(key, val, int64(len(val)))
    if l.redis != nil { _ = l.redis.Set(ctx, key, val, l.ttl).Err() }
}

func (l *Layered) Delete(ctx context.Context, key string) {
    l.local.Del(key)
    if l.redis != nil { _ = l.redis.Del(ctx, key).Err() }
}
