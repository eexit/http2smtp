package api

import (
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/eexit/http2smtp/internal/converter"
	"github.com/eexit/http2smtp/internal/smtp"
)

func TestAPI_Mux(t *testing.T) {
	tests := []struct {
		name      string
		method    string
		routePath string
		wantCode  int
	}{
		{
			name:      "unknown route returns 404",
			method:    http.MethodGet,
			routePath: "/ghost",
			wantCode:  http.StatusNotFound,
		},
		{
			name:      "GET healthcheck route returns 200",
			method:    http.MethodGet,
			routePath: "/healthcheck",
			wantCode:  http.StatusOK,
		},
		{
			name:      "HEAD healthcheck route returns 200",
			method:    http.MethodHead,
			routePath: "/healthcheck",
			wantCode:  http.StatusOK,
		},
		{
			name:      "POST sparkpost transmission route returns 201",
			method:    http.MethodPost,
			routePath: "/sparkpost/api/v1/transmissions",
			wantCode:  http.StatusCreated,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &API{
				smtpClient:        &smtp.Stub{},
				converterProvider: converter.NewProvider(&converter.Stub{StubID: converter.SparkPostID}),
			}

			api := httptest.NewServer(s.Mux())
			defer api.Close()

			req, err := http.NewRequest(tt.method, api.URL+tt.routePath, nil)
			if err != nil {
				t.Fatalf("could not create request: %v", err)
			}

			client := &http.Client{Timeout: 1 * time.Second}

			resp, err := client.Do(req)
			if err != nil {
				t.Errorf("request failed: %v", err)
			}
			defer resp.Body.Close()

			if code := resp.StatusCode; code != tt.wantCode {
				t.Errorf("route %v returned status code %v, want %v", tt.routePath, code, tt.wantCode)
			}
		})
	}
}
