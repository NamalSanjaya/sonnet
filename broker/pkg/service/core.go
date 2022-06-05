package service

import (
	"context"
	"fmt"
	"log"

	histbrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/historytable"
	redisrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/redis_cache"
	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
)

type Core struct {
	Cache redisrepo.Interface
	Queue store.Interface
	HistRepo histbrepo.Interface
}

func New(cache redisrepo.Interface, queue store.Interface, histtb histbrepo.Interface) *Core {
	return &Core{
		Cache: cache,
		Queue: queue,
		HistRepo: histtb,
	}
}

func (c *Core) JobExec(ctx context.Context, initChan chan int) {
	for {
		num := <- initChan
		switch num {
		case 1:
			c.exec(ctx)
			fmt.Println("------ Job end ------")
		case 2:
			fmt.Println("error")
		default:
			fmt.Println("unExpected error occuried...")
		}
	}
}

func (c *Core) exec(ctx context.Context) {
	fmt.Println("------ Job start ------")
	for c.Queue.Len() > 0 {
		task := c.Queue.Dequeue()
		if err := c.MoveMemoryRowsToDB(ctx, task.OwnerHistTb); err != nil {
			fmt.Println(err.Error())
		}
		fmt.Println("finished...now queue size: ", c.Queue.Len(), task.OwnerHistTb)
	}
}

func (c *Core) MoveMemoryRowsToDB(ctx context.Context, histTb string) error {
	metadata, err := c.Cache.GetAllMetadata(ctx,histTb) 
	if err != nil {
		return fmt.Errorf("unable to get metadata in ds2 of %s due to, %s", histTb,err.Error())
	}
	fmt.Println("fetch all metadata...")
	if err := c.Cache.LockMemory(ctx, histTb); err != nil {
		return fmt.Errorf("unable to lock ds2 of %s due to %s", histTb, err.Error())
	}
	memRows, err := c.Cache.ListMemoryRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead)
	if err != nil {
		return fmt.Errorf("unable to list memory rows of %s due to %s", histTb, err.Error())
	}
	if len(memRows) == 0{
		log.Println("no memory rows are found to move to DB")
		return nil
	}
	fmt.Println("list all memory rows..")
	// creating query DB query
	dataStr := ""
	for i, memRow := range memRows {
		if i == 0 {
			dataStr = fmt.Sprintf("(%d,0x%s,%d)", memRow.Timestamp, memRow.Data, memRow.Size)
			continue
		}
		dataStr = dataStr + "," + fmt.Sprintf("(%d,0x%s,%d)", memRow.Timestamp, memRow.Data, metadata.MemSize)
	}
	if err = c.HistRepo.InsetMsgs(ctx, histTb, dataStr); err != nil {
		return fmt.Errorf("msg insertion of %s to DB was failed due to %s", histTb, err.Error())
	}
	fmt.Println("inserted msgs to db")
	if err = c.Cache.RemoveMemRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead); err != nil {
		return fmt.Errorf("unable to remove memory rows of %s which are already moved to DB due to, %s", histTb, err.Error())
	}
	fmt.Println("remove memory rows from redis-memory")
	if err = c.Cache.SetLastDel(ctx, histTb, metadata.LastRead); err != nil {
		return fmt.Errorf("falied to set `lastdeleted` metadata of %s correctly due to, %s", histTb, err.Error())
	}
	return c.Cache.UnlockMemory(ctx, histTb)
}
