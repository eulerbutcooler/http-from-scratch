package headers

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHeaderParse(t *testing.T) {
	// Test: Valid single header
	headers := NewHeaders()
	data := []byte("Host: localhost:42069\r\nFoo:     barbar  \r\n")
	n, done, err := headers.Parse(data)
	require.NoError(t, err)
	require.NotNil(t, headers)
	host, hostOk := headers.Get("Host")
	assert.True(t, hostOk)
	assert.Equal(t, "localhost:42069", host)
	missing, missingOk := headers.Get("MissingKey")
	assert.False(t, missingOk)
	assert.Equal(t, "", missing)
	assert.Equal(t, 42, n)
	assert.False(t, done)

	// Test: Invalid spacing header
	headers = NewHeaders()
	data = []byte("       Host : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	//Test: Invalid token characters in field name
	headers = NewHeaders()
	data = []byte("H@st : localhost:42069       \r\n\r\n")
	n, done, err = headers.Parse(data)
	require.Error(t, err)
	assert.Equal(t, 0, n)
	assert.False(t, done)

	//Test: Multivalued headers
	headers = NewHeaders()
	data = []byte("Host: localhost:42069\r\nHost: localhost:42069\r\nHost: localhost:42068 \r\n")
	n, done, err = headers.Parse(data)
	require.NoError(t, err)
	assert.NotNil(t, headers)
	hostMulti, hostMultiOk := headers.Get("HOST")
	assert.True(t, hostMultiOk)
	assert.Equal(t, "localhost:42069,localhost:42069,localhost:42068", hostMulti)
	assert.False(t, done)
}
