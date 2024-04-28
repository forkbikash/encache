package encache

import (
	"context"
	"errors"
	"sync"
	"time"

	"github.com/redis/go-redis/v9"
)

type MuLockImpl struct {
	mu sync.Mutex
}

func NewMuLockImpl() *MuLockImpl {
	return &MuLockImpl{}
}

func (lockImpl *MuLockImpl) lock(_ ...string) error {
	lockImpl.mu.Lock()
	return nil
}

func (lockImpl *MuLockImpl) unlock(_ ...string) error {
	lockImpl.mu.Unlock()
	return nil
}

type RedisLockImpl struct {
	client      redis.UniversalClient
	lockTimeout time.Duration
}

func NewRedisLockImpl(client redis.UniversalClient, lockTimeout time.Duration) *RedisLockImpl {
	return &RedisLockImpl{
		client:      client,
		lockTimeout: lockTimeout,
	}
}

func (lockImpl *RedisLockImpl) lock(key ...string) error {
	ctx := context.Background()
	result, err := lockImpl.client.SetNX(ctx, lockImpl.getLockKey(key[0]), "1", lockImpl.lockTimeout).Result()
	if err != nil {
		return err
	}

	if !result {
		return errors.New("SetNX returned false")
	}

	return nil
}

func (lockImpl *RedisLockImpl) unlock(key ...string) error {
	ctx := context.Background()
	_, err := lockImpl.client.Del(ctx, lockImpl.getLockKey(key[0])).Result()
	return err
}

func (lockImpl *RedisLockImpl) getLockKey(key string) string {
	return "lock_cache_func_" + key
}

type NoLockImpl struct{}

func NewNoLockImpl() *NoLockImpl {
	return &NoLockImpl{}
}

func (lockImpl *NoLockImpl) lock(_ ...string) error {
	return nil
}

func (lockImpl *NoLockImpl) unlock(_ ...string) error {
	return nil
}
