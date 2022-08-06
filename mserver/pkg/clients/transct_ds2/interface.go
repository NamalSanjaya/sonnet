package transct_ds2

import (
	"context"

	rds "github.com/go-redis/redis/v8"
)

type Interface interface {
	AddPipeliner(p rds.Pipeliner)
	GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error)
	SetLastRead(ctx context.Context, histTb string, lastRead int) error
}
