package datasource2

import (
	"context"
	"fmt"
	"strconv"

	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
)

const PrefixDs2   string   = "ds2#"
const RegHistTbs string = "reghistorytbs"

const (
	userid      string   = "userid"  // to userid
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
	if err = rdb.createHistTb(ctx, newUserId, pairHistTb.Tx2Rx_HistTb); err != nil {
		return err
	}
	return rdb.createHistTb(ctx, userId, pairHistTb.Rx2Tx_HistTb)
}

func (rdb *redisDbRepo) createHistTb(ctx context.Context, userId, histTb string) error {
	if err := rdb.cmder.HSet(ctx, makeHistoryTbKey(histTb), userid, userId, lastmsg, "0", lastread, "0", 
	lastdeleted, "0", memsize, "0", state, "1"); err != nil {
		return err
	}
	return nil
}

// TODO: notFoundErr in GET methods should return ("", nil)
func (rdb *redisDbRepo) GetToUser(ctx context.Context, histTb string) (string, error) {
	return rdb.cmder.HGet(ctx, makeHistoryTbKey(histTb), userid)
}

func (rdb *redisDbRepo) GetLastMsg(ctx context.Context, histTb string) (int, error) {
	lstMsg, err := rdb.cmder.HGet(ctx, makeHistoryTbKey(histTb), lastmsg)
	if err != nil {
		return 0, nil
	}
	return strconv.Atoi(lstMsg)
}

func (rdb *redisDbRepo) SetLastRead(ctx context.Context, histTb string, lastRead int) error {
	lastReadStr := strconv.Itoa(lastRead)
	return rdb.cmder.HSet(ctx, makeHistoryTbKey(histTb), lastread ,lastReadStr)
}

func makeAllHistoryTbKey() string {
	return fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
}

func makeHistoryTbKey(histTb string) string {
	return fmt.Sprintf("%s%s", PrefixDs2, histTb)
}
