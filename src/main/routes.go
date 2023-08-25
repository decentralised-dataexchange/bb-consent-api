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

	r.Handle("/v1/organizations", m.Chain(handler.AddOrganization, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}", m.Chain(handler.GetOrganizationByID, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/organizations/{organizationID}", m.Chain(handler.UpdateOrganization, m.Logger(), m.Authenticate())).Methods("PATCH")
	r.Handle("/v1/organizations/{organizationID}/coverimage", m.Chain(handler.UpdateOrganizationCoverImage, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/logoimage", m.Chain(handler.UpdateOrganizationLogoImage, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/image/{imageID}", m.Chain(handler.GetOrganizationImage, m.Logger(), m.Authenticate())).Methods("GET")

	r.Handle("/v1/organizations/{organizationID}/eulaURL", m.Chain(handler.UpdateOrgEula, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/eulaURL", m.Chain(handler.DeleteOrgEula, m.Logger(), m.Authenticate())).Methods("DELETE")

	r.Handle("/v1/organizations/{organizationID}/admins", m.Chain(handler.AddOrgAdmin, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/organizations/{organizationID}/admins", m.Chain(handler.GetOrgAdmins, m.Logger(), m.Authenticate())).Methods("GET")

	//Login
	r.Handle("/v1/users/register", m.Chain(handler.RegisterUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/login", m.Chain(handler.LoginUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1.1/users/login", m.Chain(handler.LoginUserV11, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/validate/email", m.Chain(handler.ValidateUserEmail, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/validate/phone", m.Chain(handler.ValidatePhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/verify/phone", m.Chain(handler.VerifyPhoneNumber, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/verify/otp", m.Chain(handler.VerifyOtp, m.LoggerNoAuth())).Methods("POST")

	// Admin login
	r.Handle("/v1/users/admin/login", m.Chain(handler.LoginAdminUser, m.LoggerNoAuth())).Methods("POST")
	r.Handle("/v1/users/token", http.HandlerFunc(handler.GetToken)).Methods("POST")
	r.Handle("/v1/user/password/reset", m.Chain(handler.ResetPassword, m.Logger(), m.Authenticate())).Methods("PUT")
	r.Handle("/v1/user/password/forgot", m.Chain(handler.ForgotPassword, m.LoggerNoAuth())).Methods("PUT")
	r.Handle("/v1/users/logout", m.Chain(handler.LogoutUser, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/users/unregister", m.Chain(handler.UnregisterUser, m.Logger(), m.Authenticate())).Methods("POST")

	//user
	r.Handle("/v1/user", m.Chain(handler.GetCurrentUser, m.Logger(), m.Authenticate())).Methods("GET")
	r.Handle("/v1/user", m.Chain(handler.UpdateCurrentUser, m.Logger(), m.Authenticate())).Methods("PATCH")
	r.Handle("/v1/user/register/ios", m.Chain(handler.UserClientRegister, m.Logger(), m.Authenticate())).Methods("POST")
	r.Handle("/v1/user/register/android", m.Chain(handler.UserClientRegister, m.Logger(), m.Authenticate())).Methods("POST")

}
