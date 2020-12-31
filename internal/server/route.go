package server

import (
	"net/http"

	"github.com/eexit/httpsmtp/internal/server/handler"
	"github.com/gorilla/mux"
)

// routeHandler returns the app routes
func (s *Server) routeHandler() http.Handler {
	r := mux.NewRouter()

	r.Handle("/healthcheck", handler.Healthcheck()).
		Methods(http.MethodHead, http.MethodGet)

	r.Handle("/sparkpost/api/v1/transmissions", handler.SparkPost(s.smtpClient)).
		Methods(http.MethodPost)

	return r
}
