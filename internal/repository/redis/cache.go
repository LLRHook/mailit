package redis

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/google/uuid"
	goredis "github.com/redis/go-redis/v9"
)

// Cache provides Redis-backed caching and idempotency checking.
type Cache struct {
	client *goredis.Client
}

// NewCache creates a new Cache backed by the given Redis client.
func NewCache(client *goredis.Client) *Cache {
	return &Cache{client: client}
}

// CheckIdempotency checks if an idempotency key has been used for a team.
// Returns the cached email ID if found, or an empty string if the key is new.
func (c *Cache) CheckIdempotency(ctx context.Context, teamID uuid.UUID, key string) (string, error) {
	cacheKey := fmt.Sprintf("idempotency:%s:%s", teamID.String(), key)
	val, err := c.client.Get(ctx, cacheKey).Result()
	if err == goredis.Nil {
		return "", nil
	}
	if err != nil {
		return "", fmt.Errorf("checking idempotency key: %w", err)
	}
	return val, nil
}

// SetIdempotency stores an idempotency key with the associated email ID.
// The key expires after 24 hours.
func (c *Cache) SetIdempotency(ctx context.Context, teamID uuid.UUID, key string, emailID string) error {
	cacheKey := fmt.Sprintf("idempotency:%s:%s", teamID.String(), key)
	return c.client.Set(ctx, cacheKey, emailID, 24*time.Hour).Err()
}

// CacheJSON stores a JSON-serializable value under the given key with a TTL.
func (c *Cache) CacheJSON(ctx context.Context, key string, value interface{}, ttl time.Duration) error {
	data, err := json.Marshal(value)
	if err != nil {
		return fmt.Errorf("marshalling cache value: %w", err)
	}
	return c.client.Set(ctx, key, data, ttl).Err()
}

// GetJSON retrieves and unmarshals a cached JSON value into dest.
// Returns goredis.Nil if the key does not exist.
func (c *Cache) GetJSON(ctx context.Context, key string, dest interface{}) error {
	data, err := c.client.Get(ctx, key).Bytes()
	if err != nil {
		return err
	}
	return json.Unmarshal(data, dest)
}

// Delete removes a cached value by key.
func (c *Cache) Delete(ctx context.Context, key string) error {
	return c.client.Del(ctx, key).Err()
}
