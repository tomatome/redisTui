package main

import (
	"context"

	"github.com/go-redis/redis/v8"
)

type RedisClient struct {
	*redis.Client
}

func newRedisClient(addr, password string, db int) *RedisClient {
	return &RedisClient{
		Client: redis.NewClient(&redis.Options{
			Addr:     addr,
			Password: password,
			DB:       db,
		}),
	}
}
func (rc *RedisClient) TestPing() (string, error) {
	return rc.Ping(context.Background()).Result()
}

// 获取 Redis 中所有的键
func (rc *RedisClient) GetAllKeys() ([]string, error) {
	var cursor uint64
	var keys []string
	for {
		var err error
		var scanKeys []string
		scanKeys, cursor, err = rc.Scan(context.Background(), cursor, "*", 100).Result()
		if err != nil {
			return nil, err
		}
		keys = append(keys, scanKeys...)
		if cursor == 0 {
			break
		}
	}
	return keys, nil
}

// 获取每个键的类型
func (rc *RedisClient) GetType(key string) (string, error) {
	return rc.Type(context.Background(), key).Result()
}

func (rc *RedisClient) GetValues(keys []string) ([]interface{}, error) {
	var values []interface{}
	for _, k := range keys {
		val, err := rc.Get(context.Background(), k).Result()
		if err != nil {
			return nil, err
		}
		values = append(values, val)
	}
	return values, nil
}

// 获取哈希表所有键
func (rc *RedisClient) GetAllHashKeys(key string) ([]string, error) {
	return rc.HKeys(context.Background(), key).Result()
}

// 获取哈希表所有值
func (rc *RedisClient) GetAllHashValues(key string) (map[string]string, error) {
	return rc.HGetAll(context.Background(), key).Result()
}

// 获取字符串类型的全部值
func (rc *RedisClient) GetAllStrings() ([]string, error) {
	var cursor uint64
	var values []string
	for {
		var err error
		var scanValues []string
		scanValues, cursor, err = rc.Scan(context.Background(), cursor, "*", 100).Result()
		if err != nil {
			return nil, err
		}
		values = append(values, scanValues...)
		if cursor == 0 {
			break
		}
	}
	return values, nil
}

// 获取列表类型的全部值
func (rc *RedisClient) GetAllLists(key string) ([]string, error) {
	return rc.LRange(context.Background(), key, 0, -1).Result()
}

// 获取集合类型的全部值
func (rc *RedisClient) GetAllSets(key string) ([]string, error) {
	return rc.SMembers(context.Background(), key).Result()
}

// 获取有序集合类型的全部值
func (rc *RedisClient) GetAllSortedSets(key string) ([]string, error) {
	var values []string
	var cursor uint64
	for {
		var err error
		var scanValues []string
		scanValues, cursor, err = rc.ZScan(context.Background(), key, cursor, "*", 100).Result()
		if err != nil {
			return nil, err
		}
		values = append(values, scanValues...)
		if cursor == 0 {
			break
		}
	}
	return values, nil
}

// 设置hash类型的值
func (rc *RedisClient) SetHashKeyVal(key, field, value string) error {
	return rc.HSet(context.Background(), key, field, value).Err()
}
func (rc *RedisClient) SetStringKeyVal(key, value string) error {
	return rc.Set(context.Background(), key, value, 0).Err()
}
