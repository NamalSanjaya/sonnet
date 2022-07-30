package redis

import (
	"context"
)

type Interface interface{
	Del(ctx context.Context, keys ...string) error
	RPush(ctx context.Context, key string, values ...string) error
	LRange(ctx context.Context, key string) ([]string, error)
	LIndex(ctx context.Context, key string, indx int)(string, error)
	HSet(ctx context.Context, key string, values ...string) error
	HGet(ctx context.Context,key, field string) (string, error)
	HDel(ctx context.Context, key string, fields ...string) error
	HMGet(ctx context.Context, key string, fields ...string)([]interface{}, error)
	HVals(ctx context.Context, key string)([]string,error)
	ZRemRangeByScore(ctx context.Context, key, min, max string) error
	ZRangeByScore(ctx context.Context, key string, minScore, maxScore string) ([]string, error)
	SMembers(ctx context.Context, key string)([]string, error)
	SRem(ctx context.Context, key string, values ...string) error
}
