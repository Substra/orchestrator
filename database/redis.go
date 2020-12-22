// Package database provides implementations of persistence layer for different databases
package database

import "github.com/go-redis/redis/v8"

type RedisDB struct {
	redis.Client
}

func (r *RedisDB) PutState(key string, data []byte) error {
	return r.Set(ctx, key, data, 0).Err()
}

func (r *RedisDB) GetState(key string) ([]byte, error) {
	return r.Get(ctx, key).Result()
}
