package server

import (
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"
	"time"

	"github.com/eexit/httpsmtp/internal/ctx"
)

type handlerTester struct {
	assertion func(w http.ResponseWriter, r *http.Request)
}

func (ht *handlerTester) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ht.assertion(w, r)
}

func Test_responseHeaderHandler(t *testing.T) {
	type args struct {
		key   string
		value string
	}
	tests := []struct {
		name string
		args args
		want http.Header
	}{
		{
			name: "empty key and value",
			args: args{},
			want: http.Header{},
		},
		{
			name: "non-empty key and empty value",
			args: args{key: "TestKey"},
			want: http.Header{},
		},
		{
			name: "empty key and non-empty value",
			args: args{value: "TestValue"},
			want: http.Header{},
		},
		{
			name: "non-empty key and non-empty value",
			args: args{key: "TestKey", value: "TestValue"},
			want: http.Header{"Testkey": []string{"TestValue"}}, // <- note: header name normalization
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := &handlerTester{
				assertion: func(w http.ResponseWriter, r *http.Request) {
					if !reflect.DeepEqual(w.Header(), tt.want) {
						t.Errorf("responseHeaderHandler() = %#v, want %#v", w.Header(), tt.want)
					}
				},
			}
			handler := responseHeaderHandler(tt.args.key, tt.args.value)

			// Creates a fake server that wraps tester with our handler
			ts := httptest.NewServer(handler(tester))
			defer ts.Close()

			// Do a dummy request
			resp, err := http.Get(ts.URL)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
		})
	}
}

func Test_traceIDHeaderHandler(t *testing.T) {
	type args struct {
		seekForHeader  string
		reqHeaderName  string
		reqHeaderValue string
	}
	tests := []struct {
		name string
		args args
		want string
	}{
		{
			name: "no header = no trace ID",
			args: args{},
			want: "",
		},
		{
			name: "no header value = no trace ID",
			args: args{
				seekForHeader: "Test",
				reqHeaderName: "Test",
			},
			want: "",
		},
		{
			name: "seeking header not matching trace header = no trace ID",
			args: args{
				seekForHeader:  "Ghost",
				reqHeaderName:  "Test",
				reqHeaderValue: "803da24f22b9d6dbbe94006bf31bfb20",
			},
			want: "",
		},
		{
			name: "seeking header matches",
			args: args{
				seekForHeader:  "Test",
				reqHeaderName:  "Test",
				reqHeaderValue: "803da24f22b9d6dbbe94006bf31bfb20",
			},
			want: "803da24f22b9d6dbbe94006bf31bfb20",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tester := &handlerTester{
				assertion: func(w http.ResponseWriter, r *http.Request) {
					// Get the trace ID from the HTTP request
					traceID := ctx.TraceID(r.Context())

					if traceID != tt.want {
						t.Errorf("traceIDHeaderHandler() = %#v, want %#v", traceID, tt.want)
					}
				},
			}

			handler := traceIDHeaderHandler(tt.args.seekForHeader)

			// Creates a fake server that wraps tester with our handler
			ts := httptest.NewServer(handler(tester))
			defer ts.Close()

			client := &http.Client{Timeout: 2 * time.Second}

			// Creates a new request
			req, err := http.NewRequest("GET", ts.URL, nil)
			if err != nil {
				t.Fatal(err)
			}

			// Adds the test header to the request
			if tt.args.reqHeaderName != "" {
				req.Header.Add(tt.args.reqHeaderName, tt.args.reqHeaderValue)
			}

			// Sends the request to the fake server
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
		})
	}
}
