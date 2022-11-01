package kv

import (
	"context"
	"os"

	"github.com/go-redis/redis/v8"
)

// Redis implements the http.KV interface for storage backed onto a
// Redis instance.
type Redis struct {
	db *redis.Client
}

// NewRedis returns a Redis Key/Value store.
func NewRedis() *Redis {
	return &Redis{db: redis.NewClient(&redis.Options{Addr: os.Getenv("REDIS_ADDR")})}
}

// Keys returns a list of string of keys below a certain globbed
// prefix.
func (r *Redis) Keys(ctx context.Context, prefix string) ([]string, error) {
	res := r.db.Keys(ctx, prefix)
	return res.Val(), res.Err()
}

// Get returns a single key's value.
func (r *Redis) Get(ctx context.Context, key string) ([]byte, error) {
	return r.db.Get(ctx, key).Bytes()
}

// Put stores the provided value into the designated key.
func (r *Redis) Put(ctx context.Context, key string, val []byte) error {
	return r.db.Set(ctx, key, val, 0).Err()
}

// Ping checks if the storage backend is available.
func (r *Redis) Ping(ctx context.Context) error {
	return r.db.Ping(ctx).Err()
}

// Close gracefully shuts down the connection to the redis server.
func (r *Redis) Close() error {
	return r.db.Close()
}
