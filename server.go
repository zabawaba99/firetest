package firetest

import (
	"encoding/base32"
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
func (ft *Firetest) Start() error {
	l, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		if l, err = net.Listen("tcp6", "[::1]:0"); err != nil {
			return fmt.Errorf("failed to listen on a port: %v", err)
		}
	}
	ft.listener = l

	s := http.Server{Handler: ft}
	go func() {
		if err := s.Serve(l); err != nil {
			log.Printf("error serving: %s", err)
		}

		// close up shop when this exits
	}()
	ft.URL = "http://" + ft.listener.Addr().String()
	return nil
}

// Close closes the server
func (ft *Firetest) Close() error {
	if ft.listener != nil {
		return ft.listener.Close()
	}
	return nil
}

func (ft *Firetest) ServeHTTP(w http.ResponseWriter, req *http.Request) {
	log.Printf("Got request: %s %s\n", req.Method, req.URL.String())

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

func (ft *Firetest) set(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var n node
	if err := json.Unmarshal(body, &n); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}

	ft.db.add(sanitizePath(req.URL.Path), &n)
	w.Write(body)
}

func (ft *Firetest) update(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var n node
	if err := json.Unmarshal(body, &n); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}

	ft.db.update(sanitizePath(req.URL.Path), &n)
	w.Write(body)
}

func (ft *Firetest) create(w http.ResponseWriter, req *http.Request) {
	body, err := ioutil.ReadAll(req.Body)
	if err != nil || len(body) == 0 {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(missingBody))
		return
	}

	var n node
	if err := json.Unmarshal(body, &n); err != nil {
		w.WriteHeader(http.StatusBadRequest)
		w.Write([]byte(invalidJSON))
		return
	}

	src := []byte(fmt.Sprint(time.Now().UnixNano()))
	name := base32.StdEncoding.EncodeToString(src)
	path := fmt.Sprintf("%s/~%s", sanitizePath(req.URL.Path), name)
	ft.db.add(path, &n)

	rtn := map[string]string{"name": name}
	if err := json.NewEncoder(w).Encode(rtn); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ft *Firetest) del(w http.ResponseWriter, req *http.Request) {
	ft.db.del(sanitizePath(req.URL.Path))
}

func (ft *Firetest) get(w http.ResponseWriter, req *http.Request) {
	n := ft.db.get(sanitizePath(req.URL.Path))
	w.Header().Add("Content-Type", "application/json")
	if err := json.NewEncoder(w).Encode(n); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
