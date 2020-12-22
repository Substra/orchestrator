// Package database provides implementations of persistence layer for different databases
package database

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisDB struct {
	rdb *redis.Client
}

func NewRedisDB(rdb *redis.Client) *RedisDB {
	return &RedisDB{
		rdb: rdb,
	}
}

func (r *RedisDB) PutState(key string, data []byte) error {
	return r.rdb.Set(context.Background(), key, data, 0).Err()
}

func (r *RedisDB) GetState(key string) ([]byte, error) {
	s, err := r.rdb.Get(context.Background(), key).Result()
	return []byte(s), err
}
