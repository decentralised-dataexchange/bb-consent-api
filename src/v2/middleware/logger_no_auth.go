package middleware

import (
	"log"
	"net/http"
	"time"
)

// LoggerNoAuth Logs API(s) that doesnt have tokens in the calls.
func LoggerNoAuth() Middleware {
	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			start := time.Now()
			defer func() { log.Println(r.URL.Path, time.Since(start)) }()

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
