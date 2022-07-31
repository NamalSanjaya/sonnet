package ds1_handler

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
)

type Interface interface {
	InsertMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	AddBlockUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	AddNewContact(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
	RemoveBlockUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse
}
