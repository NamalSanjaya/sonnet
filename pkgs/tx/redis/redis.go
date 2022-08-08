package redis

import (
	"context"

	rds "github.com/go-redis/redis/v8"
)

// txer uses redis tx, pipline and watch features
type txer struct {
	tx *rds.Tx
}

var _ Interface = (*txer)(nil)

func CreateTx(tx *rds.Tx) *txer {
	return &txer{
		tx: tx,
	}
}

func (t *txer) HVals(ctx context.Context, key string)([]string, error) {
	return t.tx.HVals(ctx, key).Result()
}

func (t *txer) TxPipelined(ctx context.Context, fn func(p rds.Pipeliner) error) (string, error){
	_, err := t.tx.TxPipelined(ctx, fn)
	return "", err
}

func (t *txer) ZRangeWithScore(ctx context.Context, key, min, max string, 
	rev bool, offset, count int) ([]string, error) {
	return t.tx.ZRangeArgs(ctx, rds.ZRangeArgs{
		Key: key, Start: max, Stop: min, ByScore: true, ByLex: false, 
		Rev: rev, Offset: int64(offset), Count: int64(count),
	}).Result()
}

func (t *txer) HSet(ctx context.Context, key string, values ...string) error {
	return t.tx.HSet(ctx, key, values).Err()
}

func (t *txer) LIndex(ctx context.Context, key string, indx int) (string, error) {
	return t.tx.LIndex(ctx, key, int64(indx)).Result()
}
