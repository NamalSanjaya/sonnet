package middleware

import (
	"net/http"
	"encoding/json"
)

func SetResponseHeaders(w http.ResponseWriter ,statusCode int, headers map[string]string){
	for key, val := range headers{
		w.Header().Add(key, val)
	}
	w.WriteHeader(statusCode)
}

func SendResponse(w http.ResponseWriter, content *ResponseMsg){
	var body []byte
	body, _ = json.Marshal(content)
	w.Write(body)
}
