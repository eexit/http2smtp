package converter

import (
	"io"
)

// Converter converts an input to a ConvertedMessage
type Converter interface {
	Convert(data io.ReadSeeker) (*Message, error)
}
