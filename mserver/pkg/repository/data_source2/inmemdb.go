package datasource2

import (
	"context"
	"fmt"

	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
)

const PrefixDs2   string   = "ds2#"
const RegHistTbs string = "reghistorytbs"

const (
	userid      string   = "userid"
 	lastmsg     string   = "lastmsg"
	lastread    string   = "lastread"
	lastdeleted string   = "lastdeleted"
	memsize     string   = "memsize"
	state       string   = "state"
)

type redisDbRepo struct {
	cmder redis.Interface
}

var _ Interface = (*redisDbRepo)(nil)

func NewRepo(cmder redis.Interface) *redisDbRepo {
	return &redisDbRepo{cmder: cmder}
}

// create history tables for both direction(owner <--> friend)
func (rdb *redisDbRepo) CreateHistTbs(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error {
	var err error
	if err = rdb.cmder.SAdd(ctx, makeAllHistoryTbKey(), pairHistTb.Tx2Rx_HistTb, pairHistTb.Rx2Tx_HistTb); err != nil {
		return err
	}
	if err = rdb.createHistTb(ctx, userId, pairHistTb.Tx2Rx_HistTb); err != nil {
		return err
	}
	return rdb.createHistTb(ctx, newUserId, pairHistTb.Rx2Tx_HistTb)
}

func (rdb *redisDbRepo) createHistTb(ctx context.Context, userId, histTb string) error {
	if err := rdb.cmder.HSet(ctx, makeHistoryTbKey(histTb), userid, userId, lastmsg, "0", lastread, "0", 
	lastdeleted, "0", memsize, "0", state, "1"); err != nil {
		return err
	}
	return nil
}

func makeAllHistoryTbKey() string {
	return fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
}

func makeHistoryTbKey(histTb string) string {
	return fmt.Sprintf("%s%s", PrefixDs2, histTb)
}