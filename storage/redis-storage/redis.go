package redisstorage

import (
	"context"

	"github.com/go-redis/redis"
	"github.com/wholesome-ghoul/persona-prototype-6/config"
	"github.com/wholesome-ghoul/persona-prototype-6/logger"
)

type Redis struct {
	Key string
	redis.Client
}

func NewRedisClient(ctx context.Context, key string, config *config.Config) *Redis {
	client := &Redis{
		Key: key,
		Client: *redis.NewClient(&redis.Options{
			Addr:     config.RedisAddress,
			Password: config.RedisPassword,
			DB:       config.RedisDB,
		}).WithContext(ctx),
	}

	return client
}

func (r *Redis) NoElementsInSet(topElements []redis.Z) bool {
	return len(topElements) == 0
}

func (r *Redis) Add(members ...redis.Z) *redis.IntCmd {
	return r.ZAdd(r.Key, members...)
}

func (r *Redis) Card() *redis.IntCmd {
	return r.ZCard(r.Key)
}

func (r *Redis) TotalElems() int {
	return int(r.Card().Val())
}

func (r *Redis) PopMax() redis.Z {
	result, _ := r.ZPopMax(r.Key).Result()

	if len(result) == 0 {
		return redis.Z{}
	}

	return result[0]
}

func (r *Redis) Release() {
	code := r.Del(r.Key).Val()
	l := logger.Log()
	l.Info().Msgf("`%s` released. Code %d", r.Key, code)
}
