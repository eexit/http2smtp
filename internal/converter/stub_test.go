package converter

import "testing"

func TestStub_ID(t *testing.T) {
	s := &Stub{}

	if id := s.ID(); id != StubConverterID {
		t.Errorf("ID() = %#v, want %#v", id, StubConverterID)
	}

	s = &Stub{StubID: ID("foo")}

	if id := s.ID(); id != ID("foo") {
		t.Errorf("ID() = %#v, want %#v", id, ID("foo"))
	}
}
