package converter

import (
	"bytes"
	"encoding/base64"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"reflect"
	"testing"

	"github.com/eexit/http2smtp/pkg"
)

// http --offline --verbose -f POST :8080/mailgun/api/v3/messages.mime to=foo@bar.com message@./examples/message-with-attachment.mime

func TestNewMailgun(t *testing.T) {
	t.Run("constructor returns a converter", func(t *testing.T) {
		want := &mg{
			rfc5322Converter: NewRFC5322(),
			validator:        val,
			decoder:          decoder,
		}

		if got := NewMailgun(); !reflect.DeepEqual(got, want) {
			t.Errorf("NewMailgun() = %+v, want %+v", got, want)
		}
	})
}

func Test_mg_ID(t *testing.T) {
	t.Run("ID() returns converter ID", func(t *testing.T) {
		m := &mg{}
		if got := m.ID(); got != MailgunID {
			t.Errorf("ID() = %#v, want %#v", got, MailgunID)
		}
	})
}

func Test_mg_Convert(t *testing.T) {
	tests := []struct {
		name         string
		req          *http.Request
		decoderError bool
		wantNil      bool
		wantErr      bool
	}{
		{
			name:         "body form parse error",
			req:          httptest.NewRequest(http.MethodPost, "/", nil),
			decoderError: false,
			wantNil:      true,
			wantErr:      true,
		},
		{
			name: "body is too large",
			req: func() *http.Request {
				data := `--test
Content-Disposition: form-data; name="to"

data
--test
Content-Disposition: form-data; name="message"; filename="example.txt"
Content-Type: text/plain
Content-Transfer-Encoding: Base64

`
				data += pkg.Chunk(base64.StdEncoding.EncodeToString(make([]byte, 1<<10)), 77, "\r\n")
				data += `
--test--`

				// fmt.Printf("\n\n%v\n\n", data)

				r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(data))
				r.Header.Set("content-type", "multipart/form-data; boundary=test")

				rr, err := httputil.DumpRequest(r, true)
				if err != nil {
					panic(err)
				}
				fmt.Printf("%v\n", string(rr))

				return r
			}(),
			decoderError: false,
			wantNil:      true,
			wantErr:      true,
		},
		// 		{
		// 			name: "decoder error",
		// 			reqBody: bytes.NewBufferString(`--test
		// Content-Disposition: form-data; name="to"

		// data
		// --test--`),
		// 			decoderError: true,
		// 			wantNil:      true,
		// 			wantErr:      true,
		// 		},
		// 		{
		// 			name: "non-mime payload format",
		// 			req: func() *http.Request {
		// 				r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`--test
		// Content-Disposition: form-data; name="foo"

		// data
		// --test--`))
		// 				r.Header.Set("content-type", "multipart/form-data; boundary=test")
		// 				return r
		// 			}(),
		// 			decoderError: false,
		// 			wantNil:      true,
		// 			wantErr:      true,
		// 		},
		// 		{
		// 			name: "field validation failure",
		// 			req: func() *http.Request {
		// 				r := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(`--test
		// Content-Disposition: form-data; name="foo"

		// data
		// --test
		// Content-Disposition: form-data; name="message"; filename="message.mime"
		// Content-Type: text/plain

		// data
		// --test--`))
		// 				r.Header.Set("content-type", "multipart/form-data; boundary=test")
		// 				return r
		// 			}(),
		// 			decoderError: false,
		// 			wantNil:      true,
		// 			wantErr:      true,
		// 		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			mg := &mg{
				rfc5322Converter: NewRFC5322(),
				validator:        val,
				decoder:          decoder,
			}

			if tt.decoderError {
				mg.decoder.RegisterCustomTypeFunc(func(vals []string) (interface{}, error) {
					return nil, errors.New("Bad Type Conversion")
				}, "")
			}

			got, err := NewMailgun().Convert(tt.req)
			t.Log(err)
			if (err != nil) != tt.wantErr {
				t.Errorf("mg.Convert() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if (got == nil) != tt.wantNil {
				t.Errorf("mg.Convert() error = %v, wantNil %v", got, tt.wantNil)
			}
		})
	}
}
