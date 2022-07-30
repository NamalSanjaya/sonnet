package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) error
}