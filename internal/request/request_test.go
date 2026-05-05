package request

import (
	"io"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type chunkReader struct {
	data            string
	numBytesPerRead int
	pos             int
}

func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := cr.pos + cr.numBytesPerRead
	if endIndex > len(cr.data) {
		endIndex = len(cr.data)
	}
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

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

func TestRequestLineChunkParse(t *testing.T) {
	reader := &chunkReader{
		data:            "GET / HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data:            "GET /coffee HTTP/1.1\r\nHost: localhost:42069\r\nUser-Agent: curl/7.81.0\r\nAccept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HttpVersion)
}
