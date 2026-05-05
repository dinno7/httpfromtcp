package request

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRequestLineParse(t *testing.T) {
	r, err := RequestFromReader(
		strings.NewReader(
			"GET / HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "/", r.RequestTarget)
	assert.Equal(t, "1.1", r.HttpVersion)

	r, err = RequestFromReader(
		strings.NewReader(
			"GET /coffee HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.Method)
	assert.Equal(t, "/coffee", r.RequestTarget)
	assert.Equal(t, "1.1", r.HttpVersion)

	_, err = RequestFromReader(
		strings.NewReader(
			"/coffee HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.ErrorIs(t, err, ErrInvalidRequestHeader)

	_, err = RequestFromReader(
		strings.NewReader(
			"POST coffee HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.ErrorIs(t, err, ErrInvalidTargetPath)

	_, err = RequestFromReader(
		strings.NewReader(
			"WRONG /coffee HTTP/1.1\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.ErrorIs(t, err, ErrInvalidHttpMethod)

	_, err = RequestFromReader(
		strings.NewReader(
			"POST /coffee HTTP/2\r\nHost: localhost:3000\r\nUser-Agent: curl/8.19.0\r\nAccept: */*\r\n\r\n",
		),
	)
	require.ErrorIs(t, err, ErrUnsupportedHttpVersion)
}
