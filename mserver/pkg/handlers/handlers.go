package handlers

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	redisrepo "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/redis_cache"
)

type Handlers struct {
	redisCache redisrepo.Interface
	// userProfileRepo string
	// histRepo string
	// blockUserRepo string
}

func New(cache redisrepo.Interface) *Handlers{
	return &Handlers{
		redisCache: cache,
	}
}

func (h *Handlers) InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	if !mdw.ValidateUUID(userId) {
		return MakeHandlerResponse(fmt.Errorf("invalid user id %s tries to insert metadata to ds1", userId), 
		1, http.StatusBadRequest)
	}
	data, err := mdw.ReadDS1Json(r)
	if err != nil {
		return MakeHandlerResponse(fmt.Errorf("could read request body with user id %s due to %w", userId, err), 
		1, http.StatusBadRequest)
	}
	if err = h.redisCache.SetDs1Metadata(ctx, userId, &redisrepo.DS1Metadata{Info: *data, State: 1}); err != nil {
		return MakeHandlerResponse(err, 2, http.StatusInternalServerError)
	}
	return MakeHandlerResponse(nil, 0, http.StatusCreated)
}

func (h *Handlers) AddBlockUserToDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) *HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	blockUsr := r.URL.Query().Get("userid")
	if userId == blockUsr {
		return MakeHandlerResponse(fmt.Errorf("userid and blockuserid can't be same with user id %s", userId),
		3, http.StatusBadRequest)
	}
	if !mdw.ValidateUUID(userId) {
		return MakeHandlerResponse(fmt.Errorf("invalid user id %s tries add blockUser metadata to ds1", userId),
		3, http.StatusBadRequest) 
	}
	if !mdw.ValidateUUID(blockUsr) {
		return MakeHandlerResponse(fmt.Errorf("invalid user id %s can't add to blockUser metadata in ds1", blockUsr),
		3, http.StatusBadRequest)
	}
	if err := h.redisCache.AddBlockUser(ctx, userId, blockUsr); err != nil {
		return MakeHandlerResponse(err, 3, http.StatusInternalServerError)
	}
	return MakeHandlerResponse(nil, 0, http.StatusCreated)
}
