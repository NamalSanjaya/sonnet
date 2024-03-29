package data_source1

import (
	"context"
	"fmt"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"
	mdw "github.com/NamalSanjaya/sonnet/mserver/pkg/middleware"
)

const ds1 string = "ds1#"

const username string = "username"
const email string = "email"
const tx2rx_HistTb string = "tx2rx"
const rx2tx_HistTb string = "rx2tx"

type redisRepo struct {
	cmder redis.Interface
}

var _ Interface = (*redisRepo)(nil)

func NewRepo(cmder redis.Interface) *redisRepo {
	return &redisRepo{cmder: cmder}
}

//  0:lock  , 1: open
func (r *redisRepo) SetDs1Metadata(ctx context.Context, userId string, metadata *DS1Metadata) error {
	var err, setErr error
	infoKey := makeDs1InfoKey(userId)
	if err = r.cmder.HSet(ctx, infoKey, username, metadata.Info.Username, email, metadata.Info.Email); err != nil {
		return err
	}
	blockListKey := makeDs1BlockUserListKey(userId)
	if err = r.cmder.SSet(ctx, blockListKey, metadata.Info.BlockUserList...); err != nil {
		return err
	}
	var errList []error
	var toUserIdList []string
	for toUserId, pairHistTb := range metadata.Info.HistTbs {
		histTbKey := makeDs1HistTbKey(userId, toUserId)
		if err = r.cmder.HSet(ctx, histTbKey, tx2rx_HistTb, pairHistTb.Tx2Rx_HistTb, rx2tx_HistTb, pairHistTb.Rx2Tx_HistTb); err != nil {
			errList = append(errList, err)
		}
		toUserIdList = append(toUserIdList, toUserId)
	}
	// TODO: remove previous list, if exist
	allHistTbsKey := makeDs1AllUserIdsKey(userId)
	if err = r.cmder.SSet(ctx, allHistTbsKey, toUserIdList...); err != nil {
		errList = append(errList, err)
	}
	if len(errList) > 0 {
		setErr = fmt.Errorf("errors occuried during ds1 caching process of %s and list of errors %v", userId, errList)
	}
	// TODO: this should return nil and error should log at here
	return setErr
}

func (r *redisRepo) AddBlockUser(ctx context.Context, userId, blockedUserId string) error {
	blockListKey := makeDs1BlockUserListKey(userId)
	return r.cmder.SAdd(ctx, blockListKey, blockedUserId)
}

func (r *redisRepo) CreateNewContact(ctx context.Context, userId, newUserId string, pairHistTb *mdw.PairHistTb) error {
	var err error
	allHistTbsKey := makeDs1AllUserIdsKey(userId)
	if err = r.cmder.SAdd(ctx, allHistTbsKey, newUserId); err != nil {
		return err
	}
	histTbKey := makeDs1HistTbKey(userId, newUserId)
	return r.cmder.HSet(ctx, histTbKey, tx2rx_HistTb, pairHistTb.Tx2Rx_HistTb, rx2tx_HistTb, pairHistTb.Rx2Tx_HistTb)
}

func (r *redisRepo) RemoveBlockUser(ctx context.Context, userId, rmUserId string) error {
	blockListKey := makeDs1BlockUserListKey(userId)
	return r.cmder.SRem(ctx, blockListKey, rmUserId)
}

func (r *redisRepo) ListHistTbs(ctx context.Context, userId string)(map[string]*mdw.PairHistTb, error){
	histTbs := map[string]*mdw.PairHistTb{}
	toUserIds, err := r.cmder.SMembers(ctx, makeDs1AllUserIdsKey(userId))
	if err != nil {
		return histTbs, err
	}
	var pairHistTb []string
	for _, id := range toUserIds {
		if pairHistTb, err = r.cmder.HMGet(ctx, makeDs1HistTbKey(userId, id), tx2rx_HistTb, rx2tx_HistTb); err != nil {
			// TODO: need to collect this error. log at here, don't pass it top(server) layer.
			continue
		}
		histTbs[id] = &mdw.PairHistTb{Tx2Rx_HistTb: pairHistTb[0], Rx2Tx_HistTb: pairHistTb[1]}
	}
	return histTbs, nil
}

func makeDs1InfoKey(usrId string) string {
	return fmt.Sprintf("%s%s", ds1, usrId)
}

func makeDs1BlockUserListKey(usrId string) string {
	return fmt.Sprintf("%sblockuserlist#%s", ds1, usrId)
}

func makeDs1AllUserIdsKey(usrId string) string {
	return fmt.Sprintf("%salltouserids#%s", ds1, usrId)
}

func makeDs1HistTbKey(usrId, toUsrId string) string {
	return fmt.Sprintf("%shisttb#%s#%s", ds1, usrId, toUsrId)
}
