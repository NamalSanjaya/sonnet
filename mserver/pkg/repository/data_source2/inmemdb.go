package datasource2

import (
	"context"
	"fmt"
	"sort"
	"strconv"

	rds "github.com/go-redis/redis/v8"

	trds2 "github.com/NamalSanjaya/sonnet/mserver/pkg/clients/transct_ds2"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	rdtx "github.com/NamalSanjaya/sonnet/pkgs/tx/redis"
)

const 
(
	prefixMem  string = "mem#"
	PrefixDs2  string = "ds2#"
    RegHistTbs string = "reghistorytbs"
)

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
	if err = rdb.cmder.SAdd(ctx, rdb.MakeAllHistoryTbKey(), pairHistTb.Tx2Rx_HistTb, pairHistTb.Rx2Tx_HistTb); err != nil {
		return err
	}
	if err = rdb.createHistTb(ctx, newUserId, pairHistTb.Tx2Rx_HistTb); err != nil {
		return err
	}
	return rdb.createHistTb(ctx, userId, pairHistTb.Rx2Tx_HistTb)
}

func (rdb *redisDbRepo) createHistTb(ctx context.Context, userId, histTb string) error {
	if err := rdb.cmder.HSet(ctx, rdb.MakeHistoryTbKey(histTb), userid, userId, lastmsg, "0", lastread, "0", 
	lastdeleted, "0", memsize, "0", state, "1"); err != nil {
		return err
	}
	return nil
}

// TODO: notFoundErr in GET methods should return ("", nil)
func (rdb *redisDbRepo) GetToUser(ctx context.Context, histTb string) (string, error) {
	id, err := rdb.cmder.HGet(ctx, rdb.MakeHistoryTbKey(histTb), userid)
	if err == nil {
		return id, nil
	}
	if err == rds.Nil{
		return "", nil
	}
	return "", err
}

func (rdb *redisDbRepo) GetLastMsg(ctx context.Context, histTb string) (int, error) {
	lstMsg, err := rdb.cmder.HGet(ctx, rdb.MakeHistoryTbKey(histTb), lastmsg)
	if err == nil {
		return strconv.Atoi(lstMsg)
	}
	if err == rds.Nil {
		return 0, nil
	}
	return 0, err
}

func (rdb *redisDbRepo) Watch(ctx context.Context, txFn func(trds2.Interface) error, comtFn func(trds2.Interface) error, keys ...string) error {
	return rdb.cmder.Watch(ctx, func(tx *rds.Tx) error {
		newTransct := trds2.BeginTransct(rdtx.CreateTx(tx))
		err := txFn(newTransct)
		if err != nil {
			return err
		}
		_, err = tx.TxPipelined(ctx, func(p rds.Pipeliner) error {
			newTransct.AddPipeliner(p)
			return comtFn(newTransct)
		})
		return err
	}, keys...)
}

func (rdb *redisDbRepo) MakeAllHistoryTbKey() string {
	return fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
}

func (rdb *redisDbRepo) MakeHistoryTbKey(histTb string) string {
	return fmt.Sprintf("%s%s", PrefixDs2, histTb)
}

func (rdb *redisDbRepo) MakeHistMemKey(histTb string) string {
	return fmt.Sprintf("%s%s", prefixMem, histTb)
}

func makeMemRowKey(histTb, tmStamp string) string {
	return fmt.Sprintf("%s%s#%s", prefixMem, histTb, tmStamp)
}

func (rdb *redisDbRepo) ListMemoryRows(ctx context.Context, histTb string, start, end int) (MemoryRows, error) {
	memoryRows := MemoryRows{}
	histMemKey := rdb.MakeHistMemKey(histTb)
	minScore := fmt.Sprintf("%d", start)
	mxScore := fmt.Sprintf("%d", end)
	listTimestamp, err := rdb.cmder.ZRangeByScore(ctx, histMemKey, minScore, mxScore)
	if err != nil {	
		return memoryRows, fmt.Errorf("falied to list Memory rows of %s history table due to %w", histTb, err)
	}
	for _, timestmp := range listTimestamp {
		memRowKey := makeMemRowKey(histTb, timestmp)
		rowData, err := rdb.cmder.LRange(ctx, memRowKey)
		if err != nil || len(rowData) < 2 {
			continue
		}
		memRow := rdb.prepareMemoryRow(timestmp, rowData[0], rowData[1])
		if memRow == nil {
			continue
		}
		memoryRows = append(memoryRows, memRow)
	}
	return memoryRows, nil
}

func (rdb *redisDbRepo) prepareMemoryRow(tmStamp, data, size string) *MemoryRow {
	timestampInt, err := strconv.Atoi(tmStamp)
	if err != nil {
		return nil
	}
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return &MemoryRow{Timestamp: timestampInt, Data: "", Size: 0}
	}
	return &MemoryRow{Timestamp: timestampInt, Data: data, Size: sizeInt}
}

// check the equality ToUserId in given history table with given userId
func (rdb *redisDbRepo) IsSameToUser(ctx context.Context, userId, histTb string) (bool, error) {
	toUserId, err := rdb.cmder.HGet(ctx, rdb.MakeHistoryTbKey(histTb), userid)
	if err != nil {
		return false, fmt.Errorf("unable to get TouserId from history tb %s with %w", histTb, err)
	}
	return toUserId == userId, nil
}

// combine two history tables and arrange them based on thier timestamp
/* error considerations: timestamp = 0 or data = "" : discard those situations */
func (rdb *redisDbRepo) CombineHistTbs(mem1, mem2 MemoryRows) MemoryRows {
	result := append(mem1, mem2...)
	sort.Slice(result, func(i, j int) bool {
		return result[i].Timestamp > result[j].Timestamp
	})
	return result
}
