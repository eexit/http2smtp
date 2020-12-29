package handler

import (
	"encoding/json"
	"net/http"
)

// Healthcheck handles healthcheck route
func Healthcheck() http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		json.NewEncoder(w).Encode("I am alive")
	}
}
