package firetest

import (
	"encoding/json"
	"fmt"
	"log"
	"strconv"
	"strings"
)

type node struct {
	value     interface{}
	children  map[string]*node
	sliceKids bool
}

func newNode(data interface{}) *node {
	n := &node{children: map[string]*node{}}

	switch data := data.(type) {
	case map[string]interface{}:
		for k, v := range data {
			child := newNode(v)
			n.children[k] = child
		}
	case []interface{}:
		n.sliceKids = true
		for i, v := range data {
			child := newNode(v)
			n.children[fmt.Sprint(i)] = child
		}
	case string, int, int8, int16, int32, int64, float32, float64, bool:
		n.value = data
	default:
		log.Printf("node - %T %v\n", data, data)
	}

	return n
}

func (n *node) UnmarshalJSON(b []byte) error {
	var v interface{}
	if err := json.Unmarshal(b, &v); err != nil {
		return err
	}

	return nil
}

func (n *node) MarshalJSON() ([]byte, error) {
	return json.Marshal(n.objectify())
}

func (n *node) objectify() interface{} {
	if n.value != nil {
		return n.value
	}

	if n.sliceKids {
		obj := make([]interface{}, len(n.children))
		for k, v := range n.children {
			index, err := strconv.Atoi(k)
			if err != nil {
				continue
			}
			obj[index] = v.objectify()
		}
		return obj
	}

	obj := map[string]interface{}{}
	for k, v := range n.children {
		obj[k] = v.objectify()
	}
	return obj
}

type treeDB struct {
	rootNode *node
}

func newTree() *treeDB {
	return &treeDB{
		rootNode: &node{
			children: map[string]*node{},
		},
	}
}

func (tree *treeDB) add(path string, n *node) {
	if path == "" {
		tree.rootNode = n
		return
	}

	rabbitHole := strings.Split(path, "/")
	previous := tree.rootNode
	var current *node
	for i := 0; i < len(rabbitHole)-1; i++ {
		step := rabbitHole[i]
		var ok bool
		current, ok = previous.children[step]
		if !ok {
			current = &node{children: map[string]*node{}}
			previous.children[step] = current
		}

		previous, current = current, nil
	}

	lastPath := rabbitHole[len(rabbitHole)-1]
	previous.children[lastPath] = n
}

func (tree *treeDB) get(path string) (current *node) {
	current = tree.rootNode

	rabbitHole := strings.Split(path, "/")
	for i := 0; i < len(rabbitHole); i++ {
		var ok bool
		current, ok = current.children[rabbitHole[i]]
		if !ok {
			return nil
		}
	}
	return current
}
