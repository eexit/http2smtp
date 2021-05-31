package api

import (
	"net/http"

	"github.com/eexit/http2smtp/internal/api/handler"
	"github.com/gorilla/mux"
)

// Mux returns the app routes
func (a *API) Mux() http.Handler {
	r := mux.NewRouter()

	r.Handle("/healthcheck", handler.Healthcheck(Version)).
		Methods(http.MethodHead, http.MethodGet)

	r.Handle("/sparkpost/api/v1/transmissions", handler.SparkPost(a.smtpClient, a.converterProvider)).
		Methods(http.MethodPost)

	r.Handle("/mailgun/api/v3/messages.mime", handler.Mailgun(a.smtpClient, a.converterProvider)).
		Methods(http.MethodPost)

	return r
}
