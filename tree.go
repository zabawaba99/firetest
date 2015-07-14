package firetest

import "strings"

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
