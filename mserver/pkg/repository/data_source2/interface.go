package datasource2

import ( 
	"context" 
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type Interface interface{
	CreateHistTbs(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error
	GetToUser(ctx context.Context, histTb string) (string, error)
	GetLastMsg(ctx context.Context, histTb string) (int, error)
	SetLastRead(ctx context.Context, histTb string, lastRead int) error
	GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error)
}
