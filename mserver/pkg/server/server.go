package server

import (
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
)

type Server struct {
	h hnd.Interface
}

func New(handlers hnd.Interface) *Server{
	return &Server{
		h: handlers,
	}
}

func (s *Server) InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params){
	if err := s.h.InsertMetadataDS1(w,r,p); err != nil{
		fmt.Fprint(w, "failed namal")
		return
	}
	fmt.Fprint(w, "okay namal")
}
