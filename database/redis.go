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

func (r *RedisDB) PutState(resource string, key string, data []byte) error {
	return r.rdb.HSet(context.Background(), resource, key, data).Err()
}

func (r *RedisDB) GetState(resource string, key string) ([]byte, error) {
	s, err := r.rdb.HGet(context.Background(), resource, key).Result()
	return []byte(s), err
}

func (r *RedisDB) GetAll(resource string) (result [][]byte, err error) {
	s, err := r.rdb.HGetAll(context.Background(), resource).Result()
	for _, v := range s {
		result = append(result, []byte(v))
	}
	return
}
