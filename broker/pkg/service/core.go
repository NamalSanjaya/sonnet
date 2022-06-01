package service

import (
	"fmt"
	"time"

	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
)

type Core struct {
	Cache redis.Interface
	Queue store.Interface
}

func New(cache redis.Interface, queue store.Interface) *Core {
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
