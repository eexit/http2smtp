package converter

import (
	"fmt"
	"io"
	"net/mail"
	"strings"
)

type rfc5322 struct{}

// NewRFC5322 returns a new message converter for RFC 5322 format
func NewRFC5322() Converter {
	return &rfc5322{}
}

func (rfc *rfc5322) Convert(data io.ReadSeeker) (*Message, error) {
	m, err := mail.ReadMessage(data)
	if err != nil {
		return nil, fmt.Errorf("failed to parse input: %w", err)
	}

	// Resets the reader
	(data.Seek(0, 0))

	return NewMessage(
		m.Header.Get("From"),
		parse(m.Header, "To"),
		parse(m.Header, "Cc"),
		parse(m.Header, "Bcc"),
		data,
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
