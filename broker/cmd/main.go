package main

import (
	"context"
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

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
	queue  := store.NewQueue()

	redisClient  := redis.NewClient(redisCfg)
	redisRepo    := redisrepo.NewRepo(redisClient)

	histRepo := histbrepo.NewRepo(ctx, dbCfg)
    logger   := lg.New("sonnet-broker")
	logger.EnableColor()
	broker   := core.New(redisRepo, queue, histRepo, logger)
	jobCtrlChan := make(chan int)
	go broker.JobExec(ctx, jobCtrlChan)

	router.PUT("/ms-b/queue", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		w.Header().Set("Date","")
		w.Header().Set("Content-Length","0")

		if err := r.ParseForm(); err != nil {
			logger.Error("unable to parse request body to get history name")
			w.WriteHeader(http.StatusInternalServerError)
			return
		}
		histTbList := r.PostForm["historyTable"]
		if len(histTbList) != 1{
			logger.Info("Bad incoming request without historyTable field")
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		preLenZero := false
		if broker.Queue.IsEmpty() {
			preLenZero = true
		}
		broker.Queue.Enqueue(histTbList[0])
		if broker.Queue.Len() == 1 && preLenZero {
			jobCtrlChan <- 1
		}
		w.WriteHeader(http.StatusOK)
	})
	fmt.Println("listening....localhost:8000")
	log.Fatal(http.ListenAndServe("localhost:8000", router))
}
