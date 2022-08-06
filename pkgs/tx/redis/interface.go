package redis

import ( 
	"context" 
	rds "github.com/go-redis/redis/v8"
)

type Interface interface {
	HVals(ctx context.Context, key string)([]string,error)
	TxPipelined(ctx context.Context, fn func(p rds.Pipeliner) error) (string, error)
}
