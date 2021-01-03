package converter

import "io"

// StubConverterID is the Stub converter ID
const StubConverterID ID = "stub"

// Stub is the stub converter used for testing purposes
type Stub struct {
	StubID  ID
	Message *Message
	Err     error
}

// ID implements the Converter interface. It returns a stub ID if provided
// or returns the default StubConverterID otherwise
func (s *Stub) ID() ID {
	if s.StubID != "" {
		return ID(s.StubID)
	}
	return StubConverterID
}

// Convert implements the Converter interface. It returns the stub
// message and error.
func (s *Stub) Convert(data io.ReadSeeker) (*Message, error) {
	return s.Message, s.Err
}
