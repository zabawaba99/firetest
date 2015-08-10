package firetest

import (
	"encoding/json"
	"strings"
	"sync"
	"time"
)

type event struct {
	Name string
	Data eventData
}

type eventData struct {
	Path string `json:"path"`
	Data *node  `json:"data"`
}

func (ed eventData) MarshalJSON() ([]byte, error) {
	type eventData2 eventData
	ed2 := eventData2(ed)
	ed2.Path = "/" + ed2.Path
	return json.Marshal(ed2)
}

func newEvent(name, path string, n *node) event {
	return event{
		Name: "put",
		Data: eventData{
			Path: path,
			Data: n,
		},
	}
}

type treeDB struct {
	rootNode *node

	watchersMtx sync.RWMutex
	watchers    map[string][]chan event
}

func newTree() *treeDB {
	return &treeDB{
		rootNode: &node{
			children: map[string]*node{},
		},
		watchers: map[string][]chan event{},
	}
}

func (tree *treeDB) add(path string, n *node) {
	defer func() { go tree.notify(newEvent("put", path, n)) }()
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

	go tree.notify(newEvent("patch", path, n))
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

	go tree.notify(newEvent("put", path, nil))
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

func (tree *treeDB) notify(e event) {
	tree.watchersMtx.RLock()
	for path, listeners := range tree.watchers {
		if !strings.HasPrefix(e.Data.Path, path) {
			continue
		}

		for _, c := range listeners {
			select {
			case c <- e:
			case <-time.After(250 * time.Millisecond):
				continue
			}
		}
	}
	tree.watchersMtx.RUnlock()
}

func (tree *treeDB) stopWatching(path string, c chan event) {
	tree.watchersMtx.Lock()
	index := -1
	for i, ch := range tree.watchers[path] {
		if ch == c {
			index = i
			break
		}
	}

	if index > -1 {
		a := tree.watchers[path]
		tree.watchers[path] = append(a[:index], a[index+1:]...)
		close(c)
	}
	tree.watchersMtx.Unlock()
}

func (tree *treeDB) watch(path string) chan event {
	c := make(chan event)

	tree.watchersMtx.Lock()
	tree.watchers[path] = append(tree.watchers[path], c)
	tree.watchersMtx.Unlock()

	return c
}
