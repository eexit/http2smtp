package testutil

import (
	"errors"
	"testing"
)

func TestFailingReader_Read(t *testing.T) {
	fr := &FailingReader{}
	if _, err := fr.Read([]byte{}); err == nil {
		t.Errorf("expected FailingReader.Read() to return an error")
	}
}

func TestStubWriteCloser_Write(t *testing.T) {
	type fields struct {
		WriteErr error
		CloseErr error
	}
	tests := []struct {
		name    string
		fields  fields
		p       []byte
		want    int
		wantErr bool
	}{
		{
			name:    "returns no error and length of written input",
			fields:  fields{},
			p:       []byte(`foo`),
			want:    3,
			wantErr: false,
		},
		{
			name:    "returns no error when there is a closing error",
			fields:  fields{CloseErr: errors.New("closing error")},
			p:       []byte(`foo`),
			want:    3,
			wantErr: false,
		},
		{
			name:    "returns error and length of written input",
			fields:  fields{WriteErr: errors.New("some error")},
			p:       []byte(`foo`),
			want:    3,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StubWriteCloser{
				WriteErr: tt.fields.WriteErr,
				CloseErr: tt.fields.CloseErr,
			}
			got, err := s.Write(tt.p)
			if (err != nil) != tt.wantErr {
				t.Errorf("StubWriteCloser.Write() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if got != tt.want {
				t.Errorf("StubWriteCloser.Write() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestStubWriteCloser_Close(t *testing.T) {
	type fields struct {
		WriteErr error
		CloseErr error
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		{
			name:    "returns no error",
			fields:  fields{},
			wantErr: false,
		},
		{
			name:    "returns no error when error is a writing error",
			fields:  fields{WriteErr: errors.New("write error")},
			wantErr: false,
		},
		{
			name:    "returns errors",
			fields:  fields{CloseErr: errors.New("closing error")},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &StubWriteCloser{
				WriteErr: tt.fields.WriteErr,
				CloseErr: tt.fields.CloseErr,
			}
			if err := s.Close(); (err != nil) != tt.wantErr {
				t.Errorf("StubWriteCloser.Close() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
