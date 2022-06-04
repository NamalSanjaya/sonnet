package service

import (
	"context"
	"fmt"
	"log"
	"time"

	hstb "github.com/NamalSanjaya/sonnet/broker/pkg/repository/historytable"
	rcache "github.com/NamalSanjaya/sonnet/broker/pkg/repository/redis_cache"
	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
)

// configuration parameters
const maxNoLockChecks int = 5
const waitingTime time.Duration = time.Millisecond * 800

type Core struct {
	Cache rcache.Interface
	Queue store.Interface
	HistRepo hstb.Interface
}

func New(cache rcache.Interface, queue store.Interface) *Core {
	return &Core{
		Cache: cache,
		Queue: queue,
	}
}

func (c *Core) JobExec(initChan chan int) {
	for {
		num := <- initChan
		switch num {
		case 1:
			c.exec()
			fmt.Println("------ Job end ------")
		case 2:
			fmt.Println("error")
		default:
			fmt.Println("unExpected error occuried...")
		}
	}
}

func (c *Core) exec() {
	fmt.Println("------ Job start ------")
	for c.Queue.Len() > 0 {
		task := c.Queue.Dequeue()
		fmt.Println("finished...now queue size: ", c.Queue.Len(), task.OwnerId)
	}
}

func (c *Core) MoveMemoryRowsToDB(ctx context.Context, histTb string) error {
	metadata, err := c.Cache.GetAllMetadata(ctx,histTb) 
	if err != nil {
		return err
	}
	lockCheckCount := 0
	for lockCheckCount < maxNoLockChecks { // check 4s to lock the resources
		st, err := c.Cache.GetState(ctx, histTb)
		if err != nil {
			lockCheckCount++
			continue
		} 
		if st == 1 { // check current unlock state
			if lockErr := c.Cache.Lock(ctx, histTb); lockErr == nil {
				break
			}
		}
		lockCheckCount++
		time.Sleep(waitingTime)
	}
	if lockCheckCount == maxNoLockChecks {
		return fmt.Errorf("unable to lock DS2 to remove memory rows in %s", histTb)
	}
	memRows, err := c.Cache.ListMemoryRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead)
	if err != nil {
		return err
	}
	// creating query DB query
	dataStr := ""
	for i, memRow := range memRows {
		if i == 0 {
			dataStr = fmt.Sprintf("(%d,0x%s,%d)", memRow.Timestamp, memRow.Data, metadata.MemSize)
			continue
		}
		dataStr = dataStr + "," + fmt.Sprintf("(%d,0x%s,%d)", memRow.Timestamp, memRow.Data, metadata.MemSize)
	}
	if err = c.HistRepo.InsetMsgs(ctx, histTb, dataStr); err != nil {
		return err
	}
	if err = c.Cache.RemoveMemRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead); err != nil {
		return err
	}
	if err = c.Cache.SetLastDel(ctx, histTb, metadata.LastRead); err != nil {
		return err
	}
	tryUnlockCount := 0
	for tryUnlockCount < maxNoLockChecks {
		if unlockErr := c.Cache.Unlock(ctx, histTb); unlockErr == nil {
			return nil
		}
		tryUnlockCount++
		time.Sleep(waitingTime)
	}
	log.Printf("unable to unlock metadata in DS2 of %s own by userId %s\n", histTb, metadata.UserId)
	return nil
}
