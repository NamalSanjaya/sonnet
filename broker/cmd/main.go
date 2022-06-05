package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	histbrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/historytable"
	core "github.com/NamalSanjaya/sonnet/broker/pkg/service"
	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	ms "github.com/NamalSanjaya/sonnet/pkgs/database/mssql"
	redisrepo "github.com/NamalSanjaya/sonnet/broker/pkg/repository/redis_cache"
)

func main(){
	ctx := context.Background()
	redisCfg := &redis.Config{
		Host: "localhost",
		Port: 6379,
		PassWord: "",
		DB: 0,
	}
	dbCfg := &ms.Config{
		Schema: "sqlserver", Hostname: "localhost", Database: "sonnet",
		Username: "SA", Password: "Test#123#test",Port: 1433,
	}
	router := httprouter.New()
	queue := store.NewQueue()

	redisClient  := redis.NewClient(redisCfg)
	redisRepo := redisrepo.NewRepo(redisClient)

	histRepo := histbrepo.NewRepo(ctx, dbCfg)

	broker := core.New(redisRepo, queue, histRepo)
	jobCtrlChan := make(chan int)
	go broker.JobExec(ctx, jobCtrlChan)

	router.GET("/ms-b/queue", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		preLenZero := false
		if broker.Queue.IsEmpty() {
			preLenZero = true
		}
		broker.Queue.Enqueue(store.Task{
			TimeStamp: 1,
			Type: "Store in DB",
			OwnerHistTb: "af_hist1",
		})

		if broker.Queue.Len() == 1 && preLenZero {
			jobCtrlChan <- 1
		}
		fmt.Fprintf(w,"Queue length %v", broker.Queue.Len())
	})
	router.GET("/ms-b/queue/len", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		fmt.Fprintf(w,"current len, %v",broker.Queue.Len())
	})
	log.Fatal(http.ListenAndServe("localhost:8000", router))
}
