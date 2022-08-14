package main

import (
	"fmt"
	"log"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	msrv "github.com/NamalSanjaya/sonnet/mserver/pkg/server"
	ds1hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers/data_source1"
	ds2hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers/data_source2"
	dsrc1 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source1"
	dsrc2 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
)

func main()  {
	redisCfg := &redis.Config{
		Host:     "localhost",
		Port:     6379,
		PassWord: "",
		DB:       0,
	}
	logger := lg.New("sonnet-mserver")
	logger.EnableColor()

	redisClient := redis.NewClient(redisCfg)
	ds1Repo := dsrc1.NewRepo(redisClient)
	ds2Repo := dsrc2.NewRepo(redisClient)

	router := httprouter.New()
	ds1handler := ds1hnd.New(ds1Repo)
	ds2handler := ds2hnd.New(ds2Repo)
	srv    := msrv.New(ds1handler, ds2handler, logger)
	
	// PUT request - ds1
	router.PUT("/ms/set-ds1/:userId", srv.InsertMetadataDS1)
	router.PUT("/ms/set-blockuser/:userId", srv.AddBlockUserToDS1)
	router.PUT("/ms/set-newcontact-ds1/:userId", srv.AddNewContactToDS1)

	// DELETE request - ds1
	router.DELETE("/ms/del-blockuser/:userId", srv.RemoveBlockUserFromDS1)

	// PUT request - ds2
	router.PUT("/ms/set-newcontact-ds2/:userId", srv.AddNewContactToDS2)
	router.PUT("/ms/set-lastread/:userId", srv.MoveLastReadInDS2)
	router.PUT("/ms/set-lastmsg/:userId", srv.UpdateLastMsgInDS2)

	// DELETE request - ds2
	router.DELETE("/ms/del-msg/:userId", srv.DeleteMsgInDS2)

	// GET request - ds2
	router.GET("/ms/load-msgs", srv.LoadMsgsInDS2)
	router.GET("/ms/load-all-contactnmsgs", srv.LoadAllContactsMsgs)

	fmt.Println("Listen....8000")
	log.Fatal(http.ListenAndServe(":8000", router))
}
