package ds2_handler

import (
	"context"
	"fmt"
	"net/http"

	"github.com/julienschmidt/httprouter"

	trds2 "github.com/NamalSanjaya/sonnet/mserver/pkg/clients/transct_ds2"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	dsrc2 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
	rds "github.com/go-redis/redis/v8"
)

const mxRetry int = 10

type Handler struct {
	dataSrc2 dsrc2.Interface
}

func New(ds2Db dsrc2.Interface) *Handler {
	return &Handler{dataSrc2: ds2Db}
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

/* move last read value in ds2
----------- validations --------------
1. userid(`userid`), friend's history tb id(`tohist`) should be in uuid4 format
2. next lastread (`nxtread`) should be a numerical value
3. should not block user
4. verify userid has enough permission to move the last read in friend's history tb
    a. Friend HistTb[touserid] == userid
	b. lastread <= lastmsg && Newlastread >= preLastRead && lastread >= lastdeleted
*/
func (h *Handler) MoveLastRead(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	friendHistTb := r.URL.Query().Get("tohist")

	// userid(`userid`), friend's history tb id(`tohist`) should be in uuid4 format
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to move last-read of %s in ds2 due to invalied userid %s",
		friendHistTb, userId), hnd.FailedMvLastReadDS2, http.StatusBadRequest) 
	}
	if mdw.IsInvalidateUUID(friendHistTb) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to move last-read in ds2 due to invalied history id %s", friendHistTb),
		hnd.FailedMvLastReadDS2, http.StatusBadRequest) 
	}

	// check Friend's HistTb[touserid] == userid
	nxtLastRead, err := mdw.ToInt(r.URL.Query().Get("nxtread"))
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("failed to move last read of %s in ds2 due to invalid move number %s due to %w",
		friendHistTb, r.URL.Query().Get("nxtread"), err), hnd.FailedMvLastReadDS2, http.StatusBadRequest)
	}

	// begin tx with watch
	errCode := hnd.NoError
	statusCode := http.StatusCreated
	key := h.dataSrc2.MakeHistoryTbKey(friendHistTb)
	for i:=0; i< mxRetry; i++ {
		err = h.dataSrc2.Watch(ctx, func(tr trds2.Interface) error {
			metadata, err := tr.GetAllMetadata(ctx, friendHistTb)
			if err != nil {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusBadRequest
				return fmt.Errorf("failed to verify the access of userid %s to %s table due to %w",userId, friendHistTb, err)
			}
			if userId != metadata.UserId {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusUnauthorized
				return fmt.Errorf("denied accessing history tb %s, not a friend of user id %s",friendHistTb, userId)
			}
			if nxtLastRead < metadata.LastDeleted || nxtLastRead < metadata.LastRead ||  nxtLastRead > metadata.Lastmsg {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusBadRequest
				return fmt.Errorf("last read msg index is out of range of hist tb %s", friendHistTb)
			}
			return nil
		}, func(tr trds2.Interface) error {
			if err := tr.SetLastRead(ctx, friendHistTb, nxtLastRead); err != nil {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("failed to set lastread metadata of %s in ds2 by %s due to %w", friendHistTb, userId, err)
			}
			return nil
		}, key)
		if err == nil {
			errCode = hnd.NoError
			statusCode = http.StatusCreated
			break
		} else if err != rds.TxFailedErr {
			errCode = hnd.FailedMvLastReadDS2
			statusCode = http.StatusInternalServerError
			break
		}
		if i == mxRetry - 1 {
			errCode = hnd.FailedMvLastReadDS2
			statusCode = http.StatusInternalServerError
			err = fmt.Errorf("unable to lock the history tb %s to move lastread metadata of user id %s", 
			friendHistTb, userId)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}
 