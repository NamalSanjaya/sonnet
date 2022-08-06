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
