package firetest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func sanitizePath(p string) string {
	// remove slashes from the front and back
	//	/foo/.json -> foo/.json
	s := strings.Trim(p, "/")

	// remove .json extension
	//	foo/.json -> foo/
	s = strings.TrimSuffix(s, ".json")

	// trim an potential trailing slashes
	//	foo/ -> foo
	return strings.TrimSuffix(s, "/")
}

func unmarshal(w http.ResponseWriter, r io.Reader) ([]byte, interface{}, bool) {
	body, err := ioutil.ReadAll(r)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(missingBody)
		return nil, nil, false
	}

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write(invalidJSON)
		return nil, nil, false
	}
	return body, v, true
}
