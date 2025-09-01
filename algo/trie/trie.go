// This Trie implementation is designed to support strings that includes
// :param and *splat parameters. Strings that are commonly used for URL
// routing.
package trie

import (
	"errors"
)

func splitParam(remaining string) (string, string) {
	i := 0
	for len(remaining) > i && remaining[i] != '/' && remaining[i] != '.' {
		i++
	}
	return remaining[:i], remaining[i:]
}

type node struct {
	Anchor         any
	Children       map[string]*node
	ChildrenKeyLen int
	ParamChild     *node
	ParamName      string
	SplatChild     *node
	SplatName      string
}

func (n *node) add(path string, anchor any) error {
	if len(path) == 0 {
		// end of the path, set the Anchor
		if n.Anchor != nil {
			return errors.New("node.Anchor already set, duplicated path")
		}
		n.Anchor = anchor
		return nil
	}

	token := path[0:1]
	remaining := path[1:]
	var nextNode *node

	switch token[0] {
	case ':':
		// :param case
		var name string
		name, remaining = splitParam(remaining)
		if n.ParamChild == nil {
			n.ParamChild = &node{}
			n.ParamName = name
		}
		nextNode = n.ParamChild
	case '*':
		// *splat case
		name := remaining
		remaining = ""
		if n.SplatChild == nil {
			n.SplatChild = &node{}
			n.SplatName = name
		}
		nextNode = n.SplatChild
	default:
		// general case
		if n.Children == nil {
			n.Children = map[string]*node{}
			n.ChildrenKeyLen = 1
		}
		if n.Children[token] == nil {
			n.Children[token] = &node{}
		}
		nextNode = n.Children[token]
	}

	return nextNode.add(remaining, anchor)
}

type Match struct {
	// same Anchor as in node
	Anchor any
	// params matched for this result
	Params Params
}

func (n *node) find(path string, ps Params) []Match {
	matches := []Match{}

	// anchor found!
	if n.Anchor != nil && path == "" {
		matches = append(matches, Match{n.Anchor, ps})
	}

	if len(path) == 0 {
		return matches
	}

	// *splat branch
	if n.SplatChild != nil {
		ps.add(n.SplatName, path)
		matches = append(
			matches,
			n.SplatChild.find("", ps)...,
		)
		ps.pop()
	}

	// :param branch
	if n.ParamChild != nil {
		value, remaining := splitParam(path)
		ps.add(n.ParamName, value)
		// fmt.Println("==", ps)
		matches = append(
			matches,
			n.ParamChild.find(remaining, ps)...,
		)
		// ps.pop()
		// fmt.Println("++", ps)
	}

	// main branch
	length := n.ChildrenKeyLen
	if len(path) < length {
		return matches
	}
	token := path[0:length]
	remaining := path[length:]
	if n.Children[token] != nil {
		matches = append(
			matches,
			n.Children[token].find(remaining, ps)...,
		)
	}

	return matches
}

func (n *node) compress() {
	// *splat branch
	if n.SplatChild != nil {
		n.SplatChild.compress()
	}
	// :param branch
	if n.ParamChild != nil {
		n.ParamChild.compress()
	}
	// main branch
	if len(n.Children) == 0 {
		return
	}
	// compressable ?
	canCompress := true
	for _, node := range n.Children {
		if node.Anchor != nil || node.SplatChild != nil || node.ParamChild != nil {
			canCompress = false
		}
	}
	// compress
	if canCompress {
		merged := map[string]*node{}
		for key, node := range n.Children {
			for gdKey, gdNode := range node.Children {
				mergedKey := key + gdKey
				merged[mergedKey] = gdNode
			}
		}
		n.Children = merged
		n.ChildrenKeyLen++
		n.compress()
	} else {
		for _, node := range n.Children {
			node.compress()
		}
	}
}

type Trie struct {
	root *node
}

// Instanciate a Trie.
// routes: path->anchor
func New(routes map[string]any) (*Trie, error) {
	trie := &Trie{&node{}}
	for path, anchor := range routes {
		if path == "" || path[0] != '/' {
			return nil, errors.New("path must begin with '/' in path '" + path + "'")
		}

		if err := trie.root.add(path, anchor); err != nil {
			return nil, err
		}
	}

	// Reduce the size of the tree, must be done after the last add.
	trie.root.compress()

	return trie, nil
}

// Given a path, return all the matchin anchors.
func (trie *Trie) Find(path string) []Match {
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}
	return trie.root.find(path, Params{})
}
