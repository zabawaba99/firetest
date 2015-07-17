/*
Package firetest provides utilities for Firebase testing

*/
package firetest

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"log"
	"net"
	"net/http"
	"strings"
	"sync"
	"time"
)

var (
	missingJSONExtension = []byte("append .json to your request URI to use the REST API")
	missingBody          = []byte(`{"error":"Error: No data supplied."}`)
	invalidJSON          = []byte(`{"error":"Invalid data; couldn't parse JSON object, array, or value. Perhaps you're using invalid characters in your key names."}`)
	invalidAuth          = []byte(`{"error" : "Could not parse auth token."}`)
)

// Firetest is a Firebase server implementation
type Firetest struct {
	// URL of form http://ipaddr:port with no trailing slash
	URL string
	// Secret used to authenticate with server
	Secret string

	listener net.Listener
	db       *treeDB

	authMtx     sync.RWMutex
	requireAuth bool
}

// New creates a new Firetest server
func New() *Firetest {
	return &Firetest{
		db:     newTree(),
		Secret: base64.URLEncoding.EncodeToString([]byte(fmt.Sprint(time.Now().UnixNano()))),
	}
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
	if !strings.HasSuffix(req.URL.Path, ".json") {
		w.WriteHeader(http.StatusForbidden)
		w.Write([]byte(missingJSONExtension))
		return
	}

	ft.authMtx.RLock()
	authenticate := ft.requireAuth
	ft.authMtx.RUnlock()
	if authenticate {
		var authenticated bool
		authHeader := req.URL.Query().Get("auth")
		switch {
		case strings.Contains(authHeader, "."):
			authenticated = ft.validJWT(authHeader)
		default:
			authenticated = authHeader == ft.Secret
		}

		if !authenticated {
			w.WriteHeader(http.StatusUnauthorized)
			w.Write(invalidAuth)
			return
		}
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

func decodeSegment(seg string) ([]byte, error) {
	if l := len(seg) % 4; l > 0 {
		seg += strings.Repeat("=", 4-l)
	}

	return base64.URLEncoding.DecodeString(seg)
}

func (ft *Firetest) validJWT(val string) bool {
	parts := strings.Split(val, ".")
	if len(parts) != 3 {
		return false
	}

	// validate header
	hb, err := decodeSegment(parts[0])
	if err != nil {
		log.Println("error decoding header", err)
		return false
	}
	var header map[string]string
	if err := json.Unmarshal(hb, &header); err != nil {
		log.Println("error unmarshaling header", err)
		return false
	}
	if header["alg"] != "HS256" || header["typ"] != "JWT" {
		return false
	}

	// validate claim
	cb, err := decodeSegment(parts[1])
	if err != nil {
		log.Println("error decoding claim", err)
		return false
	}
	var claim map[string]interface{}
	if err := json.Unmarshal(cb, &claim); err != nil {
		log.Println("error unmarshaling claim", err)
		return false
	}
	if e, ok := claim["exp"]; ok {
		// make sure not expired
		exp, ok := e.(float64)
		if !ok {
			log.Println("expiration not a number")
			return false
		}
		if int64(exp) < time.Now().Unix() {
			log.Println("token expired")
			return false
		}
	}
	// ensure uid present
	data, ok := claim["d"]
	if !ok {
		log.Println("missing data in claim")
		return false
	}

	d, ok := data.(map[string]interface{})
	if !ok {
		log.Println("claim['data'] is not map")
		return false
	}

	if _, ok := d["uid"]; !ok {
		log.Println("claim['data'] missing uid")
		return false
	}

	if sig, err := decodeSegment(parts[2]); err == nil {
		hasher := hmac.New(sha256.New, []byte(ft.Secret))
		signedString := strings.Join(parts[:2], ".")
		hasher.Write([]byte(signedString))

		if !hmac.Equal(sig, hasher.Sum(nil)) {
			log.Println("invalid jwt signature")
			return false
		}
	}

	return true
}

func (ft *Firetest) set(w http.ResponseWriter, req *http.Request) {
	body, v, ok := unmarshal(w, req.Body)
	if !ok {
		return
	}

	ft.Set(req.URL.Path, v)
	w.Write(body)
}

func (ft *Firetest) update(w http.ResponseWriter, req *http.Request) {
	body, v, ok := unmarshal(w, req.Body)
	if !ok {
		return
	}
	ft.Update(req.URL.Path, v)
	w.Write(body)
}

func (ft *Firetest) create(w http.ResponseWriter, req *http.Request) {
	_, v, ok := unmarshal(w, req.Body)
	if !ok {
		return
	}

	name := ft.Create(req.URL.Path, v)
	rtn := map[string]string{"name": name}
	if err := json.NewEncoder(w).Encode(rtn); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}

func (ft *Firetest) del(w http.ResponseWriter, req *http.Request) {
	ft.Delete(req.URL.Path)
}

func (ft *Firetest) get(w http.ResponseWriter, req *http.Request) {
	w.Header().Add("Content-Type", "application/json")

	v := ft.Get(req.URL.Path)
	if err := json.NewEncoder(w).Encode(v); err != nil {
		log.Printf("Error encoding json: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
	}
}
