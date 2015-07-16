package firetest

import (
	"bytes"
	"encoding/json"
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

func TestNew(t *testing.T) {
	ft := New()
	assert.NotNil(t, ft)
}

func TestURL(t *testing.T) {
	ft := New()
	ft.Start()
	assert.Regexp(t, regexp.MustCompile(`https?://127.0.0.1:\d+`), ft.URL)

	ft.Close()
}

func TestClose(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	// ACT
	ft.Close()

	// ASSERT
	_, err := http.Get(ft.URL)
	assert.Error(t, err)
	assert.IsType(t, (*url.Error)(nil), err)
}

func TestServeHTTP(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	// ACT
	req, err := http.NewRequest("GET", ft.URL+"/.json", nil)
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)
}

func TestServeHTTP_MissingJSON(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	// ACT
	req, err := http.NewRequest("GET", ft.URL, nil)
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusForbidden, resp.Code)
	assert.Equal(t, missingJSONExtension, resp.Body.Bytes())
}

func TestServerInvalidBody(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	for _, method := range []string{"PATCH", "POST", "PUT"} {
		// ACT
		req, err := http.NewRequest(method, ft.URL+"/.json", strings.NewReader("{asd}"))
		require.NoError(t, err)
		resp := httptest.NewRecorder()
		ft.serveHTTP(resp, req)

		// ASSERT
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, invalidJSON, resp.Body.Bytes())
	}
}

func TestServerMissingBody(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	for _, method := range []string{"PATCH", "POST", "PUT"} {
		// ACT
		req, err := http.NewRequest(method, ft.URL+"/.json", bytes.NewReader(nil))
		require.NoError(t, err)
		resp := httptest.NewRecorder()
		ft.serveHTTP(resp, req)

		// ASSERT
		assert.Equal(t, http.StatusBadRequest, resp.Code)
		assert.Equal(t, missingBody, resp.Body.Bytes())
	}
}

func TestServerCreate(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	// ACT
	body := `"bar"`
	req, err := http.NewRequest("POST", ft.URL+"/foo.json", strings.NewReader(body))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)

	var v map[string]string
	err = json.NewDecoder(resp.Body).Decode(&v)
	require.NoError(t, err)

	name, ok := v["name"]
	assert.True(t, ok)
	assert.NotEmpty(t, name)
}

func TestServerSet(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	// ACT
	body := `"bar"`
	req, err := http.NewRequest("PUT", ft.URL+"/foo.json", strings.NewReader(body))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, body, string(respBody))
}

func TestServerDel(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()
	path := "foo/bar"
	n := newNode(2)
	ft.db.add(path, n)

	// ACT
	req, err := http.NewRequest("DELETE", ft.URL+"/"+path+".json", nil)
	require.NoError(t, err)

	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)
	n = ft.db.get(path)
	assert.Nil(t, n)
}

func TestServerUpdate(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	path := "some/awesome/path"
	body := map[string]interface{}{
		"foo":  "bar",
		"fooy": true,
		"bar":  []interface{}{false, "lolz"},
	}
	ft.db.add(path, newNode(body))

	// ACT
	newVal := `"notbar"`
	req, err := http.NewRequest("PATCH", ft.URL+"/some/awesome/path/foo.json", strings.NewReader(newVal))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)
	respBody, err := ioutil.ReadAll(resp.Body)
	require.NoError(t, err)
	assert.Equal(t, newVal, string(respBody))
}

func TestServerGet(t *testing.T) {
	// ARRANGE
	ft := New()
	ft.Start()

	path := "some/awesome/path"
	body := map[string]interface{}{
		"foo":  "bar",
		"fooy": true,
		"bar":  []interface{}{false, "lolz"},
	}
	ft.db.add(path, newNode(body))

	b, err := json.Marshal(&body)
	require.NoError(t, err)

	// ACT
	req, err := http.NewRequest("GET", fmt.Sprintf("%s/%s.json", ft.URL, path), bytes.NewReader(b))
	require.NoError(t, err)
	resp := httptest.NewRecorder()
	ft.serveHTTP(resp, req)

	// ASSERT
	assert.Equal(t, http.StatusOK, resp.Code)
	var respBody map[string]interface{}
	require.NoError(t, json.NewDecoder(resp.Body).Decode(&respBody))

	assert.EqualValues(t, body, respBody)
}
