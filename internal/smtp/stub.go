package smtp

import (
	"context"

	"github.com/eexit/http2smtp/internal/converter"
)

// Stub is a test struct that implements Client
type Stub struct {
	SentCount int
	Err       error
}

// Send implements the Client.Send() method
func (s *Stub) Send(_ context.Context, _ *converter.Message) (int, error) {
	return s.SentCount, s.Err
}

// Close implements the Client.Close() method
func (s *Stub) Close() error {
	return nil
}
