package middleware

import (
	"log"
	"net/http"
	"time"

	"github.com/bb-consent/api/internal/token"
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
				if token.GetUserID(r) != "" {
					log.Println("name:", token.GetUserName(r), "id:", token.GetUserID(r), time.Since(start), r.Method, r.URL.Path)
				} else {
					log.Println(time.Since(start), r.Method, r.URL.Path)
				}

			}()

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
