package lock

import (
	"context"
	"fmt"
	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	"time"
)

type IRedisLock interface {
	CleanUp()
	TryLock(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error)
	Unlock(ctx context.Context, lock *redislock.Lock) error
}

type RedisLock struct {
	LockHandler *redislock.Client
	client      *redis.Client
}

func NewRedisLock(client *redis.Client, lockHandler *redislock.Client) IRedisLock {
	return &RedisLock{client: client, LockHandler: lockHandler}
}

func (r RedisLock) CleanUp() {
	defer r.client.Close()
}

func (r RedisLock) TryLock(ctx context.Context, key string, ttl time.Duration, opt *redislock.Options) (*redislock.Lock, error) {
	lock, err := r.LockHandler.Obtain(ctx, key, ttl, opt)
	if err == redislock.ErrNotObtained {
		return lock, fmt.Errorf("could not obtain lock for key %s", key)
	} else if err != nil {
		return lock, err
	}

	return lock, nil
}

func (r RedisLock) Unlock(ctx context.Context, lock *redislock.Lock) error {
	err := lock.Release(ctx)
	return err
}
