package transct_ds2

import (
	"context"
	"fmt"
	"strconv"

	rdtx "github.com/NamalSanjaya/sonnet/pkgs/tx/redis"
	txpipe "github.com/NamalSanjaya/sonnet/pkgs/txpipeline/redis"
	rds "github.com/go-redis/redis/v8"
)

const 
(
	PrefixDs2   string   = "ds2#"
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

// func makeAllHistoryTbKey() string {
// 	return fmt.Sprintf("%s%s", PrefixDs2, RegHistTbs)
// }

func makeHistoryTbKey(histTb string) string {
	return fmt.Sprintf("%s%s", PrefixDs2, histTb)
}
