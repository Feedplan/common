package cache

import (
	"context"
	"os"
	"sync"
	"time"

	"github.com/go-redis/redis/v8"
	"github.com/spf13/viper"

	"gitlab.com/feedplan-libraries/common/constants"
	"gitlab.com/feedplan-libraries/common/logger"
)

var once sync.Once
var redisClient *RedisClientImp

type IRedisClient interface {
	Get(key string) (string, error)
	Set(key string, value interface{}, ttl time.Duration) (string, error)
	Del(keys ...string) (int64, error)
}

type RedisClientImp struct {
	RedisClient *redis.Client
}

// GetRedisClientImp : Returns new redis client after initializing and validating the connection to the redis distributed cache
func GetRedisClientImp() IRedisClient {
	once.Do(func() {
		redisURL := viper.GetString(constants.RedisURLKey)
		redisUser := viper.GetString(constants.RedisUserKey)
		redisPassword := viper.GetString(constants.RedisPasswordKey)
		client := redis.NewClient(&redis.Options{
			Addr:     redisURL,
			DB:       0, // use default DB
			Password: redisPassword,
			Username: redisUser,
		})

		pingResponse, err := client.Ping(context.Background()).Result()
		if err != nil {
			logger.SugarLogger.Errorf("Error while pinging redis cluster. ErrorMessage: %s", err.Error())
			os.Exit(1)
		}
		logger.SugarLogger.Infof("Pinged redis server. Response: %s", pingResponse)
		redisClient = &RedisClientImp{RedisClient: client}
	})
	return redisClient
}

func (u RedisClientImp) Get(key string) (string, error) {
	return u.RedisClient.Get(context.Background(), key).Result()
}

func (u RedisClientImp) Set(key string, value interface{}, ttl time.Duration) (string, error) {
	return u.RedisClient.Set(context.Background(), key, value, ttl).Result()
}

func (u RedisClientImp) Del(keys ...string) (int64, error) {
	return u.RedisClient.Del(context.Background(), keys...).Result()
}

func (u RedisClientImp) GetRedisClient() *RedisClientImp {
	return redisClient
}
