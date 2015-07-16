package firetest

import (
	"encoding/json"
	"io"
	"io/ioutil"
	"net/http"
	"strings"
)

func sanitizePath(p string) string {
	s := strings.Trim(p, "/")
	return strings.TrimSuffix(s, ".json")
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
