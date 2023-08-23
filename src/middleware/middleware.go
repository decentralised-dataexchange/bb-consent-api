package middleware

import (
	"fmt"
	"log"
	"net/http"
	"time"

	"github.com/bb-consent/api/src/apikey"
	"github.com/gorilla/mux"

	"github.com/bb-consent/api/src/common"
	"github.com/bb-consent/api/src/handler"
	"github.com/bb-consent/api/src/token"
	"github.com/bb-consent/api/src/user"
)

// Middleware Middleware function type declaration
type Middleware func(http.HandlerFunc) http.HandlerFunc

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

// Authenticate Validates the token and sets the token to the context.
func Authenticate() Middleware {

	// Create a new Middleware
	return func(f http.HandlerFunc) http.HandlerFunc {

		// Define the http.HandlerFunc
		return func(w http.ResponseWriter, r *http.Request) {
			authType, key, err := token.DecodeAuthHeader(r)
			if err != nil {
				log.Printf("Authorization header decoding failed: %v", err)
				m := fmt.Sprintf("Invalid authorization header, Authorization failed")
				common.HandleError(w, http.StatusUnauthorized, m, nil)
				return
			}

			if authType == token.AuthorizationToken {
				t, err := token.ParseToken(key)
				if err != nil {
					log.Printf("token decoding failed: %v", err)
					m := fmt.Sprintf("Invalid token, Authorization failed")
					common.HandleError(w, http.StatusUnauthorized, m, nil)
					return
				}
				token.Set(r, t)
				u, err := handler.GetUserByIamID(token.GetIamID(r))
				if err != nil {
					log.Printf("User not found err: %v", err)
					m := fmt.Sprintf("User does not exist, Authorization failed")
					common.HandleError(w, http.StatusUnauthorized, m, nil)
					return
				}
				token.SetUserID(r, u.ID.Hex())
				token.SetUserRoles(r, handler.GetUserRoles(u.Roles))
			}
			if authType == token.AuthorizationAPIKey {
				claims, err := apikey.Decode(key)
				if err != nil {
					log.Printf("api key decoding failed: %v", err)
					m := fmt.Sprintf("Invalid token, Authorization failed")
					common.HandleError(w, http.StatusUnauthorized, m, nil)
					return
				}

				if claims.Audience == "" {
					var u user.User

					if userID, ok := mux.Vars(r)["userID"]; ok {
						u, err = handler.GetUser(userID)
						if err != nil {
							log.Printf("User not found err: %v", err)
							m := fmt.Sprintf("Invalid API Key, Authorization failed")
							common.HandleError(w, http.StatusUnauthorized, m, nil)
							return
						}
					} else {
						u, err = handler.GetUser(claims.UserID)
						if err != nil {
							log.Printf("User not found err: %v", err)
							m := fmt.Sprintf("Invalid API Key, Authorization failed")
							common.HandleError(w, http.StatusUnauthorized, m, nil)
							return
						}
						if u.APIKey != key {
							log.Printf("API Key does not match")
							m := fmt.Sprintf("Invalid API Key, Authorization failed")
							common.HandleError(w, http.StatusUnauthorized, m, nil)
							return
						}
					}

					t := token.AccessToken{}
					t.IamID = u.IamID
					t.Name = u.Name
					t.Email = u.Email

					token.Set(r, t)
					token.SetUserID(r, u.ID.Hex())
					token.SetUserRoles(r, handler.GetUserRoles(u.Roles))
				}

			}
			// Call the next middleware/handler in chain
			f(w, r)
		}
	}
}

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

// Chain applies middlewares to a http.HandlerFunc
func Chain(f http.HandlerFunc, middlewares ...Middleware) http.HandlerFunc {
	for _, m := range middlewares {
		f = m(f)
	}
	return f
}
