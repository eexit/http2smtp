package handler

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/smtp"
)

// Mailgun handles Mailgun email transmissions
func Mailgun(smtpClient smtp.Client, converterProvider converter.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		converter, err := converterProvider.Get(converter.MailgunID)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			(json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}))
			return
		}

		message, err := converter.Convert(r)
		if err != nil {
			w.WriteHeader(http.StatusBadRequest)
			(json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}))
			return
		}

		_, err = smtpClient.Send(r.Context(), message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			(json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}))
			return
		}

		w.WriteHeader(http.StatusCreated)
		(json.NewEncoder(w).Encode(struct {
			Message string `json:"message"`
			ID      string `json:"id"`
		}{
			Message: "Queued. Thank you.",
			ID:      fmt.Sprintf("<%s@http2smtp>", time.Now().Format("20060102150405.9999.999")),
		}))
	}
}
