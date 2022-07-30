package redis_cache

import (
	"context"
)

type Interface interface{
	SetDs1Metadata(ctx context.Context,userId string, metadata *DS1Metadata) error
}
