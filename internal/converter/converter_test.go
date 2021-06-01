package converter

import (
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"reflect"
	"strings"
	"sync"
	"testing"
)

// failingReader implements io.Reader and fails are reading
type failingReader struct{}

func (*failingReader) Read(p []byte) (n int, err error) {
	return 0, errors.New("read error")
}

func TestNewProvider(t *testing.T) {
	tests := []struct {
		name       string
		converters []Converter
		want       Provider
	}{
		{
			name:       "no converter given",
			converters: []Converter{},
			want:       &provider{mux: sync.Mutex{}},
		},
		{
			name:       "1 converter given",
			converters: []Converter{&Stub{}},
			want: &provider{
				mux:        sync.Mutex{},
				converters: map[ID]Converter{StubConverterID: &Stub{}},
			},
		},
		{
			name:       "duplicate converters given",
			converters: []Converter{&Stub{}, &Stub{}},
			want: &provider{
				mux:        sync.Mutex{},
				converters: map[ID]Converter{StubConverterID: &Stub{}},
			},
		},
		{
			name:       "2 different converters given",
			converters: []Converter{&Stub{}, &rfc5322{}},
			want: &provider{
				mux: sync.Mutex{},
				converters: map[ID]Converter{
					StubConverterID: &Stub{},
					RFC5322ID:       &rfc5322{},
				},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewProvider(tt.converters...); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewProvider() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_provider_IDs(t *testing.T) {
	tests := []struct {
		name       string
		converters []Converter
		want       []ID
	}{
		{
			name:       "no converter given",
			converters: []Converter{},
			want:       []ID{},
		},
		{
			name:       "1 converter given",
			converters: []Converter{&Stub{}},
			want:       []ID{StubConverterID},
		},
		{
			name:       "duplicate converter given",
			converters: []Converter{&Stub{}, &Stub{}},
			want:       []ID{StubConverterID},
		},
		{
			name:       "several converters given",
			converters: []Converter{&Stub{}, &rfc5322{}},
			want:       []ID{RFC5322ID, StubConverterID},
		},
		{
			name:       "IDs are returned in order",
			converters: []Converter{&rfc5322{}, &Stub{}},
			want:       []ID{RFC5322ID, StubConverterID},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.converters...)
			if got := p.IDs(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("provider.IDs() = %#v, want %#v", got, tt.want)
			}
		})
	}
}

func Test_provider_Get(t *testing.T) {
	tests := []struct {
		name       string
		converters []Converter
		cid        ID
		want       Converter
		wantErr    bool
	}{
		{
			name:       "provider has no converter and requested ID is empty",
			converters: []Converter{},
			cid:        ID(""),
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "requested converter does not exist",
			converters: []Converter{},
			cid:        StubConverterID,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "requested converter exist",
			converters: []Converter{&Stub{}},
			cid:        StubConverterID,
			want:       &Stub{},
			wantErr:    false,
		},
		{
			name:       "requested converter exist when there are several converters",
			converters: []Converter{&Stub{}, &rfc5322{}},
			cid:        StubConverterID,
			want:       &Stub{},
			wantErr:    false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			p := NewProvider(tt.converters...)
			got, err := p.Get(tt.cid)
			if (err != nil) != tt.wantErr {
				t.Errorf("provider.Get() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("provider.Get() = %v, want %v", got, tt.want)
			}
		})
	}
}

func Test_slurpBody(t *testing.T) {
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
			req:     httptest.NewRequest("", "/", &failingReader{}),
			want:    "",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := slurpBody(tt.req)
			if (err != nil) != tt.wantErr {
				t.Errorf("slurpBody() error = %v, wantErr %v", err, tt.wantErr)
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
				t.Errorf("slurpBody() = %#v, want %#v", gotAsStr, tt.want)
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
