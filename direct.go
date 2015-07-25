package firetest

import (
	"encoding/base64"
	"fmt"
	"sync/atomic"
	"time"
)

// RequireAuth determines whether or not a Firetest server
// will require that each request be authorized
func (ft *Firetest) RequireAuth(v bool) {
	var val int32
	if v {
		val = 1
	}
	atomic.StoreInt32(ft.requireAuth, val)
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

// Delete removes the data at the requested location.
// Any data at child locations will also be deleted.
//
// Reference https://www.firebase.com/docs/rest/api/#section-delete
func (ft *Firetest) Delete(path string) {
	ft.db.del(sanitizePath(path))
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

// Set writes data to at the given location.
// This will overwrite any data at this location and all child locations.
//
// Reference https://www.firebase.com/docs/rest/api/#section-put
func (ft *Firetest) Set(path string, v interface{}) {
	ft.db.add(sanitizePath(path), newNode(v))
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
