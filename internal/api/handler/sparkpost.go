package handler

import (
	"encoding/json"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/smtp"
)

const spIDLenght = 10000000000000000

type results struct {
	ID                      string `json:"id"`
	TotalAcceptedRecipients int    `json:"total_accepted_recipients"`
	TotalRejectedRecipients int    `json:"total_rejected_recipients"`
}

// SparkPost handles SparkPost transmission API calls
func SparkPost(smtpClient smtp.Client, converterProvider converter.Provider) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		converter, err := converterProvider.Get(converter.SparkPostID)
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

		sentCount, err := smtpClient.Send(r.Context(), message)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			(json.NewEncoder(w).Encode(map[string]string{"error": err.Error()}))
			return
		}

		w.WriteHeader(http.StatusCreated)
		(json.NewEncoder(w).Encode(struct {
			Results results `json:"results"`
		}{
			Results: results{
				TotalAcceptedRecipients: sentCount,
				ID:                      strconv.Itoa(rand.Intn(spIDLenght)),
			},
		}))
	}
}
