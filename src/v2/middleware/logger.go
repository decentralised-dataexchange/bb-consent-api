package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/token"
)

// Logger logs all requests with its path and the time it took to process
func Logger() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			// Do middleware things
			start := time.Now()
			defer func() {
				log.Println("name:", token.GetUserName(r), "id:", token.GetUserID(r), time.Since(start), r.Method, r.URL.Path)
			}()

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
