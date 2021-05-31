package pkg

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/eexit/http2smtp/pkg/testutil"
)

func Test_SlurpBody(t *testing.T) {
	tests := []struct {
		name    string
		req     *http.Request
		want    string
		wantErr bool
	}{
		{
			name:    "no body",
			req:     httptest.NewRequest("", "/", nil),
			want:    "",
			wantErr: false,
		},
		{
			name:    "test body",
			req:     httptest.NewRequest("", "/", strings.NewReader("test body")),
			want:    "test body",
			wantErr: false,
		},
		{
			name:    "body read fails",
			req:     httptest.NewRequest("", "/", &testutil.FailingReader{}),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := SlurpBody(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("SlurpBody() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				return
			}

			gotAsBytes, err := ioutil.ReadAll(got)
			if err != nil {
				t.Fatalf("got read failed: %v", err)
			}

			gotAsStr := string(gotAsBytes)

			if gotAsStr != tt.want {
				t.Errorf("SlurpBody() = %#v, want %#v", gotAsStr, tt.want)
			}

			// Ensures the request body is still there after the slurp
			body, err := ioutil.ReadAll(tt.req.Body)
			if err != nil {
				t.Fatalf("request body read failed: %v", err)
			}

			bodyAsStr := string(body)
			if bodyAsStr != tt.want {
				t.Errorf("request body = %#v, want %#v", bodyAsStr, tt.want)
			}
		})
	}
}
