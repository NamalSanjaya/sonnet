// repo implementation of redis as persitance cache, which store memory and DS2
package redis_cache

import (
	"context"
	"fmt"
	"strconv"
	"time"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
)

// configuration parameter
const maxNoLockChecks int = 5
const waitingTime time.Duration = time.Millisecond * 1200

const PrefixDs2   string   = "ds2#"
const PrefixMem   string   = "mem#"
const RegHistTbs string = "reghistorytbs"

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

var _ Interface = (*redisRepo)(nil)

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
	if len(rawdata) == 0 {
		return nil, fmt.Errorf("unable to find history table with name %s", histTb)
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

// get memory block size
func (r *redisRepo) GetMemSize(ctx context.Context, histTb string) (int, error) {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	memSzStr, err := r.cmder.HGet(ctx, histTbKey, memsize)
	if err != nil {
		return 0, err  
	}
	memSzInt, err := strconv.Atoi(memSzStr)
	if err != nil {
		return 0, err
	}
	return memSzInt, nil
}

// set `lastdeleted` metadata
func (r *redisRepo) SetLastDel(ctx context.Context, histTb string, lastDel int) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	lastDelStr := strconv.Itoa(lastDel)
	return r.cmder.HSet(ctx, histTbKey, lastdeleted ,lastDelStr)
}

func (r *redisRepo) SetMemSize(ctx context.Context, histTb string, size int) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	sizeStr := strconv.Itoa(size)
	return r.cmder.HSet(ctx, histTbKey, memsize, sizeStr)
}

// `state` metadata, 0:lock  , 1: open
func (r *redisRepo) lock(ctx context.Context, histTb string) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(ctx, histTbKey, state , "0")
}

// `state` metadata, 0:lock  , 1: open
func (r *redisRepo) unlock(ctx context.Context,histTb string) error{
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(ctx, histTbKey, state , "1")
}

// lastDel <= rows < lastRead
func (r *redisRepo) ListMemoryRows(ctx context.Context, histTb string, lastDel, lastRead int) (MemoryRows, error) {
	memoryRows := MemoryRows{}
	histMemKey := fmt.Sprintf("%s%s",PrefixMem, histTb)
	minScore := fmt.Sprintf("%d", lastDel)
	mxScore := fmt.Sprintf("(%d", lastRead)
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
	sizeInt, err := strconv.Atoi(size)
	if err != nil {
		return &MemoryRow{Timestamp: timestampInt, Data: "", Size: 0}
	}
	return &MemoryRow{Timestamp: timestampInt, Data: data, Size: sizeInt}
}

func (r *redisRepo) GetMemRowSize(ctx context.Context, histTb, tmstamp string) (int,error){
	memRowKey := fmt.Sprintf("%s%s#%s",PrefixMem, histTb, tmstamp)
	sz, err := r.cmder.LIndex(ctx, memRowKey, 1)
	if err != nil {
		return -1, err
	}
	size, err := strconv.Atoi(sz)
	if err != nil {
		return -1, err
	}
	return size, nil
}

// return total removed memory size and delete lastDel <= rows < lastRead
func (r *redisRepo) RemoveMemRows(ctx context.Context, histTb string, lastDel, lastRead int) (int, error) {
	histMemKey := fmt.Sprintf("%s%s",PrefixMem, histTb)
	maxScore := fmt.Sprintf("(%d", lastRead)
	minScore := fmt.Sprintf("%d", lastDel)
	listTimeStamp, err := r.cmder.ZRangeByScore(ctx, histMemKey, minScore, maxScore)
	if err != nil {
		return 0, err
	}
	var totalDelSize int
	var delSize int
	for _, timestmp := range listTimeStamp {
		delSize = 0
		if delSize, err = r.GetMemRowSize(ctx, histTb, timestmp); err != nil {
			return 0, err
		}
		totalDelSize += delSize
		if err = r.cmder.Del(ctx, fmt.Sprintf("%s%s#%s", PrefixMem, histTb, timestmp)); err != nil {
			return 0, err
		}
	}
	err = r.cmder.ZRemRangeByScore(ctx, histMemKey, minScore, maxScore)
	if err != nil {
		return 0, err
	}
	return totalDelSize, nil
}

// lock ds2, so that memory can be safely use ( since implicity memory also lock)
func (r *redisRepo) LockMemory(ctx context.Context, histTb string) error {
	lockCheckCount := 0
	for lockCheckCount < maxNoLockChecks { // check 4s to lock the resources
		st, err := r.GetState(ctx, histTb)
		if err != nil {
			lockCheckCount++
			continue
		} 
		if st == 1 { // check current unlock state
			if lockErr := r.lock(ctx, histTb); lockErr == nil {
				break
			}
		}
		lockCheckCount++
		time.Sleep(waitingTime)
	}
	if lockCheckCount == maxNoLockChecks {
		return fmt.Errorf("unable to lock DS2 to remove memory rows in %s", histTb)
	}
	return nil
}

func (r *redisRepo) UnlockMemory(ctx context.Context, histTb string) error { 
	tryUnlockCount := 0
	for tryUnlockCount < maxNoLockChecks {
		if unlockErr := r.unlock(ctx, histTb); unlockErr == nil {
			return nil
		}
		tryUnlockCount++
		time.Sleep(waitingTime)
	}
	return fmt.Errorf("unable to unlock ds2 & memory of %s", histTb)
}

func (r *redisRepo) ListHistTbs(ctx context.Context)([]string, error) {
	key := fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
	return r.cmder.SMembers(ctx, key)
}

func (r *redisRepo) SetMemoryRowsSize(ctx context.Context,histTb string, sz int) error{
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	size := strconv.Itoa(sz)
	return r.cmder.HSet(ctx, histTbKey, memsize , size)
}
