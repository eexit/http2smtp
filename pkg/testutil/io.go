package testutil

import "errors"

// FailingReader implements io.Reader and fails at reading
type FailingReader struct{}

// Read implements io.Reader
func (*FailingReader) Read(p []byte) (int, error) {
	return 0, errors.New("read error")
}

// StubWriteCloser implements io.WriteCloser and is able to fail
// at Write, Close or both.
type StubWriteCloser struct {
	WriteErr, CloseErr error
}

// Write implements io.Writer
func (s *StubWriteCloser) Write(p []byte) (int, error) {
	return len(p), s.WriteErr
}

// Close implements io.Closer
func (s *StubWriteCloser) Close() error {
	return s.CloseErr
}
