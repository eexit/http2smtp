package converter

import (
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sort"
	"sync"

	validator "github.com/go-playground/validator/v10"
)

var val = validator.New()

type (
	// ID is a converter ID type
	ID  string
	ids []ID
)

// Len implements type sorting
func (i ids) Len() int {
	return len(i)
}

// Swap implements type sorting
func (i ids) Swap(a, b int) {
	i[a], i[b] = i[b], i[a]
}

// Less implements type sorting
func (i ids) Less(a, b int) bool {
	return i[a] < i[b]
}

// Converter converts an input to a ConvertedMessage
type Converter interface {
	ID() ID
	Convert(r *http.Request) (*Message, error)
}

// Provider exposes the provider methods
type Provider interface {
	IDs() []ID
	Get(cid ID) (Converter, error)
}

type provider struct {
	mux        sync.Mutex
	converters map[ID]Converter
}

// NewProvider returns a new and initialized provider with given converters
func NewProvider(converters ...Converter) Provider {
	p := &provider{
		mux: sync.Mutex{},
	}

	if len(converters) > 0 {
		p.converters = make(map[ID]Converter, len(converters))
		for _, c := range converters {
			p.converters[c.ID()] = c
		}
	}
	return p
}

func (p *provider) IDs() []ID {
	p.mux.Lock()
	defer p.mux.Unlock()
	ids := ids{}

	for id := range p.converters {
		ids = append(ids, id)
	}

	sort.Sort(ids) // Sorts by name so the order is predictable

	return ids
}

func (p *provider) Get(cid ID) (Converter, error) {
	p.mux.Lock()
	defer p.mux.Unlock()

	for id, c := range p.converters {
		if id == cid {
			return c, nil
		}
	}
	return nil, fmt.Errorf("converter ID %v not found", cid)
}

// slurpBody idempotently copies a request's body
func slurpBody(r *http.Request) (io.ReadSeeker, error) {
	body, err := io.ReadAll(r.Body)
	if err != nil {
		return nil, err
	}
	defer r.Body.Close()

	r.Body = io.NopCloser(bytes.NewBuffer(body))
	return bytes.NewReader(body), nil
}
