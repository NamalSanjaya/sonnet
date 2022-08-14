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

const (
	maxRowsIterations int = 30
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
	if err == rds.Nil {
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

func (rdb *redisDbRepo) GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error) {
	metadata := &HistTbMetadata{}
	rawdata, err := rdb.cmder.HVals(ctx, rdb.MakeHistoryTbKey(histTb))
	if err != nil {
		return nil, err
	}
	if len(rawdata) == 0 { // not found
		return nil, nil
	}
	if len(rawdata) != 6 {
		return nil, fmt.Errorf("partially cached metadata was found in DS2 for %s", histTb)
	}
	metadata.UserId = rawdata[0]
	temp := []int{}
	for _,elemt := range rawdata[1:] {
		d, err := strconv.Atoi(elemt)
		if err != nil {
			return nil, err
		}
		temp = append(temp, d)
	}
	metadata.Lastmsg     = temp[0]
	metadata.LastRead    = temp[1]
	metadata.LastDeleted = temp[2]
	metadata.MemSize     = temp[3]
	metadata.State 		 = temp[4]
	return metadata, nil
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

// Memory Content, Fetched total data size
func (rdb *redisDbRepo) ListMemoryRows(ctx context.Context, histTb string, start, end int) (MemoryRows, int, error) {
	memoryRows := MemoryRows{}
	minScore := fmt.Sprintf("%d", start)
	mxScore := fmt.Sprintf("%d", end)
	listTimestamp, err := rdb.cmder.ZRangeWithScore(ctx, rdb.MakeHistMemKey(histTb), minScore, mxScore, true, 0, -1)
	if err != nil {	
		return memoryRows, 0, fmt.Errorf("falied to list Memory rows of %s history table due to %w", histTb, err)
	}
	var totalMemSz int
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
		totalMemSz += memRow.Size
		memoryRows = append(memoryRows, memRow)
	}
	return memoryRows, totalMemSz, nil
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

// NOT IN USED, 
// can fetch upto/beyond `mxSize`.
// zero timestamp means we don't need to go to db
func (rdb *redisDbRepo) listMemoryRowsForSize(ctx context.Context, histTb string, maxTimestamp string, mxSize int) (MemoryRows, int, error) {
	var totalMemSz int
	var memRows MemoryRows
	timestamps, err := rdb.cmder.ZRangeWithScore(ctx, rdb.MakeHistMemKey(histTb), "-inf", maxTimestamp, true, 0, maxRowsIterations)
	if err != nil {
		return memRows, 0, err
	}
	var tmstp string
	var tmInt int
	for _, tmstp = range timestamps {
		memRowKey := makeMemRowKey(histTb, tmstp)
		rowData, err := rdb.cmder.LRange(ctx, memRowKey)
		if err != nil || len(rowData) < 2 {
			continue
		}
		memRow := rdb.prepareMemoryRow(tmstp, rowData[0], rowData[1])
		if memRow == nil {
			continue
		}
		totalMemSz += memRow.Size
		memRows = append(memRows, memRow)
		if totalMemSz >= mxSize {
			tmInt, err = strconv.Atoi(tmstp)
			if err != nil {
				return MemoryRows{}, 0, err
			}
			return memRows, tmInt, nil
		}
	}
	tmInt, err = strconv.Atoi(tmstp)
	if err != nil {
		return MemoryRows{}, 0, err
	}
	return memRows, tmInt, nil
}

// NOT IN USED, 
// memory rows in lastest one to the top.
// `lastTimestamp` it obtained from redis-memory
func(rdb *redisDbRepo) GetMemoryContent(ctx context.Context, histTb string, minTimestamp, maxTimestamp, expectLoadSz int, isUnRead bool) (MemoryRows, int, error) {
	if isUnRead { // need to get at least `lastDeleted`
		memRows, size, err := rdb.ListMemoryRows(ctx, histTb, minTimestamp, maxTimestamp) // get memory rows upto `lastdeleted`
		if err != nil {
			return MemoryRows{}, 0, err
		}
		if expectLoadSz <= size {
			return memRows, minTimestamp, nil
		}
		memRowsAfterDel, lastTimestamp, err := rdb.listMemoryRowsForSize(ctx, histTb, fmt.Sprintf("(%d", minTimestamp), expectLoadSz - size)  // memory rows after last deleted
		// TODO: log the error
		if err != nil {
			return memRows, minTimestamp, nil
		}
		memRows = append(memRows, memRowsAfterDel...)
		return memRows, lastTimestamp, nil
	}
	return rdb.listMemoryRowsForSize(ctx, histTb, fmt.Sprintf("(%d", minTimestamp), expectLoadSz) 
}
