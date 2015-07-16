/*
Package firetest provides utilities for Firebase testing

*/
package firetest

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"log"
	"net"
	"net/http"
	"strings"
	"time"
)

const (
	missingJSONExtension = "append .json to your request URI to use the REST API"
	missingBody          = `{"error":"Error: No data supplied."}`
	invalidJSON          = `{"error":"Invalid data; couldn't parse JSON object, array, or value. Perhaps you're using invalid characters in your key names."}`
)

// Firetest is a Firebase server implementation
type Firetest struct {
	// URL of form http://ipaddr:port with no trailing slash
	URL string

	listener net.Listener
	db       *treeDB
}

// New creates a new Firetest server
func New() *Firetest {
	return &Firetest{
		db: newTree(),
	}
}

func sanitizePath(p string) string {
	s := strings.Trim(p, "/")
	return strings.TrimSuffix(s, ".json")
}

// Start starts the server
func (ft *Firetest) Start() {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			panic(fmt.Errorf("failed to listen on a port: %v", err))
		}
	}
	ft.listener = l

	s := http.Server{Handler: http.HandlerFunc(func(w http.ResponseWriter, req *http.Request) {
		ft.serveHTTP(w, req)
	})}
	go func() {
		if err := s.Serve(l); err != nil {
			log.Printf("error serving: %s", err)
		}

		ft.Close()
	}()
	ft.URL = "http://" + ft.listener.Addr().String()
}

// Close closes the server
func (ft *Firetest) Close() {
	if ft.listener != nil {
		ft.listener.Close()
	}
}

func (ft *Firetest) serveHTTP(w http.ResponseWriter, req *http.Request) {
	if !strings.HasSuffix(req.URL.String(), ".json") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(missingJSONExtension))
		return
	}

	switch req.Method {
	case "PUT":
		ft.set(w, req)
	case "PATCH":
		ft.update(w, req)
	case "POST":
		ft.create(w, req)
	case "GET":
		ft.get(w, req)
	case "DELETE":
		ft.del(w, req)
	default:
		w.WriteHeader(http.StatusMethodNotAllowed)
		log.Println("not implemented yet")
	}
}

// Set writes data to at the given location.
// This will overwrite any data at this location and all child locations.
//
// Reference https://www.firebase.com/docs/rest/api/#section-put
func (ft *Firetest) Set(path string, v interface{}) {
	ft.db.add(sanitizePath(path), newNode(v))
}

func (ft *Firetest) set(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}

	ft.Set(req.URL.Path, v)
	w.Write(body)
}

// Update writes the enumerated children to this the given location.
// This will overwrite only children enumerated in the "value" parameter
// and will leave others untouched. Note that the update function is equivalent
// to calling Set() on the named children; it does not recursively update children
// if they are objects. Passing null as a value for a child is equivalent to
// calling remove() on that child.
//
// Reference https://www.firebase.com/docs/rest/api/#section-patch
func (ft *Firetest) Update(path string, v interface{}) {
	path = sanitizePath(path)
	if v == nil {
		ft.db.del(path)
	} else {
		ft.db.update(path, newNode(v))
	}
}

func (ft *Firetest) update(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}
	ft.Update(req.URL.Path, v)
	w.Write(body)
}

// Create generates a new child under the given location
// using a unique name and returns the name
//
// Reference https://www.firebase.com/docs/rest/api/#section-post
func (ft *Firetest) Create(path string, v interface{}) string {
	src := []byte(fmt.Sprint(time.Now().UnixNano()))
	name := "~" + base64.StdEncoding.EncodeToString(src)
	path = fmt.Sprintf("%s/%s", sanitizePath(path), name)

	ft.db.add(path, newNode(v))
	return name
}

func (ft *Firetest) create(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var v interface{}
	if err := json.Unmarshal(body, &v); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}

	name := ft.Create(req.URL.Path, v)
	rtn := map[string]string{"name": name}
	if err := json.NewEncoder(w).Encode(rtn); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

// Delete removes the data at the requested location.
// Any data at child locations will also be deleted.
//
// Reference https://www.firebase.com/docs/rest/api/#section-delete
func (ft *Firetest) Delete(path string) {
	ft.db.del(sanitizePath(path))
}

func (ft *Firetest) del(w http.ResponseWriter, req *http.Request) {
	ft.Delete(req.URL.Path)
}

// Get retrieves the data and all its children at the
// requested location
//
// Reference https://www.firebase.com/docs/rest/api/#section-get
func (ft *Firetest) Get(path string) (v interface{}) {
	n := ft.db.get(sanitizePath(path))
	if n != nil {
		v = n.objectify()
	}
	return v
}

func (ft *Firetest) get(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	v := ft.Get(req.URL.Path)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
