package datasource2

import ( 
	"context" 
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type Interface interface{
	CreateHistTbs(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error
}
