package ds2_handler 

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
	
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
)

type Interface interface{
	AddNewContact(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	MoveLastRead(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	UpdateLastMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	DeleteMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
}
