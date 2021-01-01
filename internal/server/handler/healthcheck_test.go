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
		mux := http.NewServeMux()
		mux.HandleFunc("/healthcheck", Healthcheck("v0.1.0-test"))
		api := httptest.NewServer(mux)
		defer api.Close()

		resp, err := http.Get(api.URL + "/healthcheck")
		if err != nil {
			t.Errorf("could not request /healthcheck: %v", err)
		}
		defer resp.Body.Close()

		if code := resp.StatusCode; code != http.StatusOK {
			t.Errorf("Healthcheck() returned status code %v, want %v", code, http.StatusOK)
		}

		body, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			t.Errorf("could not read response body: %v", err)
		}

		want := `{"version":"v0.1.0-test"}`
		if got := strings.TrimSpace(string(body)); got != want {
			t.Errorf("Healthcheck() = %#v, want %#v", got, want)
		}
	})
}
