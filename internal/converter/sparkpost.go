package converter

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

// SparkPostTransmission represents a SparkPost transmission
// See: https://developers.sparkpost.com/api/transmissions/#transmissions-create-a-transmission
type SparkPostTransmission struct {
	Recipients []struct {
		Address struct {
			Email string `json:"email" validate:"required,email"`
		} `json:"address"`
	} `json:"recipients" validate:"required,min=1,dive,required"`
	Content struct {
		EmailRFC822 string `json:"email_rfc822" validate:"required"`
	} `json:"content" validate:"required"`
}

type spt10n struct {
	rfc5322Converter Converter
	validator        *validator.Validate
}

// NewSparkPostTransmission returns a new SparkPost transmission converter
func NewSparkPostTransmission() Converter {
	return &spt10n{
		rfc5322Converter: NewRFC5322(),
		validator:        validator.New(),
	}
}

func (s *spt10n) Convert(data io.ReadSeeker) (*Message, error) {
	t10n := &SparkPostTransmission{}

	if err := json.NewDecoder(data).Decode(t10n); err != nil {
		return nil, err
	}

	if err := s.validator.Struct(t10n); err != nil {
		return nil, err
	}

	if t10n.Content.EmailRFC822 != "" {
		return s.rfc822ToMessage(t10n)
	}

	return nil, errors.New("inline content transmission not implemented")
}

func (s *spt10n) rfc822ToMessage(t10n *SparkPostTransmission) (*Message, error) {
	raw := strings.NewReader(t10n.Content.EmailRFC822)

	// First, we need to parse the raw email to get the from address
	messageFromRFC822, err := s.rfc5322Converter.Convert(raw)
	if err != nil {
		return nil, err
	}

	// The recipient list is provided as it is in the request payload,
	// we don't parse the raw email because Bcc header should be missing.
	rcpts := []string{}
	for _, to := range t10n.Recipients {
		rcpts = append(rcpts, to.Address.Email)
	}

	return NewMessage(
		messageFromRFC822.From(),
		rcpts,
		nil,
		nil,
		raw,
	), nil
}
