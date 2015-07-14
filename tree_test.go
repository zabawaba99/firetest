package firetest

import (
	"strings"
	"testing"

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
		tree.add(test.path, test.node)

		rabbitHole := strings.Split(test.path, "/")
		previous := tree.rootNode
		for i := 0; i < len(rabbitHole); i++ {
			var ok bool
			previous, ok = previous.children[rabbitHole[i]]
			require.True(t, ok, test.path)
		}

		assert.NoError(t, equalNodes(test.node, previous), test.path)
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

	tree.del("root/only/one/child")
	assert.Nil(t, tree.get("root/only/one/child/here"))
	assert.Nil(t, tree.get("root/only/one/child"))
	assert.Nil(t, tree.get("root/only/one"))
	n := tree.get("root/only")
	require.NotNil(t, n)
	assert.Len(t, n.children, 2)
}
