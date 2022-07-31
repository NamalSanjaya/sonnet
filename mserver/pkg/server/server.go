package server

import (
	"net/http"

	"github.com/julienschmidt/httprouter"

	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type Server struct {
	h hnd.Interface
}

func New(handlers hnd.Interface) *Server{
	return &Server{
		h: handlers,
	}
}

// insert all metadata for user to ds1
func (s *Server) InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params){
	hres := s.h.InsertMetadataDS1(w,r,p)
	// TODO: add a error log, hres.Err contain the error/nil
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "" ,
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// add a new block user to ds1
func (s *Server) AddBlockUserToDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params){
	hres:= s.h.AddBlockUserToDS1(w, r, p)
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "" ,
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{ Err: hres.ErrCode })
}
