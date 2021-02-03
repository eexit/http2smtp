package converter

import (
	"io"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"testing"
)

func Test_NewRFC5322(t *testing.T) {
	t.Run("constructor returns a converter", func(t *testing.T) {
		want := &rfc5322{}

		if got := NewRFC5322(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewRFC5322() = %+v, want %+v", got, want)
		}
	})
}

func Test_rfc5322_Convert(t *testing.T) {
	tests := []struct {
		name    string
		reqBody io.ReadSeeker
		want    *Message
		wantErr bool
	}{
		{
			name:    "simple message",
			reqBody: strings.NewReader(simpleMessage),
			want: &Message{
				from: "Test <test@example.com>",
				to:   []string{"Bob <bob@example.com>"},
				raw:  strings.NewReader(simpleMessage),
			},
			wantErr: false,
		},
		{
			name:    "message with cc",
			reqBody: strings.NewReader(messageWithCc),
			want: &Message{
				from: "Test <test@example.com>",
				to:   []string{"Bob <bob@example.com>"},
				cc:   []string{"Alice <alice@example.com>", "bob@example.com"},
				raw:  strings.NewReader(messageWithCc),
			},
			wantErr: false,
		},
		{
			name:    "message with bcc",
			reqBody: strings.NewReader(messageWithBcc),
			want: &Message{
				from: "Test <test@example.com>",
				bcc:  []string{"Bob <bob@example.com>", "Alice <alice@example.com>"},
				raw:  strings.NewReader(messageWithBcc),
			},
			wantErr: false,
		},
		{
			name:    "message parsing error",
			reqBody: strings.NewReader(" From: Test <test@example.com>"),
			want:    nil,
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rfc := &rfc5322{}
			got, err := rfc.Convert(httptest.NewRequest(http.MethodPost, "/", tt.reqBody))
			if (err != nil) != tt.wantErr {
				t.Errorf("rfc5322.Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if tt.want == nil {
				return
			}

			if got.From() != tt.want.From() {
				t.Errorf("rfc5322.Convert.From() = %#v, want %#v", got.From(), tt.want.From())
			}
			if !reflect.DeepEqual(got.To(), tt.want.To()) {
				t.Errorf("rfc5322.Convert.To() = %#v, want %#v", got.To(), tt.want.To())
			}
			if !reflect.DeepEqual(got.Cc(), tt.want.Cc()) {
				t.Errorf("rfc5322.Convert.Cc() = %#v, want %#v", got.Cc(), tt.want.Cc())
			}
			if !reflect.DeepEqual(got.Bcc(), tt.want.Bcc()) {
				t.Errorf("rfc5322.Convert.Bcc() = %#v, want %#v", got.Bcc(), tt.want.Bcc())
			}

			gotRaw, err := got.Raw()
			if err != nil {
				t.Fatalf("got message raw read failed: %v", err)
			}
			wantRaw, err := tt.want.Raw()
			if err != nil {
				t.Fatalf("want message raw read failed: %v", err)
			}

			gotRawString := string(gotRaw)
			wantRawString := string(wantRaw)

			if gotRawString != wantRawString {
				t.Errorf("rfc5322.Convert.Raw() = %#v, want %#v", gotRawString, wantRawString)
			}
		})
	}
}

var simpleMessage = `From: Test <test@example.com>
To: Bob <bob@example.com>
Subject: Hello world!

Hello world!`

var messageWithCc = `From: Test <test@example.com>
To: Bob <bob@example.com>
Cc: Alice <alice@example.com>, bob@example.com
Subject: Hello world!

Hello world!`

var messageWithBcc = `From: Test <test@example.com>
Bcc: Bob <bob@example.com>,Alice <alice@example.com>
Subject: Hello world!

Hello world!`
