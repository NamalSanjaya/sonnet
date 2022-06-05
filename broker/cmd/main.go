package main

import (
	"context"
	"os"
	"os/signal"
	"syscall"
	"time"

	lg "github.com/labstack/gommon/log"

	histbrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/historytable"
	redisrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/redis_cache"
	core "github.com/NamalSanjaya/sonnet/broker/pkg/service"
	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	ms "github.com/NamalSanjaya/sonnet/pkgs/database/mssql"
)

const memCleanerInterval time.Duration = time.Second * 50
const maxSpaceSize int = 2048    // 2kB

func main(){
	ctx, cancel := context.WithCancel(context.Background())
	redisCfg := &redis.Config{
		Host: "localhost",
		Port: 6379,
		PassWord: "--",
		DB: 0,
	}
	dbCfg := &ms.Config{
		Schema: "sqlserver", Hostname: "localhost", Database: "---",
		Username: "--", Password: "---",Port: 1433,
	}
	stopSig := make(chan os.Signal, 1)
	signal.Notify(stopSig, syscall.SIGINT, syscall.SIGTERM)
	queue  := store.NewQueue()
	ticker := time.NewTimer(memCleanerInterval)

	redisClient  := redis.NewClient(redisCfg)
	redisRepo    := redisrepo.NewRepo(redisClient)

	histRepo := histbrepo.NewRepo(ctx, dbCfg)
    logger   := lg.New("sonnet-broker")
	logger.EnableColor()
	broker   := core.New(redisRepo, queue, histRepo, logger)
	jobCtrlChan := make(chan int)
	shutDwnChan := make(chan int)
	go broker.JobExec(ctx, jobCtrlChan, shutDwnChan)

	go func () {
		logger.Info("starts redis-memory cleaner")
		for {
			select {
			case <- ctx.Done():
				logger.Info("exit from redis-memory cleaner")
				return
			case <- ticker.C:
				histTbList, err := broker.Cache.ListHistTbs(ctx)
				if err != nil {
					logger.Errorf("error listing history tables to find space exceeded history tables, due to %s", err.Error())
				} else {
					for _, histTb := range histTbList {
						metadata, err2 := broker.Cache.GetAllMetadata(ctx, histTb) // define a new method to get only `memsize`
						if err2 != nil {
							logger.Error(err2)
							continue
						}
						if metadata.MemSize > maxSpaceSize {
							isLenZero := broker.Queue.Len() == 0
							broker.Queue.Enqueue(histTb)
							if broker.Queue.Len() == 1 && isLenZero {
								jobCtrlChan <- 1
							}
						}
					}
				}
			}
		}
	}()
	<- stopSig
	cancel()
	time.Sleep(time.Second * 2)
	<- shutDwnChan
	time.Sleep(time.Second * 2)
	logger.Info("Broker gracefully shutting down")
}
