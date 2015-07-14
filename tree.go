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
			current = &node{
				parent:   previous,
				children: map[string]*node{},
			}
			previous.children[step] = current
		}
		current.value = nil // no long has a value since it now has a child
		previous, current = current, nil
	}

	lastPath := rabbitHole[len(rabbitHole)-1]
	previous.children[lastPath] = n
	n.parent = previous
}

func (tree *treeDB) del(path string) {
	if path == "" {
		tree.rootNode = &node{
			children: map[string]*node{},
		}
		return
	}

	rabbitHole := strings.Split(path, "/")
	current := tree.rootNode

	// traverse to target node's parent
	var delIdx int
	for ; delIdx < len(rabbitHole)-1; delIdx++ {
		next, ok := current.children[rabbitHole[delIdx]]
		if !ok {
			// item does not exist, no need to do anything
			return
		}

		current = next
	}

	endNode := current
	leafPath := rabbitHole[len(rabbitHole)-1]
	delete(endNode.children, leafPath)

	for tmp := endNode.prune(); tmp != nil; tmp = tmp.prune() {
		delIdx--
		endNode = tmp
	}

	if endNode != nil {
		delete(endNode.children, rabbitHole[delIdx])
	}
}

func (tree *treeDB) get(path string) *node {
	current := tree.rootNode
	if path == "" {
		return current
	}

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
