package ds2_handler

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sort"
	"strconv"

	rds "github.com/go-redis/redis/v8"
	"github.com/julienschmidt/httprouter"

	trds2 "github.com/NamalSanjaya/sonnet/mserver/pkg/clients/transct_ds2"
	hnd "github.com/NamalSanjaya/sonnet/mserver/pkg/handlers"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
	dsrc2 "github.com/NamalSanjaya/sonnet/mserver/pkg/repository/data_source2"
	fmterr "github.com/NamalSanjaya/sonnet/pkgs/errors"
)

// tune theses values
const (
	mxRetry        int = 10
	maxLatestCount int = 5 // max no.of latest friend's chats get meduim size loaded☺•
)

const (
	minDel string = "MinDel"
	other  string = "Other"
)

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
	ctxInfo := []interface{}{"userId", userId, "newUserId", newUser}
	if userId == newUser {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("userId and new userId can't be same", "Falied to add new contact to owner's DS2",
			"", nil, ctxInfo...), hnd.FailedCreateNewUsrDS2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied userId", "Falied to add new contact to owner's DS2",
			"userId is not in correct uuid-v4 format", nil, ctxInfo...), hnd.FailedCreateNewUsrDS2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(newUser) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied newUserId", "Falied to add new contact to owner's DS2",
			"newUserId is not in correct uuid-v4 format", nil, ctxInfo...), hnd.FailedCreateNewUsrDS2, http.StatusBadRequest)
	}
	pairHistTbs, err := mdw.ReadHistTbJson(r)
	if err != nil {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to read request body", "Falied to add new contact to owner's DS2",
			"Falied to extract histroy tb info", err, ctxInfo...), hnd.FailedCreateNewUsrDS2, http.StatusInternalServerError)
	}
	// create history tbs for both users in ds2
	if err = h.dataSrc2.CreateHistTbs(ctx, userId, newUser, pairHistTbs); err != nil {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to create history tb in ds2", "Falied to add new contact to owner's DS2",
			"", err, ctxInfo...), hnd.SomeErrCreateNewUsrDS2, http.StatusInternalServerError)
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
	ctxInfo := []interface{}{"userId", userId, "friendHistoryTb", friendHistTb}

	// userid(`userid`), friend's history tb id(`tohist`) should be in uuid4 format
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid userId", "Falied to move last read metadata in DS2",
			"userId is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMvLastReadDS2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(friendHistTb) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid friend's userId", "Falied to move last read metadata in DS2",
			"friend's userId is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMvLastReadDS2, http.StatusBadRequest)
	}

	// check Friend's HistTb[touserid] == userid
	nxtLastRead, err := mdw.ToInt(r.URL.Query().Get("nxtread"))
	if err != nil {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("nxtread query parameter should be a numerical value", "Falied to move last read metadata in DS2",
			"next lastread value is not a number", err, ctxInfo...), hnd.FailedMvLastReadDS2, http.StatusBadRequest)
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
				return fmterr.NewFmtError("Unable to verify the access to friend's HistTb", "Falied to move last read metadata in DS2",
					"Falied to get metadata from friend's HistTb from DS2", err, append(ctxInfo, "Watch-Retry", i)...)
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmterr.NewFmtError("No history tb found", "Falied to move last read metadata in DS2", "", nil, append(ctxInfo, "Watch-Retry", i)...)
			}
			if userId != metadata.UserId {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusUnauthorized
				return fmterr.NewFmtError("Access denied to history table", "Falied to move last read metadata in DS2",
					"Friend's History tb toUserId is not equal to owner's UserId", nil, append(ctxInfo, "Watch-Retry", i)...)
			}
			if nxtLastRead < metadata.LastDeleted || nxtLastRead < metadata.LastRead || nxtLastRead > metadata.Lastmsg {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusBadRequest
				return fmterr.NewFmtError("nxtread is in wrong range", "Falied to move last read metadata in DS2", "", nil,
					append(ctxInfo, "nxtLastRead", nxtLastRead, "lastdeleted", metadata.LastDeleted, "lastread", metadata.LastRead, "lastmsg", metadata.Lastmsg, "Watch-Retry", i)...)
			}
			return nil
		}, func(tr trds2.Interface) error {
			if err := tr.SetLastRead(ctx, friendHistTb, nxtLastRead); err != nil {
				errCode = hnd.FailedMvLastReadDS2
				statusCode = http.StatusInternalServerError
				return fmterr.NewFmtError("Falied to set lastread in ds2", "Falied to move last read metadata in DS2", "", err, append(ctxInfo, "Watch-Retry", i)...)
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
			err = fmterr.NewFmtError("Unable to lock history tb", "Falied to move last read metadata in DS2", "Max limit of retrying is exceeded", nil, ctxInfo...)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}

func (h *Handler) UpdateLastMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	myHistTb := r.URL.Query().Get("hist")
	ctxInfo := []interface{}{"UserId", userId, "myHistTb", myHistTb}
	// userid(`userid`), my history tb id(`hist`) should be in uuid4 format
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid userId", "Falied to update last msg metadata in DS2",
			"userId is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(myHistTb) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid History tb", "Falied to update last msg metadata in DS2",
			"myHistTb is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
	}

	latestMsg, err := mdw.ToInt(r.URL.Query().Get("latestmsg"))
	if err != nil {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("latestmsg query param should be a numerical value", "Falied to update last msg metadata in DS2",
			"", err, append(ctxInfo, "lastestmsg", r.URL.Query().Get("latestmsg"))), hnd.FailedUpdateLastMsgDs2, http.StatusBadRequest)
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
				return fmterr.NewFmtError("Unable to verify the access to History tb", "Falied to update last msg metadata in DS2",
					"", err, append(ctxInfo, "Watch-Retry", i)...)
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmterr.NewFmtError("No history table found", "Falied to update last msg metadata in DS2", "", nil, append(ctxInfo, "Watch-Retry", i)...)
			}
			if latestMsg < metadata.LastDeleted || latestMsg < metadata.LastRead || latestMsg < metadata.Lastmsg {
				statusCode = http.StatusBadRequest
				return fmterr.NewFmtError("latestMsg is in wrong range", "Falied to update last msg metadata in DS2", "", nil,
					append(ctxInfo, "latestMsg", latestMsg, "lastdeleted", metadata.LastDeleted, "lastread", metadata.LastRead, "lastmsg", metadata.Lastmsg, "Watch-Retry", i)...)
			}
			return nil
		}, func(tr trds2.Interface) error {
			if err := tr.SetLastMsg(ctx, myHistTb, latestMsg); err != nil {
				statusCode = http.StatusInternalServerError
				return fmterr.NewFmtError("Falied to set lastmsg in history tb in ds2", "Falied to update last msg metadata in DS2", "",
					err, append(ctxInfo, "Watch-Retry", i)...)
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
			err = fmterr.NewFmtError("Unable to lock history tb", "Falied to update last msg metadata in DS2", "Max limit of retrying is exceeded", nil, ctxInfo...)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}

func (h *Handler) DeleteMsg(w http.ResponseWriter, r *http.Request, p httprouter.Params) *hnd.HandlerResponse {
	ctx := context.Background()
	userId := p.ByName("userId")
	myHistTb := r.URL.Query().Get("hist")
	ctxInfo := []interface{}{"userId", userId, "myHistTb", myHistTb}
	if mdw.IsInvalidateUUID(userId) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid UserId", "Falied to delete given msg from memory", "userId is not in correct uuid-v4 format",
			nil, ctxInfo...), hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(myHistTb) {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalid History table", "Falied to delete given msg from memory",
			"History table is not in correct uuid-v4 format", nil, ctxInfo...), hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}

	delMsg, err := mdw.ToInt(r.URL.Query().Get("delmsg")) // timestamp of msg
	if err != nil {
		return hnd.MakeHandlerResponse(fmterr.NewFmtError("delmsg should be a numerical value", "Falied to delete given msg from memory",
			"", err, append(ctxInfo, "delmsg", r.URL.Query().Get("delmsg"))), hnd.FailedDeleteMsgDs2, http.StatusBadRequest)
	}
	ctxInfo = append(ctxInfo, "delMsg", delMsg)

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
				return fmterr.NewFmtError("Unable to verify access to history tb", "Falied to delete given msg from memory",
					"Falied to get metadata from ds2", err, append(ctxInfo, "Watch-Retry", i))
			}
			if metadata == nil {
				statusCode = http.StatusNotFound
				return fmterr.NewFmtError("No history tb found", "Falied to delete given msg from memory", "", nil, append(ctxInfo, "Watch-Retry", i))
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
					return fmterr.NewFmtError("Unable to get adjacent timestamp to lastread", "Falied to delete given msg from memory", "",
						err, append(ctxInfo, "lastdeleted", metadata.LastDeleted, "lastread", metadata.LastRead, "Watch-Retry", i))
				}
			}
			if delMsg == metadata.Lastmsg {
				// change lastmsg
				// get adjacent memory timestamp [lastdeleted, lastmsg)
				unWatch = false
				if newLastMsg, err = tr.GetAdjacentTimeStamp(ctx, myHistTb, metadata.LastDeleted, metadata.Lastmsg); err != nil {
					statusCode = http.StatusInternalServerError
					return fmterr.NewFmtError("Failed to get adjacent timestamp to lastmsg", "Falied to delete given msg from memory", "",
						err, append(ctxInfo, "lastdeleted", metadata.LastDeleted, "lastmsg", metadata.Lastmsg, "Watch-Retry", i))
				}
			}
			var rowSize int
			if rowSize, err = tr.GetMemRowSize(ctx, myHistTb, delMsg); err != nil {
				statusCode = http.StatusInternalServerError
				return fmterr.NewFmtError("Unable to get memory row size", "Falied to delete given msg from memory", "", err, append(ctxInfo, "Watch-Retry", i))
			}
			newMemSize = newMemSize - rowSize
			if newMemSize < 0 {
				statusCode = http.StatusInternalServerError
				return fmterr.NewFmtError("Memory row size for given timestamp or ds2 total memory size is in wrong state", "Falied to delete given msg from memory",
					"calculated total memory is a negative value", nil, append(ctxInfo, "memsizeInDS2", newMemSize+rowSize, "rowSize", rowSize, "Watch-Retry", i))
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
			err = fmterr.NewFmtError("Unable to lock history tb", "Falied to delete given msg from memory", "Max limit of retrying is exceeded", nil, ctxInfo...)
		}
	}
	return hnd.MakeHandlerResponse(err, errCode, statusCode)
}

// TODO: msg data should be encrypted.
// TODO: msg data should be stored and store in our caches and databases.
func (h *Handler) LoadMsgs(w http.ResponseWriter, r *http.Request, p httprouter.Params) (dsrc2.MemoryRows, *hnd.HandlerResponse) {
	var msgRows dsrc2.MemoryRows
	ctx := context.Background()
	userId := r.URL.Query().Get("userid")
	toUserId := r.URL.Query().Get("touserid")
	myHistTb := r.URL.Query().Get("hist")
	toHistTb := r.URL.Query().Get("tohist")
	ctxInfo := []interface{}{"userId", userId, "toUserId", toUserId, "myHistTb", myHistTb, "toUserHistTb", toHistTb}
	if mdw.IsInvalidateUUID(userId) {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied userId", "Falied to load msgs for a given range",
			"userId is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(toUserId) {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied toUserId", "Falied to load msgs for a given range",
			"toUserId is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(myHistTb) {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Owner's history table is invalid", "Falied to load msgs for a given range",
			"Owner's history table is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}
	if mdw.IsInvalidateUUID(toHistTb) {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Friend's history table is invalid", "Falied to load msgs for a given range",
			"Friend's history table is not in the correct uuid-v4 format", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}

	start, err := strconv.Atoi(r.URL.Query().Get("start")) // `start` timestamp
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied start timestamp", "Falied to load msgs for a given range",
			"start query parameter should be a numerical value", err, append(ctxInfo, "start", r.URL.Query().Get("start"))...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}
	end, err := strconv.Atoi(r.URL.Query().Get("end")) // `end` timestamp
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Invalied end timestamp", "Falied to load msgs for a given range",
			"end query parameter should be a numerical value", err, append(ctxInfo, "end", r.URL.Query().Get("end"))...), hnd.FailedMsgsLoad, http.StatusBadRequest)
	}
	ctxInfo = append(ctxInfo, "start", start, "end", end)

	// check the permission
	var ok bool
	ok, err = h.dataSrc2.IsSameToUser(ctx, userId, toHistTb)
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to check owner's permission to access toHistTb", "Falied to load msgs for a given range",
			"", err, ctxInfo...), hnd.FailedMsgsLoad, http.StatusInternalServerError)
	}
	if !ok {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Owner does n't have permission to access toHistTb", "Falied to load msgs for a given range",
			"toUserId of toHistTb is not equal to owner's Id", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusUnauthorized)
	}
	ok, err = h.dataSrc2.IsSameToUser(ctx, toUserId, myHistTb)
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to check friend's permission to access myHistTb", "Falied to load msgs for a given range",
			"", err, ctxInfo...), hnd.FailedMsgsLoad, http.StatusInternalServerError)
	}
	if !ok {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("friend does n't have permission to access owner's history tb", "Falied to load msgs for a given range",
			"toUserId of myHistTb is not equal to friend's Id", nil, ctxInfo...), hnd.FailedMsgsLoad, http.StatusUnauthorized)
	}

	myMemRows, _, err := h.dataSrc2.ListMemoryRows(ctx, myHistTb, start, end)
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to list memory rows of myHistTb", "Falied to load msgs for a given range",
			"", err, ctxInfo...), hnd.FailedMsgsLoad, http.StatusInternalServerError)
	}
	for _, kk := range myMemRows {
		fmt.Println("just : ", *kk)
	}
	toUserMemRows, _, err := h.dataSrc2.ListMemoryRows(ctx, toHistTb, start, end)
	if err != nil {
		return msgRows, hnd.MakeHandlerResponse(fmterr.NewFmtError("Unable to list memory rows of toHistTb", "Falied to load msgs for a given range",
			"", err, ctxInfo...), hnd.FailedMsgsLoad, http.StatusInternalServerError)
	}
	return h.dataSrc2.CombineHistTbs(myMemRows, toUserMemRows), hnd.MakeHandlerResponse(nil, hnd.NoError, http.StatusOK)
}

func (h *Handler) GetSortedHistTbContent(ctx context.Context, histMp map[string]*mdw.PairHistTb) ([]byte, error) {
	var sortedInfo []*hnd.SmPairHistInfo
	var err error
	userIdScore := map[string]int{}
	for id, pairHistTb := range histMp {
		var txLink, rxLink *dsrc2.HistTbMetadata
		txLink, err = h.dataSrc2.GetAllMetadata(ctx, pairHistTb.Tx2Rx_HistTb)
		if err != nil || txLink == nil {
			// TODO: add logger to log error here. don't pass it to upper layers, specially server layer.
			continue
		}
		rxLink, err = h.dataSrc2.GetAllMetadata(ctx, pairHistTb.Rx2Tx_HistTb)
		if err != nil || rxLink == nil {
			continue
		}
		userIdScore[id] = mdw.Max(txLink.Lastmsg, rxLink.Lastmsg)
		txProp := minDel
		rxProp := other
		if txLink.LastDeleted >= rxLink.LastDeleted {
			txProp = other
			rxProp = minDel
		}
		sortedInfo = append(sortedInfo, &hnd.SmPairHistInfo{UserId: id,
			TxLink: &hnd.Info{Metadata: txLink, HistId: pairHistTb.Tx2Rx_HistTb, Prop: txProp},
			RxLink: &hnd.Info{Metadata: rxLink, HistId: pairHistTb.Rx2Tx_HistTb, Prop: rxProp}})
	}
	sort.Slice(sortedInfo, func(i, j int) bool {
		return userIdScore[sortedInfo[i].UserId] > userIdScore[sortedInfo[j].UserId]
	})

	var latestCount int
	result := map[string]interface{}{}
	for _, elemt := range sortedInfo {
		var contentTx, contentRx dsrc2.MemoryRows
		var sizeTx, sizeRx int
		otherHist := ""
		lastTimestamp := 0
		// (exist unread msgs) OR (within first set of chats)
		if elemt.RxLink.Metadata.LastRead != elemt.RxLink.Metadata.Lastmsg || maxLatestCount > latestCount {
			if elemt.TxLink.Prop == minDel {
				otherHist = elemt.RxLink.HistId
				lastTimestamp = elemt.TxLink.Metadata.LastDeleted
			} else {
				// RxLink == minDel
				otherHist = elemt.TxLink.HistId
				lastTimestamp = elemt.RxLink.Metadata.LastDeleted
			}
			if contentTx, sizeTx, err = h.dataSrc2.ListMemoryRows(ctx, elemt.TxLink.HistId, elemt.TxLink.Metadata.LastDeleted,
				elemt.TxLink.Metadata.Lastmsg); err != nil {
				continue
			}
			if contentRx, sizeRx, err = h.dataSrc2.ListMemoryRows(ctx, elemt.RxLink.HistId, elemt.RxLink.Metadata.LastDeleted,
				elemt.RxLink.Metadata.Lastmsg); err != nil {
				continue
			}
		}
		latestCount++
		result[elemt.UserId] = map[string]interface{}{
			"tx2rx": map[string]interface{}{
				"histId":  elemt.TxLink.HistId,
				"lastMsg": elemt.TxLink.Metadata.Lastmsg,
				"size":    sizeTx,
				"content": contentTx,
			},
			"rx2tx": map[string]interface{}{
				"histId":   elemt.RxLink.HistId,
				"lastMsg":  elemt.RxLink.Metadata.Lastmsg,
				"lastRead": elemt.RxLink.Metadata.LastRead,
				"size":     sizeRx,
				"content":  contentRx,
			},
			"userId":        elemt.UserId,
			"latestUpdated": mdw.Max(elemt.TxLink.Metadata.Lastmsg, elemt.RxLink.Metadata.Lastmsg),
			"dbFetchInfo":   map[string]interface{}{"link": otherHist, "lastTimestamp": lastTimestamp},
		}

	}
	result["Err"] = 0 // no error found
	return json.Marshal(result)
}
