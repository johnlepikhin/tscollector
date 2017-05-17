package httpjson

import (
	"net/http"
	"encoding/json"
)

type JsonResponseOK struct {
	Status int
}

type JsonResponseError struct {
	Status int
	Message string
}

func HttpJsonResponse (w http.ResponseWriter, handler func(w *json.Encoder)) {
	w.Header().Set("Content-Type", "application/json")
	jsonw := json.NewEncoder(w)
	handler(jsonw)
}

func HttpJsonOK (w http.ResponseWriter) {
	HttpJsonResponse(w, func (jsonw *json.Encoder) {
		jsonw.Encode(JsonResponseOK{0})
	})
}

func HttpJsonError (w http.ResponseWriter, status int, message string) {
	HttpJsonResponse(w, func (jsonw *json.Encoder) {
		jsonw.Encode(JsonResponseError{status, message})
	})
}
