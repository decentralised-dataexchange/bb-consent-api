package error_handler

import (
	"encoding/json"
	"net/http"
)

type HttpError struct {
	Status  int    `json:"errorCode"`
	Message string `json:"errorDescription"`
}

func HandleExit(w http.ResponseWriter) {
	r := recover()
	if r != nil {
		if he, ok := r.(HttpError); ok {
			response, _ := json.Marshal(he)
			w.WriteHeader(he.Status)
			w.Header().Set("Content-Type", "application/json")
			w.Write(response)
		} else {
			panic(r)
		}
	}
}

func Exit(status int, message string) {
	panic(HttpError{Status: status, Message: message})
}
