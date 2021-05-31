package converter

import (
	"io"
	"reflect"
	"strings"
	"testing"

	"github.com/eexit/http2smtp/pkg/testutil"
)

func TestNewMessage(t *testing.T) {
	type args struct {
		from string
		to   []string
		cc   []string
		bcc  []string
		raw  io.Reader
	}
	tests := []struct {
		name string
		args args
		want *Message
	}{
		{
			name: "default values",
			args: args{},
			want: &Message{},
		},
		{
			name: "nil recipients",
			args: args{
				from: "from@example.com",
				to:   nil,
				cc:   nil,
				bcc:  nil,
			},
			want: &Message{from: "from@example.com"},
		},
		{
			name: "non-nil but empty recipients",
			args: args{
				from: "from@example.com",
				to:   []string{},
				cc:   []string{},
				bcc:  []string{},
			},
			want: &Message{
				from: "from@example.com",
				to:   []string{},
				cc:   []string{},
				bcc:  []string{},
			},
		},
		{
			name: "with raw email",
			args: args{raw: strings.NewReader("foo bar")},
			want: &Message{raw: strings.NewReader("foo bar")},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewMessage(tt.args.from, tt.args.to, tt.args.cc, tt.args.bcc, tt.args.raw); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewMessage() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func TestMessage_From(t *testing.T) {
	tests := []struct {
		name string
		from string
		want string
	}{
		{
			name: "empty from",
			from: "",
			want: "",
		},
		{
			name: "non-empty from",
			from: "from@example.com",
			want: "from@example.com",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				from: tt.from,
			}
			if got := m.From(); got != tt.want {
				t.Errorf("Message.From() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_Getters(t *testing.T) {
	type fields struct {
		to  []string
		cc  []string
		bcc []string
	}
	tests := []struct {
		name    string
		fields  fields
		wantTo  []string
		wantCc  []string
		wantBcc []string
	}{
		{
			name:    "nil fields",
			fields:  fields{},
			wantTo:  nil,
			wantCc:  nil,
			wantBcc: nil,
		},
		{
			name: "empty fields",
			fields: fields{
				to:  []string{},
				cc:  []string{},
				bcc: []string{},
			},
			wantTo:  []string{},
			wantCc:  []string{},
			wantBcc: []string{},
		},
		{
			name: "non-empty fields",
			fields: fields{
				to:  []string{"to@example.com"},
				cc:  []string{"cc@example.com"},
				bcc: []string{"bcc@example.com"},
			},
			wantTo:  []string{"to@example.com"},
			wantCc:  []string{"cc@example.com"},
			wantBcc: []string{"bcc@example.com"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				to:  tt.fields.to,
				cc:  tt.fields.cc,
				bcc: tt.fields.bcc,
			}
			if got := m.To(); !reflect.DeepEqual(got, tt.wantTo) {
				t.Errorf("To() = %v, want %v", got, tt.wantTo)
			}
			if got := m.Cc(); !reflect.DeepEqual(got, tt.wantCc) {
				t.Errorf("Cc() = %v, want %v", got, tt.wantCc)
			}
			if got := m.Bcc(); !reflect.DeepEqual(got, tt.wantBcc) {
				t.Errorf("Bcc() = %v, want %v", got, tt.wantBcc)
			}
		})
	}
}

func TestMessage_Raw(t *testing.T) {
	tests := []struct {
		name    string
		raw     io.Reader
		want    []byte
		wantErr bool
	}{
		{
			name:    "empty reader",
			raw:     &strings.Reader{},
			want:    []byte{},
			wantErr: false,
		},
		{
			name:    "reader with data",
			raw:     strings.NewReader("foo bar"),
			want:    []byte(`foo bar`),
			wantErr: false,
		},
		{
			name:    "read error",
			raw:     &testutil.FailingReader{},
			want:    []byte{},
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{raw: tt.raw}
			got, err := m.Raw()
			if (err != nil) != tt.wantErr {
				t.Errorf("Message.Raw() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Message.Raw() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestMessage_HasRecipients(t *testing.T) {
	type fields struct {
		to  []string
		cc  []string
		bcc []string
	}
	tests := []struct {
		name   string
		fields fields
		want   bool
	}{
		{
			name:   "all empty",
			fields: fields{},
			want:   false,
		},
		{
			name:   "to has 1 recipient",
			fields: fields{to: []string{"to@example.com"}},
			want:   true,
		},
		{
			name:   "cc has 1 recipient",
			fields: fields{cc: []string{"cc@example.com"}},
			want:   true,
		},
		{
			name:   "bcc has 1 recipient",
			fields: fields{bcc: []string{"bcc@example.com"}},
			want:   true,
		},
		{
			name: "to, cc, bcc have 2 recipients",
			fields: fields{
				to:  []string{"to1@example.com", "to2@example.com"},
				cc:  []string{"cc1@example.com", "cc2@example.com"},
				bcc: []string{"bcc1@example.com", "bcc2@example.com"},
			},
			want: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &Message{
				to:  tt.fields.to,
				cc:  tt.fields.cc,
				bcc: tt.fields.bcc,
			}
			if got := m.HasRecipients(); got != tt.want {
				t.Errorf("Message.HasRecipients() = %v, want %v", got, tt.want)
			}
		})
	}
}
