package converter

import (
	"io"
	"reflect"
	"sync"
	"testing"
)

func TestNewProvider(t *testing.T) {
	baseProvider := provider{
		mux: sync.Mutex{},
	}

	tests := []struct {
		name       string
		converters []Converter
		want       Provider
	}{
		{
			name:       "no converter given",
			converters: []Converter{},
			want:       &baseProvider,
		},
		{
			name:       "1 converter given",
			converters: []Converter{&testConverter{}},
			want: func() Provider {
				p := baseProvider
				p.converters = map[ID]Converter{testConverterID: &testConverter{}}
				return &p
			}(),
		},
		{
			name:       "duplicate converters given",
			converters: []Converter{&testConverter{}, &testConverter{}},
			want: func() Provider {
				p := baseProvider
				p.converters = map[ID]Converter{testConverterID: &testConverter{}}
				return &p
			}(),
		},
		{
			name:       "2 different converters given",
			converters: []Converter{&testConverter{}, &rfc5322{}},
			want: func() Provider {
				p := baseProvider
				p.converters = map[ID]Converter{
					testConverterID: &testConverter{},
					RFC5322ID:       &rfc5322{},
				}
				return &p
			}(),
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
			converters: []Converter{&testConverter{}},
			want:       []ID{testConverterID},
		},
		{
			name:       "duplicate converter given",
			converters: []Converter{&testConverter{}, &testConverter{}},
			want:       []ID{testConverterID},
		},
		{
			name:       "several converters given",
			converters: []Converter{&testConverter{}, &rfc5322{}},
			want:       []ID{RFC5322ID, testConverterID},
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
			cid:        testConverterID,
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "requested converter exist",
			converters: []Converter{&testConverter{}},
			cid:        testConverterID,
			want:       &testConverter{},
			wantErr:    false,
		},
		{
			name:       "requested converter exist when there are several converters",
			converters: []Converter{&rfc5322{}, &testConverter{}},
			cid:        testConverterID,
			want:       &testConverter{},
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

const testConverterID ID = "testConverter"

type testConverter struct {
	id      string
	message *Message
	err     error
}

func (t *testConverter) ID() ID {
	return testConverterID
}

func (t *testConverter) Convert(data io.ReadSeeker) (*Message, error) {
	return t.message, t.err
}
