package redis

import (
	"context"
	"fmt"

	rds "github.com/go-redis/redis/v8"
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

func (rc *client) Del(ctx context.Context, keys ...string) error {
	return rc.client.Del(ctx, keys...).Err()
}

func (rc *client) RPush(ctx context.Context, key string, values ...string) error {
	return rc.client.RPush(ctx, key, values).Err()
}

func (rc *client) LRange(ctx context.Context, key string) ([]string, error){
	return rc.client.LRange(ctx, key, 0, -1).Result()
}

func (rc *client) LIndex(ctx context.Context, key string, indx int)(string, error){
	return rc.client.LIndex(ctx, key, int64(indx)).Result()
}

func (rc *client) HSet(ctx context.Context, key string, values ...string) error {
	return rc.client.HSet(ctx, key, values).Err()
}

func (rc *client) HGet(ctx context.Context,key, field string) (string, error) {
	return rc.client.HGet(ctx, key, field).Result()
} 

func (rc *client) HDel(ctx context.Context, key string, fields ...string) error {
	return rc.client.HDel(ctx, key, fields...).Err()
}

func (rc *client) HMGet(ctx context.Context, key string, fields ...string)([]interface{}, error) {
	return rc.client.HMGet(ctx, key, fields...).Result()
}

func (rc *client) HVals(ctx context.Context, key string)([]string,error) {
	return rc.client.HVals(ctx, key).Result()
}

func  (rc *client) ZRemRangeByScore(ctx context.Context, key, min, max string) error {
	return rc.client.ZRemRangeByScore(ctx, key, min, max).Err()
}

func (rc *client) ZRangeByScore(ctx context.Context, key string, minScore, maxScore string) ([]string, error) {
	arg := rds.ZRangeArgs{Key: key, Start: minScore, Stop: maxScore, ByScore: true}
	return rc.client.ZRangeArgs(ctx, arg).Result()
} 

func (rc *client) SMembers(ctx context.Context, key string)([]string, error) {
	return rc.client.SMembers(ctx, key).Result()
}

func (rc *client) SRem(ctx context.Context, key string, values ...string) error {
	return rc.client.SRem(ctx, key, values).Err()
}

func (rc *client) SAdd(ctx context.Context, key string, value ...string) error {
	return rc.client.SAdd(ctx, key, value).Err()
}

// addding set of blockusers with uniqueness
func (rc *client) SSet(ctx context.Context, key string, value ...string) error {
	if err := rc.client.Del(ctx, key).Err(); err != nil {
		return err
	}
	if len(value) == 0 {
		return nil
	}
	return rc.client.SAdd(ctx, key, value).Err()
}
