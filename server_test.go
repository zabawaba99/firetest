package firetest

import (
	"bytes"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/url"
	"regexp"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestNewFiretest(t *testing.T) {
	ft := NewFiretest()
	assert.NotNil(t, ft)
}

func TestURL(t *testing.T) {
	ft := NewFiretest()
	assert.NoError(t, ft.Start())
	assert.Regexp(t, regexp.MustCompile(`https?://127.0.0.1:\d+`), ft.URL)

	ft.Close()
}

func TestClose(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	assert.NoError(t, ft.Close())

	_, err := http.Get(ft.URL)
	assert.Error(t, err)
	assert.IsType(t, (*url.Error)(nil), err)
}

func TestCloseFailure(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	assert.NoError(t, ft.Close())
	assert.Error(t, ft.Close())
}

func TestServeHTTP(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	req, err := http.NewRequest("GET", ft.URL+"/.json", nil)
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestServeHTTP_MissingJSON(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	req, err := http.NewRequest("GET", ft.URL, nil)
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.ServeHTTP(resp, req)

	assert.Equal(t, http.StatusForbidden, resp.Code)
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, missingJSONExtension, string(b))
}

func TestSet(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	path := "foo"
	body := `"bar"`
	req, err := http.NewRequest("PUT", fmt.Sprintf("%s/%s.json", ft.URL, path), strings.NewReader(body))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.set(resp, req)

	// TODO
	assert.Equal(t, http.StatusOK, resp.Code)
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(respBody))

}

func TestSet_NoBody(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	req, err := http.NewRequest("PUT", ft.URL+"/.json", bytes.NewReader(nil))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.set(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, missingBody, string(b))
}

func TestSet_InvalidBody(t *testing.T) {
	ft := NewFiretest()
	require.NoError(t, ft.Start())

	req, err := http.NewRequest("PUT", ft.URL+"/.json", strings.NewReader("{asd}"))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.set(resp, req)

	assert.Equal(t, http.StatusBadRequest, resp.Code)
	b, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, invalidJSON, string(b))
}
