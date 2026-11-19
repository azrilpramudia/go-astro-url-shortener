package cache

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"time"

	"github.com/redis/go-redis/v9"
	"github.com/sirupsen/logrus"
)

const (
	linkKeyPrefix    = "link:"
	pendingClicksKey = "clicks:pending"
	flushingKey      = "clicks:flushing"
	flushLockKey     = "clicks:flush-lock"

	notFoundMarker = "-"
	notFoundTTL    = 60 * time.Second
)

type Lookup int

const (
	Miss Lookup = iota
	Hit
	NegativeHit
)

type LinkEntry struct {
	ID        uint       `json:"id"`
	URL       string     `json:"url"`
	ExpiresAt *time.Time `json:"expires_at,omitempty"`
}

type LinkCache struct {
	client *redis.Client
	log    *logrus.Logger
	ttl    time.Duration
}

func NewLinkCache(client *redis.Client, log *logrus.Logger, ttl time.Duration) *LinkCache {
	return &LinkCache{client: client, log: log, ttl: ttl}
}

func (c *LinkCache) Get(ctx context.Context, code string) (*LinkEntry, Lookup) {
	raw, err := c.client.Get(ctx, linkKeyPrefix+code).Result()
	if errors.Is(err, redis.Nil) {
		return nil, Miss
	}
	if err != nil {
		c.log.Warnf("cache get failed: %v", err)
		return nil, Miss
	}
	if raw == notFoundMarker {
		return nil, NegativeHit
	}

	entry := new(LinkEntry)
	if err := json.Unmarshal([]byte(raw), entry); err != nil {
		c.log.Warnf("cache entry for %s is corrupt: %v", code, err)
		return nil, Miss
	}
	return entry, Hit
}

func (c *LinkCache) Set(ctx context.Context, code string, entry *LinkEntry) {
	ttl := c.ttl
	if entry.ExpiresAt != nil {
		if remaining := time.Until(*entry.ExpiresAt); remaining < ttl {
			ttl = remaining
		}
	}
	if ttl <= 0 {
		return
	}

	data, err := json.Marshal(entry)
	if err != nil {
		c.log.Warnf("failed to encode cache entry: %v", err)
		return
	}
	if err := c.client.Set(ctx, linkKeyPrefix+code, data, ttl).Err(); err != nil {
		c.log.Warnf("cache set failed: %v", err)
	}
}

func (c *LinkCache) SetNotFound(ctx context.Context, code string) {
	if err := c.client.Set(ctx, linkKeyPrefix+code, notFoundMarker, notFoundTTL).Err(); err != nil {
		c.log.Warnf("cache set failed: %v", err)
	}
}

func (c *LinkCache) Delete(ctx context.Context, code string) {
	if err := c.client.Del(ctx, linkKeyPrefix+code).Err(); err != nil {
		c.log.Warnf("cache delete failed: %v", err)
	}
}

func (c *LinkCache) IncrementClicks(ctx context.Context, linkID uint) error {
	return c.client.HIncrBy(ctx, pendingClicksKey, strconv.FormatUint(uint64(linkID), 10), 1).Err()
}

func (c *LinkCache) AcquireFlushLock(ctx context.Context, ttl time.Duration) (bool, error) {
	return c.client.SetNX(ctx, flushLockKey, "1", ttl).Result()
}

func (c *LinkCache) ReleaseFlushLock(ctx context.Context) {
	if err := c.client.Del(ctx, flushLockKey).Err(); err != nil {
		c.log.Warnf("failed to release flush lock: %v", err)
	}
}

func (c *LinkCache) LoadClickBatch(ctx context.Context) (map[uint]int64, error) {
	exists, err := c.client.Exists(ctx, flushingKey).Result()
	if err != nil {
		return nil, err
	}
	if exists == 0 {
		if err := c.client.Rename(ctx, pendingClicksKey, flushingKey).Err(); err != nil {
			if strings.Contains(strings.ToLower(err.Error()), "no such key") {
				return nil, nil
			}
			return nil, err
		}
	}

	raw, err := c.client.HGetAll(ctx, flushingKey).Result()
	if err != nil {
		return nil, err
	}

	batch := make(map[uint]int64, len(raw))
	for field, value := range raw {
		id, err1 := strconv.ParseUint(field, 10, 64)
		n, err2 := strconv.ParseInt(value, 10, 64)
		if err1 != nil || err2 != nil || n <= 0 {
			continue
		}
		batch[uint(id)] = n
	}
	return batch, nil
}

func (c *LinkCache) CommitClickBatch(ctx context.Context) error {
	return c.client.Del(ctx, flushingKey).Err()
}
