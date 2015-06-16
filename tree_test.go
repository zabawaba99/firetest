package firetest

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func newTestNode(v interface{}) *node {
	return &node{
		value:    v,
		children: map[string]*node{},
	}
}

func newTestNodeWithKids(children map[string]*node) *node {
	return &node{
		children: children,
	}
}

func equalNodes(expected, actual *node) error {
	if ec, ac := len(expected.children), len(actual.children); ec != ac {
		return fmt.Errorf("Children count is not the same\n\tExpected: %d\n\tActual: %d", ec, ac)
	}

	if len(expected.children) == 0 {
		if !assert.ObjectsAreEqualValues(expected.value, actual.value) {
			return fmt.Errorf("Node values not equal\n\tExpected: %T %v\n\tActual: %T %v", expected.value, expected.value, actual.value, actual.value)
		}
		return nil
	}

	for child, n := range expected.children {
		n2, ok := actual.children[child]
		if !ok {
			return fmt.Errorf("Expected node to have child: %s", child)
		}

		err := equalNodes(n, n2)
		if err != nil {
			return err
		}
	}
	return nil
}

func TestNewNode(t *testing.T) {

	for _, test := range []struct {
		name string
		node *node
	}{
		{
			name: "scalars/string",
			node: newTestNode("foo"),
		},
		{
			name: "scalars/number",
			node: newTestNode(2),
		},
		{
			name: "scalars/decimal",
			node: newTestNode(2.2),
		},
		{
			name: "scalars/boolean",
			node: newTestNode(false),
		},
		{
			name: "arrays/strings",
			node: newTestNodeWithKids(map[string]*node{
				"0": newTestNode("foo"),
				"1": newTestNode("bar"),
			}),
		},
		{
			name: "arrays/booleans",
			node: newTestNodeWithKids(map[string]*node{
				"0": newTestNode(true),
				"1": newTestNode(false),
			}),
		},
		{
			name: "arrays/numbers",
			node: newTestNodeWithKids(map[string]*node{
				"0": newTestNode(1),
				"1": newTestNode(2),
				"2": newTestNode(3),
			}),
		},
		{
			name: "arrays/decimals",
			node: newTestNodeWithKids(map[string]*node{
				"0": newTestNode(1.1),
				"1": newTestNode(2.2),
				"2": newTestNode(3.3),
			}),
		},
		{
			name: "objects/simple",
			node: newTestNodeWithKids(map[string]*node{
				"foo": newTestNode("bar"),
			}),
		},
		{
			name: "objects/complex",
			node: newTestNodeWithKids(map[string]*node{
				"foo":  newTestNode("bar"),
				"foo1": newTestNode(2),
				"foo2": newTestNode(true),
				"foo3": newTestNode(3.42),
			}),
		},
		{
			name: "objects/nested",
			node: newTestNodeWithKids(map[string]*node{
				"dinosaurs": newTestNodeWithKids(map[string]*node{
					"bruhathkayosaurus": newTestNodeWithKids(map[string]*node{
						"appeared": newTestNode(-70000000),
						"height":   newTestNode(25),
						"length":   newTestNode(44),
						"order":    newTestNode("saurischia"),
						"vanished": newTestNode(-70000000),
						"weight":   newTestNode(135000),
					}),
					"lambeosaurus": newTestNodeWithKids(map[string]*node{
						"appeared": newTestNode(-76000000),
						"height":   newTestNode(2.1),
						"length":   newTestNode(12.5),
						"order":    newTestNode("ornithischia"),
						"vanished": newTestNode(-75000000),
						"weight":   newTestNode(5000),
					}),
				}),
				"scores": newTestNodeWithKids(map[string]*node{
					"bruhathkayosaurus": newTestNode(55),
					"lambeosaurus":      newTestNode(21),
				}),
			}),
		},
		{
			name: "objects/with_arrays",
			node: newTestNodeWithKids(map[string]*node{
				"regular": newTestNode("item"),
				"booleans": newTestNodeWithKids(map[string]*node{
					"0": newTestNode(false),
					"1": newTestNode(true),
				}),
				"numbers": newTestNodeWithKids(map[string]*node{
					"0": newTestNode(1),
					"1": newTestNode(2),
				}),
				"decimals": newTestNodeWithKids(map[string]*node{
					"0": newTestNode(1.1),
					"1": newTestNode(2.2),
				}),
				"strings": newTestNodeWithKids(map[string]*node{
					"0": newTestNode("foo"),
					"1": newTestNode("bar"),
				}),
			}),
		},
	} {
		data, err := ioutil.ReadFile("fixtures/" + test.name + ".json")
		require.NoError(t, err, test.name)

		var v interface{}
		require.NoError(t, json.Unmarshal(data, &v), test.name)

		n := newNode(v)
		assert.NoError(t, equalNodes(test.node, n), test.name)
	}
}

func TestTreeAdd(t *testing.T) {
	for _, test := range []struct {
		path string
		node *node
	}{
		{
			path: "scalars/string",
			node: newTestNode("foo"),
		},
		{
			path: "s/c/a/l/a/r/s/s/t/r/i/n/g",
			node: newTestNodeWithKids(map[string]*node{
				"0": newTestNode("foo"),
				"1": newTestNode("bar"),
			}),
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
