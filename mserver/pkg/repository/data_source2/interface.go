package datasource2

import ( 
	"context" 
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	trds2 "github.com/NamalSanjaya/sonnet/mserver/pkg/clients/transct_ds2"
)

type Interface interface{
	CreateHistTbs(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error
	GetToUser(ctx context.Context, histTb string) (string, error)
	GetLastMsg(ctx context.Context, histTb string) (int, error)
	Watch(ctx context.Context, txFn func(trds2.Interface) error, comtFn func(trds2.Interface) error, key string) error 
	MakeHistoryTbKey(histTb string) string
	MakeAllHistoryTbKey() string
}
