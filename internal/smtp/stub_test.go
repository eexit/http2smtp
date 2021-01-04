package smtp

import (
	"context"
	"errors"
	"testing"

	"github.com/eexit/http2smtp/internal/converter"
)

func TestStub(t *testing.T) {
	s := &Stub{}

	sc, err := s.Send(context.Background(), &converter.Message{})
	if err != nil {
		t.Errorf("Send() err = %v, want nil", err)
	}
	if sc != 0 {
		t.Errorf("Send() sentCount = %v, want 0", sc)
	}

	wantSentCount := 42
	wantErr := errors.New("some error")

	s = &Stub{
		SentCount: wantSentCount,
		Err:       wantErr,
	}

	sc, err = s.Send(context.Background(), &converter.Message{})
	if err != wantErr {
		t.Errorf("Send() err = %v, want %v", err, wantErr)
	}
	if sc != wantSentCount {
		t.Errorf("Send() sentCount = %v, want %v", sc, wantSentCount)
	}

	if err := s.Close(); err != nil {
		t.Errorf("Close() = %v, want nil", err)
	}
}
