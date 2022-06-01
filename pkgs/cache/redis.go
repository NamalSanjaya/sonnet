package cache

import (
	"fmt"

	redis "github.com/go-redis/redis/v7"
)

type Config struct {
	Host string
	Port string
	PassWord string
	DB int
}

type client struct {
	client *redis.Client
}

func NewClient(config Config) *client {
	opt := &redis.Options{
		Addr: fmt.Sprintf("%s:%s",config.Host,config.Port),
		Password: config.PassWord,
		DB: config.DB,
	}
	return &client{
		client: redis.NewClient(opt),
	}
}

func (rc *client) HSet(key, field, value string) error {
	return rc.client.HSet(key, field, value).Err()
}

func (rc *client) HGet(key, field string) (string, error) {
	return rc.client.HGet(key, field).Result()
}

func (rc *client) HDel(key string, fields ...string) error {
	return rc.client.HDel(key, fields...).Err()
}
