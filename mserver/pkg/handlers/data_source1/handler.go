package ds1_handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	dsrc1 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source1"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
)

type Handler struct {
	dataSrc1 dsrc1.Interface
}

func New(ds1Cache dsrc1.Interface) *Handler{
	return &Handler{
		dataSrc1: ds1Cache,
	}
}

func (h *Handler) InsertMetadata(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("invalid user id %s tries to insert metadata to ds1", userId), 
		hnd.FailedSetDS1, http.StatusBadRequest)
	}
	data, err := mdw.ReadDS1Json(r)
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("unable to read request body with user id %s due to %w", userId, err), 
		hnd.FailedSetDS1, http.StatusBadRequest)
	}
	if err = h.dataSrc1.SetDs1Metadata(ctx, userId, &dsrc1.DS1Metadata{Info: *data, State: 1}); err != nil {
		return hnd.MakeHandlerResponse(err, hnd.FailedPartiallySetDS1, http.StatusInternalServerError)
	}
	return hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusCreated)
}

func (h *Handler) AddBlockUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	blockUsr := r.URL.Query().Get("userid")
	if userId == blockUsr {
		return hnd.MakeHandlerResponse(fmt.Errorf("userid and blockuserid can't be same with user id %s", userId),
		hnd.FailedAddBlockUsrDS1, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("invalid user id %s tries add blockUser metadata to ds1", userId),
		hnd.FailedAddBlockUsrDS1, http.StatusBadRequest) 
	}
	if mdw.IsInvalidateUUID(blockUsr) {
		return hnd.MakeHandlerResponse(fmt.Errorf("invalid user id %s can't add to blockUser metadata in ds1", blockUsr),
		hnd.FailedAddBlockUsrDS1, http.StatusBadRequest)
	}
	if err := h.dataSrc1.AddBlockUser(ctx, userId, blockUsr); err != nil {
		return hnd.MakeHandlerResponse(err, hnd.FailedAddBlockUsrDS1, http.StatusInternalServerError)
	}
	return hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusCreated)
}

func (h *Handler) AddNewContact(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	newUserId := r.URL.Query().Get("userid")
	if userId == newUserId {
		return hnd.MakeHandlerResponse(fmt.Errorf("userid and new contact userid can't be same with user id %s", userId),
		hnd.FailedCreateNewUsrDS1, http.StatusBadRequest)
	}
	data, err := mdw.ReadHistTbJson(r)
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("unable to read reqeust body for userid %s due to %w", userId, err),
		hnd.FailedCreateNewUsrDS1, http.StatusBadRequest)
	}
	if err = h.dataSrc1.CreateNewContact(ctx, userId, newUserId, data); err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("unable to create new contact for userid %s with new-userid %s due to %w", 
		userId, newUserId, err), hnd.FailedCreateNewUsrDS1, http.StatusInternalServerError)
	}
	return hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusCreated)
}

func (h *Handler) RemoveBlockUser(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	rmBlockUser := r.URL.Query().Get("userid")
	if userId == rmBlockUser {
		return hnd.MakeHandlerResponse(fmt.Errorf("userid and remove blocked userid can't be same with user id %s", userId),
		hnd.FailedRmBlockUserDs1, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("invalid user id %s tries add blockUser metadata to ds1", userId),
		hnd.FailedRmBlockUserDs1, http.StatusBadRequest) 
	}
	if mdw.IsInvalidateUUID(rmBlockUser) {
		return hnd.MakeHandlerResponse(fmt.Errorf("invalid user id %s can't add to blockUser metadata in ds1", rmBlockUser),
		hnd.FailedRmBlockUserDs1, http.StatusBadRequest)
	}
	if err := h.dataSrc1.RemoveBlockUser(ctx, userId, rmBlockUser); err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("failed to remove block-userid %s from userid %s in ds1", rmBlockUser, userId),
		hnd.FailedRmBlockUserDs1, http.StatusInternalServerError)
	}
	return hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusOK)
}
