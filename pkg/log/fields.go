// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package log

import (
	"errors"
	"fmt"
)

// Fielder interface check.
var _ Fielder = &F{}

// ErrInvalidPairs occurs when not passing an even amount of arguments to a function that expects a list of key-value pairs.
var ErrInvalidPairs = errors.New("Uneven number of key-value pairs passed")

// Fielder is the interface for anything that can have fields.
type Fielder interface {
	Fields() map[string]interface{}
}

// Fields returns a new immutable fields structure.
func Fields(pairs ...interface{}) *F {
	nodes, err := pairsToMap(pairs...)
	if err != nil {
		panic(err)
	}

	return &F{
		nodes: nodes,
	}
}

// F is a Fielder that uses structural sharing to avoid copying entries.
// Setting a key is O(1), getting a key is O(n) (where n is the number of entries),
// but we only use this to accumulate fields so that's ok.
type F struct {
	parent *F
	nodes  map[string]interface{}
}

// Get returns the key from the fields in O(n), where n is the number of entries.
func (f *F) Get(key string) (interface{}, bool) {
	val, ok := f.nodes[key]
	if ok {
		return val, true
	}

	if f.parent != nil {
		return f.parent.Get(key)
	}

	return nil, false
}

func pairsToMap(pairs ...interface{}) (map[string]interface{}, error) {
	nodes := make(map[string]interface{})
	var key string

	if len(pairs)%2 != 0 {
		return nil, ErrInvalidPairs
	}

	for i, node := range pairs {
		if i%2 == 0 {
			key = fmt.Sprintf("%v", node)
		} else {
			nodes[key] = node
		}
	}

	return nodes, nil
}

// Fields implements Fielder. Returns all fields in O(n), where n is the number of entries in the map.
func (f *F) Fields() map[string]interface{} {
	var r map[string]interface{}

	if f.parent != nil {
		r = f.parent.Fields()
	} else {
		r = make(map[string]interface{})
	}

	for k, v := range f.nodes {
		r[k] = v
	}

	return r
}

// With returns a new F that has the fields in nodes.
func (f *F) With(nodes map[string]interface{}) *F {
	return &F{
		parent: f,
		nodes:  nodes,
	}
}

// WithField returns a new fielder that has the key set to value.
func (f *F) WithField(name string, val interface{}) *F {
	nodes := map[string]interface{}{
		name: val,
	}

	return f.With(nodes)
}

// WithFields returns a new fielder that has all the fields of the other fielder.
func (f *F) WithFields(fields Fielder) *F {
	return f.With(fields.Fields())
}

// WithError returns new fields that contain the passed error and all its fields (if any).
func (f *F) WithError(err error) *F {
	var fields map[string]interface{}
	if f, ok := err.(Fielder); ok {
		fields = f.Fields()
	} else {
		fields = make(map[string]interface{})
	}

	fields["error"] = err

	return f.With(fields)
}
