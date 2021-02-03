package handler

import (
	"bytes"
	"errors"
	"io"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"regexp"
	"strings"
	"testing"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/smtp"
)

func TestSparkPost(t *testing.T) {
	type args struct {
		smtpClient        smtp.Client
		converterProvider converter.Provider
		requestBody       io.Reader
	}
	tests := []struct {
		name     string
		args     args
		wantCode int
		wantBody string
	}{
		{
			name: "no converter for this route",
			args: args{
				converterProvider: converter.NewProvider(),
				requestBody:       bytes.NewReader(nil),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"converter ID sparkpost not found"}`,
		},
		{
			name: "conversion failed",
			args: args{
				converterProvider: converter.NewProvider(&converter.Stub{
					StubID: converter.SparkPostID,
					Err:    errors.New("conversion failed"),
				}),
				requestBody: bytes.NewReader([]byte{}),
			},
			wantCode: http.StatusBadRequest,
			wantBody: `{"error":"conversion failed"}`,
		},
		{
			name: "send error",
			args: args{
				converterProvider: converter.NewProvider(&converter.Stub{StubID: converter.SparkPostID}),
				smtpClient: &smtp.Stub{
					SentCount: 0,
					Err:       errors.New("smtp error"),
				},
				requestBody: bytes.NewReader([]byte{}),
			},
			wantCode: http.StatusInternalServerError,
			wantBody: `{"error":"smtp error"}`,
		},
		{
			name: "send ok",
			args: args{
				converterProvider: converter.NewProvider(&converter.Stub{StubID: converter.SparkPostID}),
				smtpClient:        &smtp.Stub{SentCount: 42},
				requestBody:       bytes.NewReader([]byte{}),
			},
			wantCode: http.StatusCreated,
			wantBody: `{"results":{"id":"id","total_accepted_recipients":42,"total_rejected_recipients":0}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			handler := SparkPost(tt.args.smtpClient, tt.args.converterProvider)

			w := httptest.NewRecorder()
			r := httptest.NewRequest(http.MethodPost, "/", tt.args.requestBody)

			handler(w, r)

			resp := w.Result()

			if c := resp.StatusCode; c != tt.wantCode {
				t.Errorf("SparkPost() code = %v, want %v", c, tt.wantCode)
			}

			rb, err := ioutil.ReadAll(resp.Body)
			if err != nil {
				t.Fatalf("could not read response body: %v", err)
			}
			defer resp.Body.Close()

			body := strings.TrimSpace(string(rb))

			// Replace the random ID number by "id"
			m := regexp.MustCompile("\"id\":\"\\d+\"")
			body = m.ReplaceAllString(body, "\"id\":\"id\"")

			if body != tt.wantBody {
				t.Errorf("SparkPost() body = %#v, want %#v", body, tt.wantBody)
			}
		})
	}
}
