package main

import (
	"bytes"
	"io"
	"net/http"
	"testing"

	"github.com/2bitburrito/http-implementation/internal/server"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestMain(t *testing.T) {
	server, err := server.Serve(45288, Handler)
	require.NoError(t, err, "error starting server")
	require.NotNil(t, server)

	defer server.Close()

	resp, err := http.Get("http://localhost:45288" + "/ping")
	require.NoError(t, err, "couldn't request server")
	require.NotNil(t, resp)
	assert.Equal(t, resp.StatusCode, http.StatusOK)

	defer resp.Body.Close()

	data, err := io.ReadAll(resp.Body)
	require.NoError(t, err, "error reading body")
	require.NotNil(t, data)
	success := bytes.Contains(data, []byte("Your request was an absolute banger."))
	assert.True(t, success)
}
