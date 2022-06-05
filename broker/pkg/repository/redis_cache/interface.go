package redis_cache

import (
	"context"
)

type Interface interface {
	GetState(ctx context.Context, histTb string) (int, error)
	SetLastDel(ctx context.Context, histTb string, lastDel int) error
	GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error)
	SetMemSize(ctx context.Context, histTb string, size int) error
	LockMemory(ctx context.Context, histTb string) error
	UnlockMemory(ctx context.Context, histTb string) error
	ListMemoryRows(ctx context.Context, histTb string, lastDel, lastRead int) (MemoryRows, error)
	RemoveMemRows(ctx context.Context, histTb string, lastDel, lastRead int) (int, error)
	GetMemRowSize(ctx context.Context, histTb, tmstamp string) (int,error)
	ListHistTbs(ctx context.Context)([]string, error)
}
