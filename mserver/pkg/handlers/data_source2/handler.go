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

const mxRetry int = 3

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
	for i := 0; i < mxRetry; i++ {
		err = h.dataSrc2.Watch(ctx, func(tr trds2.Interface) error {
			metadata, err := tr.GetAllMetadata(ctx, friendHistTb)
			if err != nil {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("failed to verify the access of userid %s to %s table to move lastread due to %w", userId, friendHistTb, err)
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmt.Errorf("no history table found under %s", friendHistTb)
			}
			if userId != metadata.UserId {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusUnauthorized
				return fmt.Errorf("denied accessing history tb %s to move lastread, not a friend of user id %s", friendHistTb, userId)
			}
			if nxtLastRead < metadata.LastDeleted || nxtLastRead < metadata.LastRead || nxtLastRead > metadata.Lastmsg {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusBadRequest
				return fmt.Errorf("falied to update lastmsg due to last read msg index is out of range of hist tb %s", friendHistTb)
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
			break
		}
		if i == mxRetry-1 {
			errCode = hnd.FailedMvLastReadDS2
			statusCode = http.StatusInternalServerError
			err = fmt.Errorf("max limit of retrying is exceeded. hence unable to lock the history tb %s to move lastread metadata of user id %s",
				friendHistTb, userId)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}

func (h *Handler) UpdateLastMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	myHistTb := r.URL.Query().Get("hist")

	// userid(`userid`), my history tb id(`hist`) should be in uuid4 format
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to update lastmsg of %s in ds2 due to invalied userid %s",
			myHistTb, userId), hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(myHistTb) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to update lastmsg in ds2 due to invalied history id %s", myHistTb),
			hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
	}

	latestMsg, err := mdw.ToInt(r.URL.Query().Get("latestmsg"))
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("failed to update lastmsg of %s in ds2 due to invalid lastmsg number %s due to %w",
			myHistTb, r.URL.Query().Get("latestmsg"), err), hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
	}

	// begin tx with watch
	errCode := hnd.FailedUpdateLastMsgDs2
	statusCode := http.StatusCreated
	key := h.dataSrc2.MakeHistoryTbKey(myHistTb)
	for i := 0; i < mxRetry; i++ {
		err = h.dataSrc2.Watch(ctx, func(tr trds2.Interface) error {
			metadata, err := tr.GetAllMetadata(ctx, myHistTb)
			if err != nil {
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("failed to verify the access of userid %s to %s table to update lastmsg due to %w", userId, myHistTb, err)
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmt.Errorf("no history table found under %s", myHistTb)
			}
			if latestMsg < metadata.LastDeleted || latestMsg < metadata.LastRead || latestMsg < metadata.Lastmsg {
				statusCode = http.StatusBadRequest
				return fmt.Errorf("falied to update lastmsg since lastmsg index is out of range of hist tb %s", myHistTb)
			}
			return nil
		}, func(tr trds2.Interface) error {
			if err := tr.SetLastMsg(ctx, myHistTb, latestMsg); err != nil {
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("failed to set lastmsg of %s in ds2 by %s due to %w", myHistTb, userId, err)
			}
			return nil
		}, key)

		if err == nil {
			errCode = hnd.NoError
			statusCode = http.StatusCreated
			break
		} else if err != rds.TxFailedErr {
			break
		}
		if i == mxRetry-1 {
			statusCode = http.StatusInternalServerError
			err = fmt.Errorf("max limit of retrying is exceeded. hence unable to lock the history tb %s to update lastmsg metadata of user id %s",
				myHistTb, userId)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}

func (h *Handler) DeleteMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	myHistTb := r.URL.Query().Get("hist")

	// userid(`userid`), my history tb id(`hist`) should be in uuid4 format
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to delete msg in %s in ds2 due to invalied userid %s",
			myHistTb, userId), hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(myHistTb) {
		return hnd.MakeHandlerResponse(fmt.Errorf("falied to delete msg in ds2 due to invalied history id %s", myHistTb),
			hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}

	delMsg, err := mdw.ToInt(r.URL.Query().Get("delmsg")) // timestamp of msg
	if err != nil {
		return hnd.MakeHandlerResponse(fmt.Errorf("failed to delete msg in %s of ds2 due to invalid delete msg index %s due to %w",
			myHistTb, r.URL.Query().Get("delmsg"), err), hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}

	// begin tx with watch - change ds2 metadata
	errCode := hnd.FailedDeleteMsgDs2
	statusCode := http.StatusCreated
	keyDs2 := h.dataSrc2.MakeHistoryTbKey(myHistTb)
	keyMem := h.dataSrc2.MakeHistMemKey(myHistTb)
	for i := 0; i < mxRetry; i++ {
		var newLastRead, newLastMsg, newMemSize int
		var unWatch bool = true
		err = h.dataSrc2.Watch(ctx, func(tr trds2.Interface) error {
			metadata, err := tr.GetAllMetadata(ctx, myHistTb)
			if err != nil {
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("failed to verify the access of userid %s to %s table to delete msg due to %w", userId, myHistTb, err)
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmt.Errorf("no history table found under %s", myHistTb)
			}
			newLastRead = metadata.LastRead
			newLastMsg = metadata.Lastmsg
			newMemSize = metadata.MemSize
			if delMsg <= metadata.LastDeleted || delMsg > metadata.Lastmsg {
				// nothing to change
				errCode = hnd.NoJobToDo
				statusCode = http.StatusOK
				return hnd.RequestIgnore{}
			}
			if delMsg == metadata.LastRead {
				// change lastread
				// get adjacent memory timestamp [lastdeleted, lastread)
				unWatch = false
				if newLastRead, err = tr.GetAdjacentTimeStamp(ctx, myHistTb, metadata.LastDeleted, metadata.LastRead); err != nil {
					statusCode = http.StatusInternalServerError
					return fmt.Errorf("failed to get adjacent timestamp to lastread %d due to %w", metadata.LastRead, err)
				}
			}
			if delMsg == metadata.Lastmsg {
				// change lastmsg
				// get adjacent memory timestamp [lastdeleted, lastmsg)
				unWatch = false
				if newLastMsg, err = tr.GetAdjacentTimeStamp(ctx, myHistTb, metadata.LastDeleted, metadata.Lastmsg); err != nil {
					statusCode = http.StatusInternalServerError
					return fmt.Errorf("failed to get adjacent timestamp to lastmsg %d due to %w", metadata.Lastmsg, err)
				}
			}
			var rowSize int
			if rowSize, err = tr.GetMemRowSize(ctx, myHistTb, delMsg); err != nil {
				statusCode = http.StatusInternalServerError
				return err
			}
			newMemSize = newMemSize - rowSize
			if newMemSize < 0 {
				statusCode = http.StatusInternalServerError
				return fmt.Errorf("memory row size %d or ds2 total memory size %d of histTb %s at %d timestamp is in a wrong state",
				rowSize, newMemSize+rowSize, myHistTb, delMsg)
			}
			return nil
		}, func(tr trds2.Interface) error {
			tr.RemMemRow(ctx, myHistTb, delMsg)
			tr.SetMemSize(ctx, myHistTb, newMemSize)
			if !unWatch {
				tr.SetLastMsg(ctx, myHistTb, newLastMsg)
				tr.SetLastRead(ctx, myHistTb, newLastRead)
			}
			return nil
		}, keyDs2, keyMem) // even though we are using mem#<histTb>#<timestamp> key, we don't need to watch it.

		if err == nil {
			errCode = hnd.NoError
			statusCode = http.StatusCreated
			break
		} else if err != rds.TxFailedErr {
			if hnd.IsRequestIgnore(err) { 
				err = nil 
			}
			break
		}
		if i == mxRetry-1 {
			statusCode = http.StatusInternalServerError
			err = fmt.Errorf("max limit of retrying is exceeded. hence unable to lock the history tb or memory %s to delete msg of user id %s", myHistTb, userId)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}
