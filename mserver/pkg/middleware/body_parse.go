package middleware

import (
	"encoding/json"
	"io/ioutil"
	"net/http"

	"github.com/google/uuid"
)

func ReadDS1Json(r *http.Request) (*DS1MetadataJson, error){
	var metadata *DS1MetadataJson
	b, err := ioutil.ReadAll(r.Body)
	defer r.Body.Close()
	if err != nil {
		return metadata, err
	}
	if err = json.Unmarshal(b, &metadata); err != nil {
		return metadata, err
	}
	return metadata, nil
}

func ValidateUUID(id string) bool {
	_, err := uuid.Parse(id) 
	return err == nil
}
