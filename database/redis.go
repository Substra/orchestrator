// Package database provides implementations of persistence layer for different databases
package database

import (
	"context"

	"github.com/go-redis/redis/v8"
)

// RedisDB is a persistence layer relying on a redis database
type RedisDB struct {
	rdb *redis.Client
}

// NewRedisDB creates a RedisDB from a redis.Client
func NewRedisDB(rdb *redis.Client) *RedisDB {
	return &RedisDB{
		rdb: rdb,
	}
}

// PutState stores data
func (r *RedisDB) PutState(resource string, key string, data []byte) error {
	return r.rdb.HSet(context.Background(), resource, key, data).Err()
}

// GetState fetches identified data
func (r *RedisDB) GetState(resource string, key string) ([]byte, error) {
	s, err := r.rdb.HGet(context.Background(), resource, key).Result()
	return []byte(s), err
}

// GetAll retrieves all data for a resource kind
func (r *RedisDB) GetAll(resource string) (result [][]byte, err error) {
	s, err := r.rdb.HGetAll(context.Background(), resource).Result()
	for _, v := range s {
		result = append(result, []byte(v))
	}
	return
}
