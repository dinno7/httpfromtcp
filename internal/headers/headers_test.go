package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaders(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFoo: Bar\r\n\r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	assert.Equal(t, "localhost:42069", headers.Get("Host"))
	assert.Equal(t, "", headers["NotExists"])
	assert.Equal(t, 35, n)
	assert.True(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.ErrorIs(t, err, ErrInvalidWhitespaceAfterHeaderKey)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Invalid character:
	headers = NewHeaders()
	data = []byte("H©st: localhost:42069\r\n\r\n")
	n, done, err = headers.Parse(data)
	require.ErrorIs(t, err, ErrInvalidCharacter)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	// Test: Multiple header value with single key

	headers = NewHeaders()
	data = []byte(
		"Set-Person: lane-loves-go\r\nSet-Person: prime-loves-zig\r\nSet-Person: tj-loves-ocaml\r\n\r\n",
	)
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.Equal(t, 86, n)
	assert.True(t, done)
	assert.Equal(t, headers.Get("Set-Person"), "lane-loves-go, prime-loves-zig, tj-loves-ocaml")
}
