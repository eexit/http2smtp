package converter

import (
	"fmt"
	"net/http"
	"net/mail"
	"strings"
)

// RFC5322ID is the ID for RFC5322 converter
const RFC5322ID ID = "rfc5322"

type rfc5322 struct{}

// NewRFC5322 returns a new message converter for RFC 5322 format
func NewRFC5322() Converter {
	return &rfc5322{}
}

func (rfc *rfc5322) ID() ID {
	return RFC5322ID
}

func (rfc *rfc5322) Convert(r *http.Request) (*Message, error) {
	body, err := readBody(r)
	if err != nil {
		return nil, err
	}

	m, err := mail.ReadMessage(body)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	(body.Seek(0, 0))

	return NewMessage(
		m.Header.Get("From"),
		parse(m.Header, "To"),
		parse(m.Header, "Cc"),
		parse(m.Header, "Bcc"),
		body,
	), nil
}

func parse(h mail.Header, key string) []string {
	var list []string
	hv := strings.TrimSpace(h.Get(key))

	if len(hv) == 0 {
		return list
	}

	for _, v := range strings.Split(hv, ",") {
		list = append(list, strings.TrimSpace(v))
	}
	return list
}
