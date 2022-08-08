package redis

import ( 
	"context" 
	rds "github.com/go-redis/redis/v8"
)

type Interface interface {
	HVals(ctx context.Context, key string)([]string,error)
	TxPipelined(ctx context.Context, fn func(p rds.Pipeliner) error) (string, error)
	ZRangeWithScore(ctx context.Context, key, min, max string, rev bool, offset, count int) ([]string, error)
	HSet(ctx context.Context, key string, values ...string) error
	LIndex(ctx context.Context, key string, indx int) (string, error)
}
