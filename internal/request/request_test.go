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

// Read reads up to len(p) or numBytesPerRead bytes from the string per call
// its useful for simulating reading a variable number of bytes per chunk from a network connection
func (cr *chunkReader) Read(p []byte) (n int, err error) {
	if cr.pos >= len(cr.data) {
		return 0, io.EOF
	}
	endIndex := min(cr.pos+cr.numBytesPerRead, len(cr.data))
	n = copy(p, cr.data[cr.pos:endIndex])
	cr.pos += n

	return n, nil
}

func TestRequestLineParse(t *testing.T) {
	// Test: Good GET Request line
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n\r\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Good GET Request line with path
	reader = &chunkReader{
		data: "GET /coffee HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Bad GET Request line with incorrect version
	reader = &chunkReader{
		data: "GET /coffee HTTP/6.9\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n\r\n",
		numBytesPerRead: 1,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Good GET Request line
	r, err = RequestFromReader(strings.NewReader("GET / HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	require.Equal(t, "localhost:42069", r.Headers["host"])
	require.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Invalid HTTP Version numbering
	_, err = RequestFromReader(strings.NewReader("POST / HTTP/1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Invalid HTTP Version numbering
	_, err = RequestFromReader(strings.NewReader("POST / HTTP/2.0\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Good GET Request line with path
	r, err = RequestFromReader(strings.NewReader("GET /coffee HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "GET", r.RequestLine.Method)
	assert.Equal(t, "/coffee", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Invalid number of parts in request line
	_, err = RequestFromReader(strings.NewReader("/coffee HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Uncapitalized method
	_, err = RequestFromReader(strings.NewReader("get /coffee HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Unknown method
	_, err = RequestFromReader(strings.NewReader("GRAB /coffee HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)

	// Test: Empty request
	r, err = RequestFromReader(strings.NewReader("POST /coffee HTTP/1.1"))
	require.Error(t, err, r)

	// Test: Good Path with "*"
	r, err = RequestFromReader(strings.NewReader("POST * HTTP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "POST", r.RequestLine.Method)
	assert.Equal(t, "*", r.RequestLine.RequestTarget)
	assert.Equal(t, "1.1", r.RequestLine.HTTPVersion)

	// Test: Invalid version in Request line
	_, err = RequestFromReader(strings.NewReader("OPTIONS /forrest/gump TCP/1.1\r\n" +
		"Host: localhost:42069\r\n" +
		"User-Agent: curl/7.81.0\r\n" +
		"Accept: */*\r\n\r\n"))
	require.Error(t, err)
}

func TestHeadersParse(t *testing.T) {
	// Test: Standard Headers
	reader := &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"User-Agent: curl/7.81.0\r\n" +
			"Accept: */*\r\n\r\n",
		numBytesPerRead: 8,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "localhost:42069", r.Headers["host"])
	assert.Equal(t, "curl/7.81.0", r.Headers["user-agent"])
	assert.Equal(t, "*/*", r.Headers["accept"])

	// Test: Malformed Header
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Host localhost:42069\r\n\r\n",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: Empty Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"r\n",
		numBytesPerRead: 2,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)

	// Test: Duplicate Headers
	reader = &chunkReader{
		data: "GET / HTTP/1.1\r\n" +
			"Captain: I am the captain now\r\n" +
			"Captain: No I am the captain now\r\n" +
			"Tom: Hanks\r\n\r\n",
		numBytesPerRead: 8,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "I am the captain now, No I am the captain now", r.Headers["captain"])
	assert.Equal(t, "Hanks", r.Headers["tom"])
}

func TestBodyParse(t *testing.T) {
	// Test: Standard Body
	reader := &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 13\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err := RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "hello world!\n", string(r.Body))

	// Test: Body shorter than reported content length
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"Content-Length: 20\r\n" +
			"\r\n" +
			"partial content",
		numBytesPerRead: 3,
	}
	_, err = RequestFromReader(reader)
	require.Error(t, err)

	// Test: No Content-Length but Body Exists
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n" +
			"hello world!\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
	assert.Equal(t, "", string(r.Body))

	// Test: No body
	reader = &chunkReader{
		data: "POST /submit HTTP/1.1\r\n" +
			"Host: localhost:42069\r\n" +
			"\r\n",
		numBytesPerRead: 3,
	}
	r, err = RequestFromReader(reader)
	require.NoError(t, err)
	require.NotNil(t, r)
}
