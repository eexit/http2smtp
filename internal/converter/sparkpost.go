package converter

import (
	"encoding/json"
	"errors"
	"io"
	"strings"

	validator "github.com/go-playground/validator/v10"
)

// SparkPostID is the ID for SparkPost converter
const SparkPostID ID = "sparkpost"

// SparkPostTransmission represents a SparkPost transmission
// See: https://developers.sparkpost.com/api/transmissions/#transmissions-create-a-transmission
type SparkPostTransmission struct {
	Recipients []Address `json:"recipients" validate:"required,min=1,dive,required"`
	Content    Content   `json:"content" validate:"required"`
}

// Address is a SparkPost address
type Address struct {
	AddressItem `json:"address"`
}

// AddressItem is a SparkPost Address item
type AddressItem struct {
	Email string `json:"email" validate:"required,email"`
}

// Content is the transmission content
type Content struct {
	EmailRFC822 string `json:"email_rfc822"`
}

var val = validator.New()

type spt10n struct {
	rfc5322Converter Converter
	validator        *validator.Validate
}

// NewSparkPost returns a new SparkPost transmission converter
func NewSparkPost() Converter {
	return &spt10n{
		rfc5322Converter: NewRFC5322(),
		validator:        val,
	}
}

func (s *spt10n) ID() ID {
	return SparkPostID
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
		rcpts = append(rcpts, to.Email)
	}

	return NewMessage(
		messageFromRFC822.From(),
		rcpts,
		nil,
		nil,
		raw,
	), nil
}
