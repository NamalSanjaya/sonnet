package redis

import (
	"context"

	rds "github.com/go-redis/redis/v8"
)


type txPipeliner struct {
	txPipeline rds.Pipeliner
}

var _ Interface = (*txPipeliner)(nil)

func CreatePipe(p rds.Pipeliner) *txPipeliner{
	return &txPipeliner{
		txPipeline: p,
	}
}

func (tp txPipeliner) HSet(ctx context.Context, key string, values ...string) error {
	return tp.txPipeline.HSet(ctx, key, values).Err()
}

func (tp txPipeliner) Exec(ctx context.Context, key string) error {
	// TODO: should return []Cmds, err
	_, err := tp.txPipeline.Exec(ctx)
	return err
}
