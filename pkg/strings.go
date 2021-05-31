package pkg

import "strings"

// Chunk splits a given string in chunk of a given length with
// a given sep string.
// If the chunk length is < 1, the given string is returned untouched.
// If the chunk lenght is >= to the given string length, the given string
// is returned untouched.
// If the sep length is < 1, the given string is returned untouched.
// If the sep length is >= to the chunck length, the given string is
// returned untouched.
func Chunk(input string, length int, sep string) string {
	if len(input) <= length || len(sep) >= length {
		return input
	}

	c := []string{}

	for len(input) > 0 {
		c = append(c, input[:length])
		input = input[length:]

		if len(input) < length {
			length = len(input)
		}
	}

	return strings.Join(c, sep)
}
