package converter

import (
	"io"
	"io/ioutil"
)

// RecipientProvider is the common func type for To(), Cc() and Bcc()
type RecipientProvider func(*Message) []string

// Message represents an email message
type Message struct {
	from        string
	to, cc, bcc []string
	raw         io.Reader
}

// NewMessage returns a new Message instance
func NewMessage(from string, to, cc, bcc []string, raw io.Reader) *Message {
	return &Message{
		from: from,
		to:   to,
		cc:   cc,
		bcc:  bcc,
		raw:  raw,
	}
}

// From returns the message's From: value
func (m *Message) From() string {
	return m.from
}

// To returns the To: recipient(s)
func (m *Message) To() []string {
	return m.to
}

// Cc returns the Cc: recipient(s)
func (m *Message) Cc() []string {
	return m.cc
}

// Bcc returns the Bcc: recipient(s)
func (m *Message) Bcc() []string {
	return m.bcc
}

// Raw returns the raw message (with its headers) as a byte stream
func (m *Message) Raw() ([]byte, error) {
	return ioutil.ReadAll(m.raw)
}

// HasRecipients returns true if the message contains as least one recipient
// amongst To, Cc and Bcc.
func (m *Message) HasRecipients() bool {
	return len(m.to)+len(m.cc)+len(m.bcc) > 0
}
