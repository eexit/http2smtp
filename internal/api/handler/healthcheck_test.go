package handler

import (
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
)

func TestHealthcheck(t *testing.T) {
	t.Run("route returns version with GET", func(t *testing.T) {
		handler := Healthcheck("v0.1.0-test")

		w := httptest.NewRecorder()
		r := httptest.NewRequest("GET", "/", nil)

		handler(w, r)

		if code := w.Code; code != http.StatusOK {
			t.Errorf("Healthcheck() returned status code %v, want %v", code, http.StatusOK)
		}

		body, err := ioutil.ReadAll(w.Body)
		if err != nil {
			t.Errorf("could not read response body: %v", err)
		}

		want := `{"version":"v0.1.0-test"}`
		if got := strings.TrimSpace(string(body)); got != want {
			t.Errorf("Healthcheck() = %#v, want %#v", got, want)
		}
	})
}
