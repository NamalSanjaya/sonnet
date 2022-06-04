// repo implementation of redis as persitance cache, which store memory and DS2
package redis_cache

import (
	"context"
	"fmt"
	"strconv"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
)

const PrefixDs2   string   = "ds2#"
const PrefixMem   string   = "mem#"

const (
	userid      string   = "userid"
 	lastmsg     string   = "lastmsg"
	lastread    string   = "lastread"
	lastdeleted string   = "lastdeleted"
	memsize     string   = "memsize"
	state       string   = "state"
)

type redisRepo struct{
	cmder redis.Interface
}

func NewRepo(cmder redis.Interface) *redisRepo {
	return &redisRepo{cmder: cmder}
}

// get metadata in DS2
func (r *redisRepo) GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error){
	metadata := &HistTbMetadata{}
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	rawdata, err := r.cmder.HVals(ctx, histTbKey)
	if err != nil {
		return nil, err
	}
	if len(rawdata) != 6 {
		return nil, fmt.Errorf("unexpected caching error in %s", histTb)
	}
	metadata.UserId = rawdata[0]
	temp := []int{}
	for _,elemt := range rawdata {
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

// `state` metadata, 0:lock  , 1: open , -1: err occuried
func (r *redisRepo) GetState(ctx context.Context, histTb string) (int, error) {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	strSt, err := r.cmder.HGet(ctx, histTbKey, state)
	if err != nil {
		return -1, err  
	}
	intSt, err := strconv.Atoi(strSt)
	if err != nil {
		return -1, err
	}
	return intSt, nil
}

// set `lastdeleted` metadata
func (r *redisRepo) SetLastDel(ctx context.Context, histTb string, lastDel string) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(ctx, histTbKey, lastdeleted ,lastDel)
}

// `state` metadata, 0:lock  , 1: open
func (r *redisRepo) Lock(ctx context.Context, histTb string) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(ctx, histTbKey, state , "0")
}

// `state` metadata, 0:lock  , 1: open
func (r *redisRepo) Unlock(ctx context.Context,histTb string) error{
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(ctx, histTbKey, state , "1")
}

func (r *redisRepo) ListMemoryRows(ctx context.Context, histTb string, lastDel, lastRead int) (MemoryRows, error) {
	memoryRows := MemoryRows{}
	histMemKey := fmt.Sprintf("%s%s",PrefixMem, histTb)
	minScore := fmt.Sprintf("(%d", lastDel)
	mxScore := fmt.Sprintf("%d", lastRead)
	listTimestamp, err := r.cmder.ZRangeByScore(ctx, histMemKey, minScore, mxScore)
	if err != nil {	
		return memoryRows, err
	}
	for _, timestmp := range listTimestamp {
		memRowKey := fmt.Sprintf("%s%s#%s", PrefixMem, histTb, timestmp)
		rowData, err := r.cmder.LRange(ctx, memRowKey)
		if err != nil || len(rowData) < 2 {
			rowData = []string{"", "0"}
		}
		memRow := r.prepareMemoryRow(timestmp, rowData[0], rowData[1])
		if memRow == nil {
			continue
		}
		memoryRows = append(memoryRows, memRow)
	}
	return memoryRows, nil
}

func (r *redisRepo) prepareMemoryRow(tmStamp, data, size string) *MemoryRow {
	timestampInt, err := strconv.Atoi(tmStamp)
	if err != nil {
		return nil
	}
	sizeInt, err := strconv.Atoi(tmStamp)
	if err != nil {
		return &MemoryRow{Timestamp: timestampInt, Data: "", Size: 0}
	}
	return &MemoryRow{Timestamp: timestampInt, Data: data, Size: sizeInt}
}

func (r *redisRepo) RemoveMemRows(ctx context.Context, histTb string, lastDel, lastRead int) error {
	minScore := strconv.Itoa(lastDel)
	maxScore := strconv.Itoa(lastRead)
	histMemKey := fmt.Sprintf("%s%s",PrefixMem, histTb)
	err := r.cmder.ZRemRangeByScore(ctx, histMemKey, minScore, maxScore)
	if err != nil {
		return err
	}
	minScoreExv := fmt.Sprintf("(%d", lastDel)
	listTimeStamp, err := r.cmder.ZRangeByScore(ctx, histMemKey, minScoreExv, maxScore)
	if err != nil {
		return err
	}
	for _, timestmp := range listTimeStamp {
		if err = r.cmder.Del(ctx, fmt.Sprintf("%s%s#%s", PrefixMem, histTb, timestmp)); err != nil {
			return err
		}
	}
	return nil
}
