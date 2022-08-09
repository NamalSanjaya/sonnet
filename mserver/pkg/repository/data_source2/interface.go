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
	Watch(ctx context.Context, txFn func(tr trds2.Interface) error, comtFn func(trds2.Interface) error, keys ...string) error 
	MakeHistoryTbKey(histTb string) string
	MakeAllHistoryTbKey() string
	MakeHistMemKey(histTb string) string
	ListMemoryRows(ctx context.Context, histTb string, start, end int) (MemoryRows, error)
	IsSameToUser(ctx context.Context, userId, histTb string) (bool, error)
	CombineHistTbs(mem1, mem2 MemoryRows) MemoryRows
}
