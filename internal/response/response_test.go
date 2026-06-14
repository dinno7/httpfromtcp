package response

import (
	"bytes"
	"strings"
	"testing"

	"github.com/dinno7/httpfromtcp/internal/headers"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestInitResponse(t *testing.T) {
	writer := new(bytes.Buffer)
	response := NewResponse(writer)
	require.NotNil(t, response)
	assert.Empty(t, response.headers)
	assert.Equal(t, responseStateStatusLine, response.state)
	assert.Equal(t, response.statusCode, StatusCodeOk)

	// Test: Setting status code
	response.SetStatusCode(StatusCodeInternalServerError)
	assert.Equal(t, response.statusCode, StatusCodeInternalServerError)
}

func TestResponseWrite(t *testing.T) {
	t.Run("writing status line", func(t *testing.T) {
		writer := new(bytes.Buffer)
		response := NewResponse(writer)
		require.NotNil(t, response)

		n, err := response.writeStatusLine()
		require.NoError(t, err)
		assert.Equal(t, n, 17)
		assert.Equal(t, "HTTP/1.1 200 Ok\r\n", writer.String())
		assert.Equal(t, responseStateHeaders, response.state)
	})

	t.Run("writing header", func(t *testing.T) {
		writer := new(bytes.Buffer)
		response := NewResponse(writer)
		require.NotNil(t, response)

		n, err := response.writeHeaders()
		require.NoError(t, err)
		assert.Equal(t, n, 83)
		assert.Equal(t, "close", response.Headers().Get("connection"))
		assert.Equal(t, "no-cache", response.Headers().Get("cache-control"))
		assert.Equal(t, responseStateBody, response.state)
		assert.NotEmpty(t, response.Headers().Get("date"))

		response.Headers().Set("connection", "keep-alive")
		assert.Equal(t, "keep-alive", response.Headers().Get("connection"))
	})

	t.Run("writing body", func(t *testing.T) {
		writer := new(bytes.Buffer)
		response := NewResponse(writer)
		require.NotNil(t, response)

		n, err := response.Write([]byte("Hellow"))
		require.NoError(t, err)
		assert.Equal(t, responseStateDone, response.state)
		assert.Equal(t, response.Headers().Get("content-length"), "6")
		assert.Equal(t, n, 6)
		assert.Equal(t, writer.Len(), 125) // Status line + headers(default + content-length) + body
	})
}

func TestWrite_EmptyBody(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	n, err := resp.Write([]byte{})
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, responseStateDone, resp.state)
	assert.Equal(t, "0", resp.Headers().Get("content-length"))
	assert.True(t, strings.HasPrefix(writer.String(), "HTTP/1.1 200 Ok\r\n"))
	assert.Contains(t, writer.String(), "content-length: 0")
}

func TestWrite_NilBody(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	n, err := resp.Write(nil)
	require.NoError(t, err)
	assert.Equal(t, 0, n)
	assert.Equal(t, "0", resp.Headers().Get("content-length"))
}

func TestWrite_WithCustomHeaders(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	resp.Headers().Set("Content-Type", "text/plain")
	resp.Headers().Set("X-Custom", "custom-value")

	n, err := resp.Write([]byte("Hello"))
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, responseStateDone, resp.state)

	output := writer.String()
	assert.True(t, strings.HasPrefix(output, "HTTP/1.1 200 Ok\r\n"))
	assert.Contains(t, output, "content-type: text/plain")
	assert.Contains(t, output, "x-custom: custom-value")
	assert.Contains(t, output, "Hello")
}

func TestWrite_WithCustomStatusCode(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	resp.SetStatusCode(StatusCodeNotFound)

	n, err := resp.Write([]byte("Not Found"))
	require.NoError(t, err)
	assert.Equal(t, 9, n)
	assert.True(t, strings.HasPrefix(writer.String(), "HTTP/1.1 404 Not Found\r\n"))
}

func TestWrite_DefaultHeadersNotOverrideCustom(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	resp.Headers().Set("Connection", "keep-alive")
	resp.Headers().Set("Cache-Control", "public, max-age=3600")

	_, err := resp.Write([]byte("data"))
	require.NoError(t, err)
	assert.Equal(t, "keep-alive", resp.Headers().Get("connection"))
	assert.Equal(t, "public, max-age=3600", resp.Headers().Get("cache-control"))
}

func TestWriteChunkedBody_SingleChunk(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	n, err := resp.WriteChunkedBody([]byte("Hello"))
	require.NoError(t, err)
	assert.Equal(t, 10, n)
	assert.Equal(t, responseStateBody, resp.state)

	output := writer.String()
	assert.True(t, strings.HasPrefix(output, "HTTP/1.1 200 Ok\r\n"))
	assert.Contains(t, output, "transfer-encoding: chunked")
	assert.NotContains(t, output, "content-length:")
	assert.Contains(t, output, "5\r\nHello\r\n")
}

func TestWriteChunkedBody_MultipleChunks(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	n1, err := resp.WriteChunkedBody([]byte("Hello"))
	require.NoError(t, err)
	assert.Equal(t, 10, n1)

	n2, err := resp.WriteChunkedBody([]byte("World"))
	require.NoError(t, err)
	assert.Equal(t, 10, n2)
	assert.Equal(t, responseStateBody, resp.state)

	output := writer.String()
	assert.Contains(t, output, "5\r\nHello\r\n5\r\nWorld\r\n")
}

func TestWriteChunkedBody_FullFlow(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	_, err := resp.WriteChunkedBody([]byte("Hello"))
	require.NoError(t, err)

	_, err = resp.WriteChunkedBody([]byte("World"))
	require.NoError(t, err)

	_, err = resp.WriteChunkedBodyDone()
	require.NoError(t, err)
	assert.Equal(t, responseStateDone, resp.state)

	h := headers.NewHeaders()
	h.Set("X-Trailer", "trailer-value")
	_, err = resp.WriteTrailers(h)
	require.NoError(t, err)

	output := writer.String()
	assert.Contains(t, output, "transfer-encoding: chunked")
	assert.Contains(t, output, "5\r\nHello\r\n")
	assert.Contains(t, output, "5\r\nWorld\r\n")
	assert.Contains(t, output, "0\r\n\r\n")
	assert.Contains(t, output, "x-trailer: trailer-value")
}

func TestWriteChunkedBody_InvalidState(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	_, err := resp.Write([]byte("done"))
	require.NoError(t, err)
	assert.Equal(t, responseStateDone, resp.state)

	n, err := resp.WriteChunkedBody([]byte("should fail"))
	assert.ErrorIs(t, err, ErrInvalidState)
	assert.Equal(t, 0, n)
}

func TestWriteChunkedBodyDone_NoChunks(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	n, err := resp.WriteChunkedBodyDone()
	require.NoError(t, err)
	assert.Equal(t, 5, n)
	assert.Equal(t, responseStateDone, resp.state)
	assert.Equal(t, "0\r\n\r\n", writer.String())
}

func TestWriteTrailers(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	h := headers.NewHeaders()
	h.Set("X-Trailer-1", "value1")
	h.Set("X-Trailer-2", "value2")

	n, err := resp.WriteTrailers(h)
	require.NoError(t, err)
	assert.Greater(t, n, 0)

	output := writer.String()
	assert.Contains(t, output, "x-trailer-1: value1")
	assert.Contains(t, output, "x-trailer-2: value2")
}

func TestWriteTrailers_Empty(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)

	h := headers.NewHeaders()
	n, err := resp.WriteTrailers(h)
	require.NoError(t, err)
	assert.Equal(t, 2, n)
	assert.Equal(t, "\r\n", writer.String())
}

func TestSetStatusCode_ZeroDefaultsTo200(t *testing.T) {
	writer := new(bytes.Buffer)
	resp := NewResponse(writer)
	resp.SetStatusCode(0)

	_, err := resp.Write([]byte("test"))
	require.NoError(t, err)

	assert.True(t, strings.HasPrefix(writer.String(), "HTTP/1.1 200 Ok\r\n"))
}
