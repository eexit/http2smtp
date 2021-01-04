package api

import (
	"net/http"

	"github.com/eexit/http2smtp/internal/api/handler"
	"github.com/gorilla/mux"
)

// routeHandler returns the app routes
func (a *API) routeHandler() http.Handler {
	r := mux.NewRouter()

	r.Handle("/healthcheck", handler.Healthcheck(Version)).
		Methods(http.MethodHead, http.MethodGet)

	r.Handle("/sparkpost/api/v1/transmissions", handler.SparkPost(a.smtpClient, a.converterProvider)).
		Methods(http.MethodPost)

	return r
}
