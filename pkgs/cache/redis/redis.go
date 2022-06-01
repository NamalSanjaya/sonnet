package redis

import (
	"fmt"

	rds "github.com/go-redis/redis/v7"
)

type Config struct {
	Host string
	Port int
	PassWord string
	DB int
}

type client struct {
	client *rds.Client
}

var _ Interface = (*client)(nil)

func NewClient(config *Config) *client {
	opt := &rds.Options{
		Addr: fmt.Sprintf("%s:%d",config.Host,config.Port),
		Password: config.PassWord,
		DB: config.DB,
	}
	return &client{
		client: rds.NewClient(opt),
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

func (rc *client) HMGet(key string, fields ...string)([]interface{}, error) {
	return rc.client.HMGet(key, fields...).Result()
}