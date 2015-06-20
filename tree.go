package firetest

import (
	"encoding/json"
	"fmt"
	"log"
	"strings"
)

type node struct {
	value    interface{}
	children map[string]*node
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
