package main

import (
	"expvar"
	"net/http"

	"github.com/AfflerZero/DeafChat/internal/signal"
	"github.com/julienschmidt/httprouter"
)

func (app *application) routes(hub *signal.Hub) http.Handler {
	router := httprouter.New()

	router.NotFound = http.HandlerFunc(app.notFoundResponse)
	router.MethodNotAllowed = http.HandlerFunc(app.methodNotAllowedResponse)

	router.HandlerFunc(http.MethodGet, "/v1/healthcheck", app.healthcheckHandler)

	router.HandlerFunc(http.MethodPost, "/v1/users", app.registerUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/activated", app.activateUserHandler)
	router.HandlerFunc(http.MethodPut, "/v1/users/password", app.updateUserPasswordHandler)

	router.HandlerFunc(http.MethodPost, "/v1/tokens/authentication", app.createAuthenticationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/activation", app.createActivationTokenHandler)
	router.HandlerFunc(http.MethodPost, "/v1/tokens/password-reset", app.createPasswordResetTokenHandler)

	if app.config.env == "development" {
		router.Handler(http.MethodGet, "/debug/vars", expvar.Handler())
	}

	restHandler := app.securityHeaders(app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.authenticate(router))))))

	signalHandler := app.securityHeaders(app.metrics(app.recoverPanic(app.enableCORS(app.rateLimit(app.signalHandler(hub))))))

	mux := http.NewServeMux()
	mux.Handle("/v1/signal", signalHandler)
	mux.Handle("/", restHandler)

	return mux
}
