package ds2_handler 

import (
	"context"
	"net/http"

	"github.com/julienschmidt/httprouter"
	
	dsrc2 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
)

type Interface interface{
	AddNewContact(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	MoveLastRead(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	UpdateLastMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	DeleteMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	LoadMsgs(w http.ResponseWriter, r *http.Request, p httprouter.Params) (dsrc2.MemoryRows,*hnd.HandlerResponse)
	GetSortedHistTbContent(ctx context.Context, histMp map[string]*mdw.PairHistTb) ([]byte, error)
}
