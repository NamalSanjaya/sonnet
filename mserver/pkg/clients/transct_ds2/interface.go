package transct_ds2

import (
	"context"

	rds "github.com/go-redis/redis/v8"
)

type Interface interface {
	AddPipeliner(p rds.Pipeliner)
	GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error)
	SetLastRead(ctx context.Context, histTb string, lastRead int) error
	SetLastMsg(ctx context.Context, histTb string, lastMsg int) error
	SetMemSize(ctx context.Context, histTb string, memSz int) error
	GetAdjacentTimeStamp(ctx context.Context, histTb string, min, max int)(int, error)
	RemMemRow(ctx context.Context, histTb string, timestamp int) error 
	GetMemRowSize(ctx context.Context, histTb string, timestamp int) (int, error)
}
