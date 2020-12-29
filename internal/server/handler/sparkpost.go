package handler

import (
	"bytes"
	"encoding/json"
	"io/ioutil"
	"math/rand"
	"net/http"
	"strconv"

	"github.com/eexit/httpsmtp/internal/converter"
	"github.com/eexit/httpsmtp/internal/smtp"
	"github.com/rs/zerolog/hlog"
)

type results struct {
	ID                      string `json:"id"`
	TotalAcceptedRecipients int    `json:"total_accepted_recipients"`
	TotalRejectedRecipients int    `json:"total_rejected_recipients"`
}

// SparkPost handles SparkPost transmission API calls
func SparkPost(sender *smtp.SMTP) http.HandlerFunc {
	spCvtr := converter.NewSparkPostTransmission()

	return func(w http.ResponseWriter, r *http.Request) {
		logger := *(hlog.FromRequest(r))

		body, err := ioutil.ReadAll(r.Body)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}
		defer r.Body.Close()

		mail, err := spCvtr.Convert(bytes.NewReader(body))
		if err != nil {
			w.WriteHeader(400)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		sent, err := sender.WithLogger(logger).Send(r.Context(), mail)
		if err != nil {
			w.WriteHeader(500)
			json.NewEncoder(w).Encode(map[string]string{"error": err.Error()})
			return
		}

		w.WriteHeader(201)
		json.NewEncoder(w).Encode(struct {
			Results results `json:"results"`
		}{
			Results: results{
				TotalAcceptedRecipients: sent,
				ID:                      strconv.Itoa(rand.Intn(10000000000000000)),
			},
		})
	}
}
