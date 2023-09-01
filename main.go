package main

import (
	"context"
	"fmt"
	"race/pkg/config"
	"race/pkg/lock"
	"time"

	"github.com/bsm/redislock"
	"github.com/redis/go-redis/v9"
	log "github.com/sirupsen/logrus"
)

func main() {
	var err error

	err = config.Load("./configs/config.yml")
	if err != nil {
		log.Fatal("Failed to load config: ", err)
	}

	// redis setup
	redisClient := ConnectToRedis()
	if err != nil {
		panic(fmt.Errorf("failed to connect Redis: %v", err))
	}
	// Create a new lock client.
	lockHandler := redislock.New(redisClient)

	redisLock := lock.NewRedisLock(redisClient, lockHandler)

	ctx := context.Background()

	go Locker1(ctx, redisLock)

	go Locker2(ctx, redisLock)

	time.Sleep(5 * time.Second)

	// EOL redis setup
}

func Locker1(ctx context.Context, lock lock.IRedisLock) {
	//Retry every 100ms, for up-to 3x
	backoff := redislock.LimitRetry(redislock.LinearBackoff(100*time.Millisecond), 50)
	locker, err := lock.TryLock(ctx, "test-locker", 1*time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		log.Error("failed to lock")
	}

	log.Info("process Locker 1")
	time.Sleep(310 * time.Millisecond)
	log.Info("process Locker 1 finish")

	subLocker, err := lock.TryLock(ctx, "test-sub-locker", 1*time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		log.Error("failed to lock")
	}
	log.Info("process sub Locker 1")
	time.Sleep(310 * time.Millisecond)
	log.Info("process sub Locker 1 finish")

	lock.Unlock(ctx, subLocker)
	lock.Unlock(ctx, locker)

}

func Locker2(ctx context.Context, lock lock.IRedisLock) {
	backoff := redislock.LimitRetry(redislock.ExponentialBackoff(20*time.Millisecond, time.Millisecond), 50)

	locker, err := lock.TryLock(ctx, "test-locker", 1*time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		log.Error("failed to lock")
	}

	log.Info("process Locker 2")
	time.Sleep(310 * time.Millisecond)
	log.Info("process Locker 2 finish")

	subLocker, err := lock.TryLock(ctx, "test-sub-locker", 1*time.Minute, &redislock.Options{
		RetryStrategy: backoff,
	})
	if err != nil {
		log.Error("failed to lock")
	}
	log.Info("process sub Locker 2")
	time.Sleep(310 * time.Millisecond)
	log.Info("process sub Locker 2 finish")

	lock.Unlock(ctx, subLocker)
	lock.Unlock(ctx, locker)
}

func ConnectToRedis() *redis.Client {
	address := fmt.Sprintf("%s:%s", config.Config.REDIS.Host, config.Config.REDIS.Port)
	protocol := config.Config.REDIS.Protocol
	pass := config.Config.REDIS.Password

	opts := &redis.Options{
		Network:  protocol,
		Addr:     address,
		Password: pass,
	}

	// Connect to redis.
	return redis.NewClient(opts)

}
