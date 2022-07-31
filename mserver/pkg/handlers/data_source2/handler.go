package ds2_handler 

import (
	"fmt"
	"net/http"
	"context"

	"github.com/julienschmidt/httprouter"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	dsrc2 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
)

type Handler struct {
	dataSrc2 dsrc2.Interface
}

func New(ds2Db dsrc2.Interface) *Handler{
	return &Handler{
		dataSrc2: ds2Db,
	}
}

func (h *Handler) AddNewContact(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	newUser := r.URL.Query().Get("userid")
	if userId == newUser {
		return hnd.MakeHandlerResponse(fmt.Errorf("userid and new-userid can't be same with user id %s", userId),
		hnd.FailedCreateNewUsrDS2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to create new contact in ds2 due to invalied userid %s", userId),
		hnd.FailedCreateNewUsrDS2, http.StatusBadRequest) 
	}
	if mdw.IsInvalidateUUID(newUser) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to create new contact with %s due to invalied new-user id %s", userId, newUser),
		hnd.FailedCreateNewUsrDS2, http.StatusBadRequest)
	}
	pairHistTbs, err := mdw.ReadHistTbJson(r)
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("unable to read HistTb Info from request body due to %w", err), 
		hnd.FailedCreateNewUsrDS2, http.StatusInternalServerError)
	}
	// create history tbs for both users in ds2
	if err = h.dataSrc2.CreateHistTbs(ctx, userId, newUser, pairHistTbs); err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("unable to create Hist tb for owner %s with new friend %s due to %w",
		userId, newUser, err), hnd.SomeErrCreateNewUsrDS2, http.StatusInternalServerError)
	}
	return hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusCreated)
}
