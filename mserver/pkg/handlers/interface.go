package handlers

import (
	"net/http"

	"github.com/julienschmidt/httprouter"
)

type Interface interface {
	InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HandlerResponse
	AddBlockUserToDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HandlerResponse
}

type HandlerResponse struct {
	Err error
	ErrCode int
	StatusCode int
}

func MakeHandlerResponse(err error, errCode, statusCode int) *HandlerResponse {
	return &HandlerResponse{
		Err: err, ErrCode: errCode, StatusCode: statusCode,
	}
}
