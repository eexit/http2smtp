package converter

import (
	"errors"
	"io"
	"io/ioutil"
	"reflect"
	"sort"
	"strings"
	"testing"
)

func TestNewSparkPost(t *testing.T) {
	t.Run("constructor returns a converter", func(t *testing.T) {
		want := &spt10n{
			rfc5322Converter: NewRFC5322(),
			validator:        val,
		}

		if got := NewSparkPost(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewSparkPost() = %+v, want %+v", got, want)
		}
	})
}

func Test_spt10n_ID(t *testing.T) {
	t.Run("ID() returns converter ID", func(t *testing.T) {
		s := &spt10n{}
		if got := s.ID(); got != SparkPostID {
			t.Errorf("ID() = %#v, want %#v", got, SparkPostID)
		}
	})
}

func Test_spt10n_Convert(t *testing.T) {
	tests := []struct {
		name    string
		data    io.ReadSeeker
		wantNil bool
		wantErr bool
	}{
		{
			name:    "data is not valid json",
			data:    strings.NewReader("<html></html>"),
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "json payload is not valid",
			data:    strings.NewReader(`{"foo":"bar"}`),
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "inline transmission is not supported",
			data:    strings.NewReader(`{"recipients":[{"address":{"email":"foo@example.com"}}],"content":{"email_rfc822":""}}`),
			wantNil: true,
			wantErr: true,
		},
		{
			name:    "RFC822 transmission is processed",
			data:    strings.NewReader(`{"recipients":[{"address":{"email":"foo@example.com"}}],"content":{"email_rfc822":"From: Test <test@example.com>\nTo: Bob <bob@example.com>\nSubject: Hello world!\n\nHello world!"}}`),
			wantNil: false,
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := NewSparkPost()
			got, err := s.Convert(tt.data)
			if (err != nil) != tt.wantErr {
				t.Errorf("spt10n.Convert() error = %v, wantErr %v", err, tt.wantErr)
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("spt10n.Convert() error = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}

func Test_spt10n_rfc822ToMessage(t *testing.T) {
	tests := []struct {
		name             string
		t10n             *SparkPostTransmission
		rfc5322Converter Converter
		want             *Message
		wantErr          bool
	}{
		{
			name:             "rfc5322 converter returns an error",
			t10n:             &SparkPostTransmission{},
			rfc5322Converter: &Stub{Err: errors.New("some error occurred")},
			want:             nil,
			wantErr:          true,
		},
		{
			name: "simple message",
			t10n: &SparkPostTransmission{
				Recipients: []Address{
					{
						AddressItem{Email: "recipient@example.com"},
					},
				},
				Content: Content{
					EmailRFC822: toString(simpleMessage),
				},
			},
			rfc5322Converter: &Stub{
				Message: &Message{from: "from@example.com"},
			},
			want: NewMessage(
				"from@example.com",
				[]string{"recipient@example.com"},
				nil,
				nil,
				simpleMessage,
			),
			wantErr: false,
		},
		{
			name: "only recipient from the payload are considered",
			t10n: &SparkPostTransmission{
				Recipients: []Address{
					{
						AddressItem{Email: "recipient1@example.com"},
					},
					{
						AddressItem{Email: "recipient2@example.com"},
					},
					{
						AddressItem{Email: "recipient3@example.com"},
					},
				},
				Content: Content{
					EmailRFC822: toString(messageWithCc),
				},
			},
			rfc5322Converter: &Stub{
				Message: &Message{from: "from@example.com"},
			},
			want: NewMessage(
				"from@example.com",
				[]string{"recipient1@example.com", "recipient2@example.com", "recipient3@example.com"},
				nil,
				nil,
				messageWithCc,
			),
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &spt10n{
				rfc5322Converter: tt.rfc5322Converter,
				validator:        val,
			}
			got, err := s.rfc822ToMessage(tt.t10n)
			if (err != nil) != tt.wantErr {
				t.Errorf("spt10n.rfc822ToMessage() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.want == nil {
				if got != nil {
					t.Errorf("spt10n.rfc822ToMessage() expected to return nil")
				}
				return
			}

			if got.From() != tt.want.From() {
				t.Errorf("spt10n.rfc822ToMessage() from = %#v, want %#v", got.From(), tt.want.From())
			}

			// Loops over all recipient list to assert them
			for _, provider := range []RecipientProvider{(*Message).To, (*Message).Cc, (*Message).Bcc} {
				gotList := provider(got)
				wantList := provider(tt.want)

				// Sorts the results for predictable result
				sort.Strings(gotList)
				sort.Strings(wantList)

				if !reflect.DeepEqual(gotList, wantList) {
					t.Errorf("spt10n.rfc822ToMessage() = %#v, want %#v", gotList, wantList)
				}
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
				t.Errorf("Raw() = %#v, want %#v", gotRawString, wantRawString)
			}
		})
	}
}

func toString(i io.ReadSeeker) string {
	s, _ := ioutil.ReadAll(i)
	(i.Seek(0, 0)) // rewind because the same content will be read another time
	return string(s)
}
