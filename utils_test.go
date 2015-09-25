package firetest

import (
	"bytes"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSanitizePath(t *testing.T) {
	for i, test := range []struct {
		path     string
		expected string
	}{
		{"/", ""},
		{"foo", "foo"},
		{"foo/", "foo"},
		{"/foo", "foo"},
		{"/foo/", "foo"},
		{"/foo/.json", "foo"}, // issue #6
	} {
		assert.Equal(t, test.expected, sanitizePath(test.path), "%d", i)
	}
}

func TestUnmarshal(t *testing.T) {
	v := "foo"
	jsonV := `"foo"`
	w := httptest.NewRecorder()
	r := strings.NewReader(jsonV)
	b, val, ok := unmarshal(w, r)
	assert.Equal(t, []byte(jsonV), b)
	assert.Equal(t, v, val)
	assert.True(t, ok)
}

func TestUnmarshal_MissingBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := bytes.NewReader(nil)
	b, val, ok := unmarshal(w, r)
	assert.Nil(t, b)
	assert.Nil(t, val)
	assert.False(t, ok)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, []byte(missingBody), w.Body.Bytes())
}

func TestUnmarshal_InvalidBody(t *testing.T) {
	w := httptest.NewRecorder()
	r := strings.NewReader("{asda}")
	b, val, ok := unmarshal(w, r)
	assert.Nil(t, b)
	assert.Nil(t, val)
	assert.False(t, ok)

	assert.Equal(t, http.StatusBadRequest, w.Code)
	assert.Equal(t, []byte(invalidJSON), w.Body.Bytes())
}
