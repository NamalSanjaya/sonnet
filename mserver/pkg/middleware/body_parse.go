package middleware

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"strconv"

	"github.com/google/uuid"
)

func ReadDS1Json(r *http.Request) (*DS1MetadataJson, error){
	var metadata DS1MetadataJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return &metadata, err
	}
	if err = json.Unmarshal(b, &metadata); err != nil {
		return &metadata, err
	}
	return &metadata, nil
}

func IsInvalidateUUID(id string) bool {
	_, err := uuid.Parse(id) 
	return err != nil
}

func ReadHistTbJson(r *http.Request)(*PairHistTb, error) {
	var pairHistTb PairHistTb
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return &pairHistTb, err
	}
	if err = json.Unmarshal(b, &pairHistTb); err != nil {
		return &pairHistTb, err
	}
	return &pairHistTb, nil
}

func ToInt(numStr string)(int, error){
	return strconv.Atoi(numStr)
}

func Max(val1, val2 int) int {
	if val1 >= val2 {
		return val1
	}
	return val2
}

// [ minValId, otherId ]
func MinDelIds(id1, id2 string, val1, val2 int)[]string{
	if val1 <= val2 {
		return []string{id1, id2}
	}
	return []string{id2, id1}
}
