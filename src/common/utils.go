package common

import (
	"encoding/json"
	"log"
	"net/http"
	"path/filepath"
	"runtime"
)

type status struct {
	Code    int
	Message string
}

// HandleError Common function to formulate error and set the status
func HandleError(w http.ResponseWriter, code int, message string, err error) {
	s := status{code, message}
	response, _ := json.Marshal(s)

	pc, fn, line, _ := runtime.Caller(1)

	log.Printf("%v with err:%v in %s[%s:%d]", message, err,
		filepath.Base(runtime.FuncForPC(pc).Name()), filepath.Base(fn), line)

	w.WriteHeader(code)
	w.Header().Set("Content-Type", "application/json")
	w.Write(response)
}
