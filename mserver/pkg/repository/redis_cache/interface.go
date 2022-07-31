package redis_cache

import (
	"context"
)

type Interface interface{
	SetDs1Metadata(ctx context.Context,userId string, metadata *DS1Metadata) error
	AddBlockUser(ctx context.Context, userId, blockedUserId string) error
}
