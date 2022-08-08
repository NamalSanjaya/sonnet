package redis

import "context"

type Interface interface {
	HSet(ctx context.Context, key string, values ...string) error
	Exec(ctx context.Context, key string) error
	Del(ctx context.Context, keys ...string) error
	ZRem(ctx context.Context, key string, members ...string) error
}
