package firetest

import (
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestTreeAdd(t *testing.T) {
	for _, test := range []struct {
		path string
		node *node
	}{
		{
			path: "scalars/string",
			node: newNode("foo"),
		},
		{
			path: "s/c/a/l/a/r/s/s/t/r/i/n/g",
			node: newNode([]interface{}{"foo", "bar"}),
		},
	} {
		tree := newTree()

		// listen for notifications
		notifications := tree.watch(test.path)
		exited := make(chan struct{})
		go func() {
			n, ok := <-notifications
			assert.True(t, ok)
			assert.Equal(t, "put", n.Name)
			assert.Equal(t, test.path, n.Data.Path)
			assert.Equal(t, test.node, n.Data.Data)
			close(exited)
		}()

		tree.add(test.path, test.node)

		rabbitHole := strings.Split(test.path, "/")
		previous := tree.rootNode
		for i := 0; i < len(rabbitHole); i++ {
			var ok bool
			previous, ok = previous.children[rabbitHole[i]]
			assert.True(t, ok, test.path)
		}

		assert.NoError(t, equalNodes(test.node, previous), test.path)
		select {
		case <-exited:
		case <-time.After(250 * time.Millisecond):
		}
		tree.stopWatching(test.path, notifications)
	}
}

func TestTreeGet(t *testing.T) {
	for _, test := range []struct {
		path string
		node *node
	}{
		{
			path: "scalars/string",
			node: newNode("foo"),
		},
		{
			path: "s/c/a/l/a/r/s/s/t/r/i/n/g",
			node: newNode([]interface{}{"foo", "bar"}),
		},
	} {
		tree := newTree()
		tree.add(test.path, test.node)

		assert.NoError(t, equalNodes(test.node, tree.get(test.path)), test.path)
	}
}

func TestTreeDel(t *testing.T) {
	existingNodes := []string{
		"root/only/two",
		"root/only/three",
		"root/only/one/child/here",
	}
	tree := newTree()
	for _, p := range existingNodes {
		tree.add(p, newNode(1))
	}

	// listen for notifications
	notifications := tree.watch("")
	exited := make(chan struct{})
	go func() {
		regex := regexp.MustCompile("(root/only/one/child|root)")
		n, ok := <-notifications
		assert.True(t, ok)
		assert.Equal(t, "put", n.Name)
		assert.Regexp(t, regex, n.Data.Path)

		n, ok = <-notifications
		assert.True(t, ok)
		assert.Equal(t, "put", n.Name)
		assert.Regexp(t, regex, n.Data.Path)
		close(exited)
	}()

	tree.del("root/only/one/child")
	assert.Nil(t, tree.get("root/only/one/child/here"))
	assert.Nil(t, tree.get("root/only/one/child"))
	assert.Nil(t, tree.get("root/only/one"))
	n := tree.get("root/only")
	require.NotNil(t, n)
	assert.Len(t, n.children, 2)
	_, exists := n.children["one"]
	assert.False(t, exists)

	tree.del("root")
	n = tree.get("")
	require.NotNil(t, n)
	assert.Len(t, n.children, 0)

	select {
	case <-exited:
	case <-time.After(250 * time.Millisecond):
	}
	tree.stopWatching("", notifications)
}
