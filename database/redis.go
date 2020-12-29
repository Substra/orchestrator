// Copyright 2020 Owkin Inc.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//      http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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
