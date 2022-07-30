package handlers

import (
	"context"
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

func (h *Handlers) InsertMetadataDS1(w http.ResponseWriter, r *http.Request, p httprouter.Params) error {
	ctx := context.Background()
	userId := p.ByName("userId")
	data, err := mdw.ReadDS1Json(r)
	if err != nil {
		return err
	}
	if err = h.redisCache.SetDs1Metadata(ctx, userId, &redisrepo.DS1Metadata{Info: *data, State: 1}); err != nil {
		return err
	}
	return nil
}
