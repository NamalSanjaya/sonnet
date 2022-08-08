package transct_ds2

import (
	"context"
	"fmt"
	"strconv"

	rdtx "github.com/NamalSanjaya/sonnet/pkgs/tx/redis"
	txpipe "github.com/NamalSanjaya/sonnet/pkgs/txpipeline/redis"
	rds "github.com/go-redis/redis/v8"
)

const (
	prefixMem string = "mem#"
	prefixDs2   string   = "ds2#"
    regHistTbs string = "reghistorytbs"
)

const (
	userid      string   = "userid"  // to userid
 	lastmsg     string   = "lastmsg"
	lastread    string   = "lastread"
	lastdeleted string   = "lastdeleted"
	memsize     string   = "memsize"
	state       string   = "state"
)

// transct repo consists a redis txpipline
type transctRepo struct {
	transct rdtx.Interface
	transctPipe txpipe.Interface
}

var _ Interface = (*transctRepo)(nil)

func BeginTransct(tx rdtx.Interface) *transctRepo {
	return &transctRepo{transct: tx}
}

func (tr *transctRepo) AddPipeliner(p rds.Pipeliner) {
	tr.transctPipe = txpipe.CreatePipe(p)
}

// NOTE: not in used
func (tr *transctRepo) TransctPipelined(ctx context.Context, fn func() error) error{
	return nil
}

func (tr *transctRepo) GetAllMetadata(ctx context.Context, histTb string) (*HistTbMetadata, error) {
	metadata := &HistTbMetadata{}
	rawdata, err := tr.transct.HVals(ctx, makeHistoryTbKey(histTb))
	if err != nil {
		return nil, err
	}
	if len(rawdata) == 0 {
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

func (tr *transctRepo) SetLastRead(ctx context.Context, histTb string, lastRead int) error {
	lastReadStr := strconv.Itoa(lastRead)
	return tr.transctPipe.HSet(ctx, makeHistoryTbKey(histTb), lastread, lastReadStr)
}

func (tr *transctRepo) SetLastMsg(ctx context.Context, histTb string, lastMsg int) error {
	lastMsgStr := strconv.Itoa(lastMsg)
	return tr.transctPipe.HSet(ctx, makeHistoryTbKey(histTb), lastmsg, lastMsgStr)
}

func (tr *transctRepo) SetMemSize(ctx context.Context, histTb string, memSz int) error {
	return tr.transctPipe.HSet(ctx, makeHistoryTbKey(histTb), memsize, fmt.Sprintf("%d", memSz))
}

func (tr *transctRepo) GetAdjacentTimeStamp(ctx context.Context, histTb string, min, max int)(int, error){
	var val int
	minScore := fmt.Sprintf("%d", min)
	maxScore := fmt.Sprintf("(%d", max)
	arr, err := tr.transct.ZRangeWithScore(ctx, makeHistMemKey(histTb), maxScore, minScore, true, 0, 1)
	if err != nil {
		return val, fmt.Errorf("unable to get adjacent memory timestamp from redis of %s due to %w", histTb, err)
	}
	if len(arr) == 0 {
		return val, fmt.Errorf("metadata of %s in ds2 is in a wrong state in timestamp range of [%d,%d)", histTb, min, max)
	}
	if val, err = strconv.Atoi(arr[0]); err != nil {
		return val, fmt.Errorf("unable to convert adjacent memory timestamp to an int of %s due to %w", histTb, err)
	}
	return val, nil
}

func (tr *transctRepo) RemMemRow(ctx context.Context, histTb string, timestamp int) error {
	var err error
	if err = tr.transctPipe.ZRem(ctx, makeHistMemKey(histTb), fmt.Sprintf("%d", timestamp)); err != nil {
		return fmt.Errorf("unable to remove timestamp %d from mem#%s due to %w", timestamp, histTb, err)
	}
	return tr.transctPipe.Del(ctx, makeMemRowKey(histTb, timestamp))
}

func (tr *transctRepo) GetMemRowSize(ctx context.Context, histTb string, timestamp int) (int, error) {
	var size int
	sizeStr, err := tr.transct.LIndex(ctx, makeMemRowKey(histTb, timestamp), 1)
	if err != nil {
		return 0, fmt.Errorf("unable to get memory size of histtb %s at %d timestamp due to %w", histTb, timestamp, err)
	}
	if size, err = strconv.Atoi(sizeStr); err != nil {
		return 0, fmt.Errorf("unable convert memory row size %s to int of histTb %s at %d timestamp due to %w", 
		sizeStr, histTb, timestamp, err)
	}
	return size, nil
}

// func makeAllHistoryTbKey() string {
// 	return fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
// }

func makeHistoryTbKey(histTb string) string {
	return fmt.Sprintf("%s%s", prefixDs2, histTb)
}

func makeHistMemKey(histTb string) string {
	return fmt.Sprintf("%s%s", prefixMem, histTb)
}

func makeMemRowKey(histTb string, tmStamp int) string {
	return fmt.Sprintf("%s%s#%d", prefixMem, histTb, tmStamp)
}
