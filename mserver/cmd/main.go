package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	msrv "github.com/NamalSanjaya/sonnet/mserver/pkg/server"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
	redisrepo "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/redis_cache"
)

func main()  {
	redisCfg := &redis.Config{
		Host:     "localhost",
		Port:     6379,
		PassWord: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisCfg)
	redisRepo := redisrepo.NewRepo(redisClient)

	router := httprouter.New()
	handlers := hnd.New(redisRepo)
	srv    := msrv.New(handlers)
	
	// PUT request
	router.PUT("/ms/set-ds1/:userId", srv.InsertMetadataDS1)
	
	fmt.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
