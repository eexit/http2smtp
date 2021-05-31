package converter

import (
	"errors"
	"net/http"
	"strings"

	form "github.com/go-playground/form/v4"
	validator "github.com/go-playground/validator/v10"
)

const (
	// MailgunID is the ID for Mailgun converter
	MailgunID ID = "mailgun"
	// Mailgun handles up to 25 MB messages
	mgSizeLimit = 1 << 20 * 25
)

// MailgunMessage represents a Mailgun payload
// See: https://documentation.mailgun.com/en/latest/api-sending.html#sending
type MailgunMessage struct {
	To []string `form:"to" validate:"gt=0,dive,required"`
}

type mg struct {
	rfc5322Converter Converter
	validator        *validator.Validate
	decoder          *form.Decoder
}

// NewMailgun returns a new Mailgun message converter
func NewMailgun() Converter {
	return &mg{
		rfc5322Converter: NewRFC5322(),
		validator:        val,
		decoder:          decoder,
	}
}

func (m *mg) ID() ID {
	return MailgunID
}

func (m *mg) Convert(r *http.Request) (*Message, error) {
	if err := r.ParseMultipartForm(mgSizeLimit); err != nil {
		return nil, err
	}
	defer r.Body.Close()

	if !isMimeFormat(r) {
		return nil, errors.New("non-mime format not implemented")
	}

	msg := &MailgunMessage{}
	if err := m.decoder.Decode(msg, r.MultipartForm.Value); err != nil {
		return nil, err
	}

	if err := m.validator.Struct(msg); err != nil {
		return nil, err
	}

	return m.mimeRequestToMessage(msg.To, r)
}

// isMimeFormat is a quick helper to determine whether or not the given
// HTTP request is of email MIME format. This logic is specific to Mailgun
func isMimeFormat(r *http.Request) bool {
	return r.MultipartForm != nil && len(r.MultipartForm.File["message"]) > 0
}

// mimeRequestToMessage converts a Mailgun mime HTTP request to a Message
func (m *mg) mimeRequestToMessage(to []string, r *http.Request) (*Message, error) {
	file, err := r.MultipartForm.File["message"][0].Open()
	if err != nil {
		return nil, err
	}

	r2 := r.Clone(r.Context())
	r2.Body = file

	// We need to parse the raw email to get the from address
	messageFromRFC822, err := m.rfc5322Converter.Convert(r2)
	if err != nil {
		return nil, err
	}

	// Resets the reader because it has been read by the rfc5322Converter
	(file.Seek(0, 0))

	return NewMessage(
		messageFromRFC822.From(),
		flattenEmails(to),
		nil,
		nil,
		file,
	), nil
}

// flattenEmails flattens and parses CSV emails
func flattenEmails(input []string) []string {
	out := []string{}
	for _, email := range input {
		out = append(out, strings.Split(email, ",")...)
	}
	return out
}
