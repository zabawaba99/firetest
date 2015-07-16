package firetest

import (
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestCreate(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = true
	)
	name := ft.Create(path, v)
	assert.True(t, strings.HasPrefix(name, "~"), "name is missing `~` prefix")

	n := ft.db.get(path + "/" + name)
	assert.Equal(t, v, n.value)
}

func TestDelete(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = true
	)

	// delete path directly
	ft.db.add(path, newNode(v))
	ft.Delete(path)
	assert.Nil(t, ft.db.get(path))

	// delete parent
	ft.db.add(path, newNode(v))
	ft.Delete("foo")
	assert.Nil(t, ft.db.get(path))
}

func TestUpdate(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = map[string]string{
			"1": "one",
			"2": "two",
			"3": "three",
		}
	)
	ft.db.add(path, newNode(v))

	ft.Update(path, map[string]string{
		"1": "three",
		"3": "one",
	})

	one := ft.db.get(path + "/1")
	three := ft.db.get(path + "/3")
	assert.Equal(t, "three", one.value)
	assert.Equal(t, "one", three.value)
}

func TestUpdateNil(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = map[string]string{
			"1": "one",
			"2": "two",
			"3": "three",
		}
	)
	ft.db.add(path, newNode(v))

	ft.Update(path, nil)
	assert.Nil(t, ft.db.get(path))
	assert.Nil(t, ft.db.get(path+"/1"))
	assert.Nil(t, ft.db.get(path+"/2"))
	assert.Nil(t, ft.db.get(path+"/3"))
}

func TestSet(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = true
	)
	ft.Set(path, v)

	n := ft.db.get(path)
	assert.Equal(t, v, n.value)
}

func TestGet(t *testing.T) {
	var (
		ft   = New()
		path = "foo/bar"
		v    = true
	)
	ft.db.add(path, newNode(v))

	val := ft.Get(path)
	assert.Equal(t, v, val)
}
