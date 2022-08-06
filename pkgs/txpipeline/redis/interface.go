package redis

import "context"

type Interface interface {
	HSet(ctx context.Context, key string, values ...string) error
	Exec(ctx context.Context, key string) error
}
