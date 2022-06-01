package redis_cache

import (
	"fmt"
	"strconv"

	"github.com/NamalSanjaya/sonnet/pkgs/cache/redis"

)

const PrefixDs2   string   = "ds2#"
const PrefixMem   string   = "mem#"

const lastread    string   = "lastread"
const lastdeleted string   = "lastdeleted"
const state       string   = "state"

type redisRepo struct{
	cmder redis.Interface
}

func NewRepo(cmder redis.Interface) *redisRepo {
	return &redisRepo{cmder: cmder}
}

// get `lastread` `lastdeleted` `state` metadata
func (r *redisRepo) GetMetadata(histTb string) (*HistTbMetadata, error){
	metadata := HistTbMetadata{}
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	rawdata, err := r.cmder.HMGet(histTbKey, lastread, lastdeleted, state)
	if err != nil {
		return nil, err
	}
	temp := []int{}
	for _,elemt := range rawdata {
		if elemt != nil {
			d, err := strconv.Atoi(elemt.(string))
			if err != nil {
				return &HistTbMetadata{}, err
			}
			temp = append(temp, d)
		} else {
			temp = append(temp, -1)
		}
	}
	metadata.LastRead = temp[0]
	metadata.LastDeleted = temp[1]
	metadata.State = temp[2]
	return &metadata, nil
}

// set `lastdeleted` metadata
func (r *redisRepo) SetMetadata(histTb string, lastDel string) error {
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(histTbKey, lastdeleted ,lastDel)
}

// `state` metadata, 0:lock  , 1: open
func (r *redisRepo) Lock(histTb string) error{
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(histTbKey, state , "0")
}

//`state` metadata, 0:lock  , 1: open
func (r *redisRepo) Unlock(histTb string) error{
	histTbKey := fmt.Sprintf("%s%s", PrefixDs2, histTb)
	return r.cmder.HSet(histTbKey, state , "1")
}
