package redis

import (
	"context"
	"time"

	rds "github.com/go-redis/redis/v8"
)

type Interface interface{
	Set(ctx context.Context, key, value string, expireTime time.Duration) error
	Get(ctx context.Context, key string) (string, error)
	Del(ctx context.Context, keys ...string) error
	RPush(ctx context.Context, key string, values ...string) error
	LRange(ctx context.Context, key string) ([]string, error)
	LIndex(ctx context.Context, key string, indx int)(string, error)
	HSet(ctx context.Context, key string, values ...string) error
	HGet(ctx context.Context,key, field string) (string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HMGet(ctx context.Context, key string, fields ...string)([]string, error)
	HVals(ctx context.Context, key string)([]string,error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) error
	ZRangeByScore(ctx context.Context, key string, minScore, maxScore string) ([]string, error)
	SMembers(ctx context.Context, key string)([]string, error)
	SRem(ctx context.Context, key string, values ...string) error
	SAdd(ctx context.Context, key string, value ...string) error
	SSet(ctx context.Context, key string, value ...string) error
	Watch(ctx context.Context, txFn func(*rds.Tx) error , keys ...string) error
	MakeTxPipeliner() rds.Pipeliner
	ZRangeWithScore(ctx context.Context, key, min, max string, rev bool, offset, count int) ([]string, error)
}
