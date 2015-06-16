package firetest

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net"
	"net/http"
	"strings"
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

// NewFiretest creates a new Firetest server
func NewFiretest() *Firetest {
	return &Firetest{
		db: newTree(),
	}
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
	default:
		log.Println("not implemented yet")
	}
}

func (ft *Firetest) set(w http.ResponseWriter, req *http.Request) {
	var n node

	if err := json.NewDecoder(req.Body).Decode(&n); err != nil {
		msg := []byte(invalidJSON)
		if err == io.EOF {
			msg = []byte(missingBody)
		}
		w.WriteHeader(http.StatusBadRequest)
		w.Write(msg)
		return
	}

	// switch
}
