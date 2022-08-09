package service

import (
	"context"
	"fmt"

	lg "github.com/labstack/gommon/log"

	histbrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/historytable"
	redisrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/redis_cache"
	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
)

type Core struct {
	Cache    redisrepo.Interface
	Queue    store.Interface
	HistRepo histbrepo.Interface
	Logger   *lg.Logger
	State    string
}

func New(cache redisrepo.Interface, queue store.Interface, histtb histbrepo.Interface, logger *lg.Logger) *Core {
	return &Core{
		Cache:    cache,
		Queue:    queue,
		HistRepo: histtb,
		Logger:   logger,
		State:    "idle",
	}
}

func (c *Core) JobExec(ctx context.Context, initChan, closeChan chan int) {
	c.Logger.Info("Job Execution starts")
	for {
		select {
		case <-ctx.Done():
			if c.State == "idle" {
				c.Logger.Info("exit from Job Execution")
				closeChan <- 1
				return
			} else {
				c.State = "off"
			}
		case <-initChan:
			c.State = "exec"
			c.exec(ctx, closeChan)
		case <-closeChan:
			c.Logger.Info("exit from Job Execution")
			return
		}
	}
}

func (c *Core) exec(ctx context.Context, closeCh chan int) {
	for c.Queue.Len() > 0 && c.State == "exec" {
		histTb := c.Queue.Dequeue()
		c.Logger.Infof("Memory cleaning starts for %s", histTb)
		var removedRowsCnt int
		var mvErr error
		if removedRowsCnt, mvErr = c.MoveMemoryRowsToDB(ctx, histTb); mvErr != nil {
			if unlockErr := c.Cache.UnlockMemory(ctx, histTb); unlockErr != nil {
				c.Logger.Error(unlockErr)
			}
			c.Logger.Error(mvErr)
			continue
		}
		c.Logger.Infof("move %dB size of redis-memory block to db for %s", removedRowsCnt, histTb)

	}
	if c.State == "off" {
		closeCh <- 1
		return
	}
	c.State = "idle"
}

func (c *Core) MoveMemoryRowsToDB(ctx context.Context, histTb string) (int, error) {
	var totalRemovedSz int = 0
	metadata, err := c.Cache.GetAllMetadata(ctx, histTb)
	if err != nil {
		return totalRemovedSz, fmt.Errorf("unable to get metadata in ds2 of %s due to, %s", histTb, err.Error())
	}
	if err := c.Cache.LockMemory(ctx, histTb); err != nil {
		return totalRemovedSz, fmt.Errorf("unable to lock ds2 of %s due to %s", histTb, err.Error())
	}
	memRows, err := c.Cache.ListMemoryRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead)
	if err != nil {
		return totalRemovedSz, fmt.Errorf("unable to list memory rows of %s due to %s", histTb, err.Error())
	}
	if len(memRows) == 0 {
		return totalRemovedSz, c.Cache.UnlockMemory(ctx, histTb)
	}
	// creating query DB query
	dataStr := ""
	for i, memRow := range memRows {
		if i == 0 {
			dataStr = fmt.Sprintf("(%d,%s,%d)", memRow.Timestamp, memRow.Data, memRow.Size)
			continue
		}
		dataStr = dataStr + "," + fmt.Sprintf("(%d,0x%s,%d)", memRow.Timestamp, memRow.Data, memRow.Size)
	}
	if err = c.HistRepo.InsetMsgs(ctx, histTb, dataStr); err != nil {
		return totalRemovedSz, fmt.Errorf("msg insertion of %s to DB was failed due to %s", histTb, err.Error())
	}
	if totalRemovedSz, err = c.Cache.RemoveMemRows(ctx, histTb, metadata.LastDeleted, metadata.LastRead); err != nil {
		return totalRemovedSz, fmt.Errorf("unable to remove memory rows of %s which are already moved to DB due to, %s", histTb, err.Error())
	}
	if err = c.Cache.SetLastDel(ctx, histTb, metadata.LastRead); err != nil {
		return totalRemovedSz, fmt.Errorf("falied to set `lastdeleted` metadata of %s correctly due to, %s", histTb, err.Error())
	}
	if metadata.MemSize-totalRemovedSz < 0 {
		return totalRemovedSz, fmt.Errorf("memory block size can't be a negative value for %s", histTb)
	}
	if err = c.Cache.SetMemSize(ctx, histTb, metadata.MemSize-totalRemovedSz); err != nil {
		return totalRemovedSz, fmt.Errorf("unable to set metadata of %s correctly due to, %s", histTb, err.Error())
	}
	return totalRemovedSz, c.Cache.UnlockMemory(ctx, histTb)
}
