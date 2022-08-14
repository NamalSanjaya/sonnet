package server

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/julienschmidt/httprouter"
	lg "github.com/labstack/gommon/log"

	ds1hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers/data_source1"
	ds2hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers/data_source2"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

type Server struct {
	ds1h   ds1hnd.Interface
	ds2h   ds2hnd.Interface
	logger *lg.Logger
}

func New(ds1Handler ds1hnd.Interface, ds2Handler ds2hnd.Interface, logger *lg.Logger) *Server {
	return &Server{
		ds1h:   ds1Handler,
		ds2h:   ds2Handler,
		logger: logger,
	}
}

// insert all metadata for user to ds1
func (s *Server) InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds1h.InsertMetadata(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	// TODO: add a error log, hres.Err contain the error/nil
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// add a new block user to ds1
func (s *Server) AddBlockUserToDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds1h.AddBlockUser(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// add new contact to ds1
func (s *Server) AddNewContactToDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds1h.AddNewContact(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// remove block user from ds1
func (s *Server) RemoveBlockUserFromDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds1h.RemoveBlockUser(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// create history tables for both users in ds2
func (s *Server) AddNewContactToDS2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds2h.AddNewContact(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

// move lastread of history tb in ds2
func (s *Server) MoveLastReadInDS2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds2h.MoveLastRead(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

func (s *Server) UpdateLastMsgInDS2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds2h.UpdateLastMsg(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

func (s *Server) DeleteMsgInDS2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	hres := s.ds2h.DeleteMsg(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	// TODO: don't need a function to sendReponse. do it here. use map[string]interface{}{ Err:0, ... }
	mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
}

func (s *Server) LoadMsgsInDS2(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	data, hres := s.ds2h.LoadMsgs(w, r, p)
	if hres.Err != nil {
		s.logger.Errorf(hres.Err.Error())
	}
	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	content := map[string]interface{}{
		"Err":  hres.ErrCode,
		"msgs": data,
	}
	body, _ := json.Marshal(content)
	w.Write(body)
}

func (s *Server) LoadAllContactsMsgs(w http.ResponseWriter, r *http.Request, p httprouter.Params) {
	ctx := context.Background()
	histTbs, hres := s.ds1h.ListHistTbs(r)
	// TODO: if len(histTbs) == 0 return from here

	mdw.SetResponseHeaders(w, hres.StatusCode, map[string]string{
		ContentType: ApplicationJson_Utf8,
		Date: "",
	})
	if hres.Err != nil {
		mdw.SendResponse(w, &mdw.ResponseMsg{Err: hres.ErrCode})
		return
	}
	// TODO: if histTb map is empty, we need to resolve it from db.

	// Get a sorted toUserId slice(sort it based to `lastmsg` metadata in ds2. higher one first)
	body, err := s.ds2h.GetSortedHistTbContent(ctx, histTbs) // will change
	if err != nil {
		mdw.SendResponse(w, &mdw.ResponseMsg{Err: 100})
		return
	}
	w.Write(body)
}
