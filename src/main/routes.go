package main

import (
	"net/http"

	"github.com/bb-consent/api/src/handler"
	m "github.com/bb-consent/api/src/middleware"
	"github.com/gorilla/mux"
)

// Root access return 200 OK for health check when the api
// is deployed in K8s with ingress controller.
func healthz(w http.ResponseWriter, r *http.Request) {
	w.WriteHeader(200)
}

// SetRoutes sets the routes that the back end server serves
func SetRoutes(r *mux.Router) {
	r.HandleFunc("/", healthz).Methods("GET")

	//Login
	r.Handle("/v1/users/register", m.Chain(handler.RegisterUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/login", m.Chain(handler.LoginUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1.1/users/login", m.Chain(handler.LoginUserV11, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/validate/email", m.Chain(handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")

	// Admin login
	r.Handle("/v1/users/admin/login", m.Chain(handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")

}
