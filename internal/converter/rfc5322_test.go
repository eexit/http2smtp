package converter

import (
	"io"
	"reflect"
	"strings"
	"testing"
)

func Test_rfc5322_Convert(t *testing.T) {
	simpleMessage := strings.NewReader(`From: Test <test@example.com>
To: Bob <bob@example.com>
Subject: Hello world!

Hello world!`)

	messageWithCc := strings.NewReader(`From: Test <test@example.com>
To: Bob <bob@example.com>
Cc: Alice <alice@example.com>, bob@example.com
Subject: Hello world!

Hello world!`)

	messageWithBcc := strings.NewReader(`From: Test <test@example.com>
Bcc: Bob <bob@example.com>,Alice <alice@example.com>
Subject: Hello world!

Hello world!`)

	tests := []struct {
		name    string
		data    io.ReadSeeker
		want    *Message
		wantErr bool
	}{
		{
			name: "simple message",
			data: simpleMessage,
			want: &Message{
				from: "Test <test@example.com>",
				to:   []string{"Bob <bob@example.com>"},
				raw:  simpleMessage,
			},
			wantErr: false,
		},
		{
			name: "message with cc",
			data: messageWithCc,
			want: &Message{
				from: "Test <test@example.com>",
				to:   []string{"Bob <bob@example.com>"},
				cc:   []string{"Alice <alice@example.com>", "bob@example.com"},
				raw:  messageWithCc,
			},
			wantErr: false,
		},
		{
			name: "message with bcc",
			data: messageWithBcc,
			want: &Message{
				from: "Test <test@example.com>",
				bcc:  []string{"Bob <bob@example.com>", "Alice <alice@example.com>"},
				raw:  messageWithBcc,
			},
			wantErr: false,
		},
		{
			name:    "message parsing error",
			data:    strings.NewReader(" From: Test <test@example.com>"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc := &rfc5322{}
			got, err := rfc.Convert(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("rfc5322.Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("rfc5322.Convert() = %#v, want %#v", got, tt.want)
			}
		})
	}
}
