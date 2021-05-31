package pkg

import (
	"bytes"
	"io"
	"io/ioutil"
	"net/http"
)

// SlurpBody idempotently copies a request's body
func SlurpBody(r *http.Request) (io.ReadSeeker, error) {
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	r.Body = ioutil.NopCloser(bytes.NewBuffer(body))
	return bytes.NewReader(body), nil
}
