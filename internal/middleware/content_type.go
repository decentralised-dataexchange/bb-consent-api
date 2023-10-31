package middleware

import (
	"net/http"

	"github.com/bb-consent/api/internal/config"
)

func AddContentType() Middleware {
	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {

			w.Header().Add(config.ContentTypeHeader, config.ContentTypeJSON)

			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}
