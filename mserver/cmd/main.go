package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	msrv "github.com/NamalSanjaya/sonnet/mserver/pkg/server"
	ds1hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers/data_source1"
	dsrc1 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source1"
)

func main()  {
	redisCfg := &redis.Config{
		Host:     "localhost",
		Port:     6379,
		PassWord: "",
		DB:       0,
	}
	redisClient := redis.NewClient(redisCfg)
	ds1Repo := dsrc1.NewRepo(redisClient)

	router := httprouter.New()
	ds1handler := ds1hnd.New(ds1Repo)
	srv    := msrv.New(ds1handler)
	
	// PUT request - ds1
	router.PUT("/ms/set-ds1/:userId", srv.InsertMetadataDS1)
	router.PUT("/ms/set-blockuser/:userId", srv.AddBlockUserToDS1)
	router.PUT("/ms/set-newcontact-ds1/:userId", srv.AddNewContactToDS1)

	//DELETE request - ds1
	router.DELETE("/ms/del-blockuser/:userId", srv.RemoveBlockUserFromDS1)
	
	fmt.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
