package ecsgen

import (
	"errors"
	"sort"
	"strings"
)

// Root defines the top level namespace within an ECS schema. It is the top level
// data structure used to create the schema tree.
type Root struct {
	// TopLevel is the top level namespace for nested objects
	TopLevel map[string]*Node

	// Index holds references to each node by absolute path
	Index map[string]*Node
}

// NewRoot creates an empty Root.
func NewRoot() *Root {
	return &Root{
		TopLevel: map[string]*Node{},
		Index:    map[string]*Node{},
	}
}

// Branch is used to resolve Nodes within the tree. It will create all
// previously unknown Node's within the graph to traverse to the specified path.
// For example, if you passed "client.as.organization.name", it would perform the
// following lookups: Node("client").Child("as").Child("organization").Child("name").
func (r *Root) Branch(branchpath string) *Node {
	if branchpath == "" {
		panic(errors.New("cannot have an empty branch path"))
	}

	// short circuit if the provided path is a top level object
	if !strings.Contains(branchpath, ".") {
		if node, found := r.TopLevel[branchpath]; found {
			// top level object already exists, return it
			return node
		}

		// create the new top level object
		node := &Node{
			Name:     branchpath,
			Path:     branchpath,
			Root:     r,
			Children: map[string]*Node{},
		}

		// add it to the index
		r.Index[branchpath] = node

		// add it to the top level tree
		r.TopLevel[branchpath] = node

		return node
	}

	var node *Node
	// we need to walk the path. lets start by splitting the path
	// into their individual elements.
	// i.e. "client.as.organization.name" => ["client", "as", "organization", "name"]
	pathelms := strings.Split(branchpath, ".")

	// get the root
	node = r.Branch(pathelms[0])

	// enumerate all children
	for i := 1; i < len(pathelms); i++ {
		node = node.Child(pathelms[i])
	}

	return node
}

// ListChildren implements the Walkable interface.
func (r *Root) ListChildren() <-chan *Node {
	// create the return channel, close it once we're done
	ret := make(chan *Node, len(r.TopLevel))
	defer close(ret)

	// short circuit if we've got no elements
	if len(r.TopLevel) == 0 {
		return ret
	}

	// get a list of child names from the map
	// and sort them
	keys := []string{}
	for k := range r.TopLevel {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	// populate the channel in a predictable order
	for _, k := range keys {
		ret <- r.TopLevel[k]
	}

	return ret
}
