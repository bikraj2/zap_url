package repository

import (
	"context"
	"errors"
	"fmt"
	"math/rand"
	"time"

	"github.com/redis/go-redis/v9"
)

type Repository struct {
	client *redis.Client
}

const (
	keyLength   = 7
	batchSize   = 1000
	redisKeySet = "short_url_pool"
	lockKey     = "kgs_lock"
	lockTTL     = 10 * time.Second // Lock timeout
	lowKeyLimit = 500              // When to trigger refill
)

func New(client *redis.Client) *Repository {
	return &Repository{client}
}

func (r *Repository) GetNewKey(ctx context.Context) (string, error) {

	count, err := r.GetKeyCount(ctx)
	if err != nil {
		return "", err
	}
	if count < lowKeyLimit {
		go r.FillRedisWithKeys(ctx, batchSize)
	}

	val, err := r.client.SPop(ctx, redisKeySet).Result()
	if err != nil {
		switch {
		case errors.Is(err, redis.Nil):
			return "", ErrNoKeyleft
		default:
			return "", nil
		}
	}
	return val, err
}

func (r *Repository) acquireLock(ctx context.Context) bool {
	ok, err := r.client.SetNX(ctx, lockKey, "locked", lockTTL).Result()
	if err != nil {
		return false
	}
	return ok
}
func (r *Repository) GetKeyCount(ctx context.Context) (int64, error) {
	count, err := r.client.SCard(ctx, redisKeySet).Result()
	if err != nil {
		return 0, fmt.Errorf("error getting key count: %w", err)
	}
	return count, nil
}

func (r *Repository) FillRedisWithKeys(ctx context.Context, batchSize int) {
	if !r.acquireLock(ctx) {
		fmt.Println("Another KGS instance is already refilling Redis. Skipping...")
		return
	}
	defer r.releaseLock(ctx) // Ensure lock is released

	count := 0
	for count < batchSize {
		key := generateShortKey()
		added, _ := r.client.SAdd(ctx, redisKeySet, key).Result()
		if added > 0 {
			count++
		}
	}
	fmt.Printf("KGS instance added %d unique keys to Redis\n", batchSize)
}
func (r *Repository) releaseLock(ctx context.Context) {
	r.client.Del(ctx, lockKey)
}

func generateShortKey() string {
	const alphabet = "0123456789abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ"
	temp_rand := rand.New(rand.NewSource(time.Now().UnixNano()))
	key := make([]byte, keyLength)
	for i := range key {
		key[i] = alphabet[temp_rand.Intn(len(alphabet))]
	}
	return string(key)
}
