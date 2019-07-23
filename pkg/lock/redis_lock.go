package lock

import (
	"github.com/go-redis/redis"
	"time"
)

type RedisLock struct {
	redisClient *redis.Client
}

func (r *RedisLock) InitRedisLock(addr string) {
	r.redisClient = redis.NewClient(&redis.Options{
		Addr: addr,
	})
}

func (r *RedisLock) LockGetFlag(key string) (error, bool) {
	value := "lock"
	bLock, err := r.redisClient.SetNX(key, value, time.Second*2).Result()
	if err != nil {
		return err, true
	}
	if bLock {
		return nil, true
	}
	return nil, false
}

func (r *RedisLock) UnLockGetFlag(key string) error {
	nDel, err := r.redisClient.Del(key).Result()
	if err != nil {
		return err
	}
	if nDel != 1 {
		return nil
	}
	return nil
}
