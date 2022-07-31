package data_source1

import (
	"context"

	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type Interface interface{
	SetDs1Metadata(ctx context.Context,userId string, metadata *DS1Metadata) error
	AddBlockUser(ctx context.Context, userId, blockedUserId string) error
	CreateNewContact(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error
	RemoveBlockUser(ctx context.Context, userId, rmUserId string) error
}
