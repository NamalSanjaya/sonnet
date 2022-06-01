package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/NamalSanjaya/sonnet/broker/pkg/store"
	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	core "github.com/NamalSanjaya/sonnet/broker/pkg/service"
)

func main(){
	config := &redis.Config{
		Host: "localhost",
		Port: 6379,
		PassWord: "",
		DB: 0,
	}
	queue := store.NewQueue()
	cache := redis.NewClient(config)
	router := httprouter.New()
	broker := core.New(cache, queue)
	jobCtrlChan := make(chan int)
	go broker.JobExec(jobCtrlChan)

	router.GET("/ms-b/queue", func(w http.ResponseWriter, r *http.Request, _ httprouter.Params) {
		preLenZero := false
		if broker.Queue.IsEmpty() {
			preLenZero = true
		}
		broker.Queue.Enqueue(store.Task{
			TimeStamp: 1,
			Type: "Store in DB",
			OwnerId: "u-11",
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
