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
	current := tree.rootNode
	for i := 0; i < len(rabbitHole)-1; i++ {
		step := rabbitHole[i]
		next, ok := current.children[step]
		if !ok {
			next = &node{
				parent:   current,
				children: map[string]*node{},
			}
			current.children[step] = next
		}
		next.value = nil // no long has a value since it now has a child
		current, next = next, nil
	}

	lastPath := rabbitHole[len(rabbitHole)-1]
	current.children[lastPath] = n
	n.parent = current
}

func (tree *treeDB) update(path string, n *node) {
	current := tree.rootNode
	rabbitHole := strings.Split(path, "/")

	for i := 0; i < len(rabbitHole); i++ {
		path := rabbitHole[i]
		if path == "" {
			// prevent against empty strings due to strings.Split
			continue
		}
		next, ok := current.children[path]
		if !ok {
			next = &node{parent: current, children: map[string]*node{}}
			current.children[path] = next
		}
		next.value = nil // no long has a value since it now has a child
		current, next = next, nil
	}

	current.merge(n)
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
