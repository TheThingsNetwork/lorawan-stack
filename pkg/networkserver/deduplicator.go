// Copyright © 2018 The Things Network
// Use of this source code is governed by the MIT license that can be found in the LICENSE file.

package networkserver

import (
	"sync"
	"time"
)

const collectionCapacity = 8

type collection struct {
	sync.Mutex
	values []interface{}
}

func newCollection(elements ...interface{}) *collection {
	n := len(elements)
	if n < collectionCapacity {
		n = collectionCapacity
	}

	c := &collection{
		values: make([]interface{}, 0, n),
	}
	c.values = append(c.values, elements...)
	return c
}

func (c *collection) Add(value interface{}) {
	c.Lock()
	c.values = append(c.values, value)
	c.Unlock()
}

func (c *collection) Reset() {
	c.Lock()
	c.values = make([]interface{}, 0, len(c.values))
	c.Unlock()
}

type deduplicator struct {
	*sync.Map
	deduplicationWindow time.Duration
	cooldownWindow      time.Duration
	collectionPool      *sync.Pool
	timerPool           *sync.Pool
}

// Deduplicate deduplicates the values, associated to key.
// If the goroutine is not first to assign a value to the key, the value is appened to already
// stored value and a nil-slice and true is returned.
// Otherwise, the function blocks for duration of at least the configured deduplication window
// and returns the accumulated values and false.
// The underlying map panics if key is of []byte type.
func (d *deduplicator) Deduplicate(key interface{}, value interface{}) (dups []interface{}, isDup bool) {
	nc := d.collectionPool.Get().(*collection)

	lv, isDup := d.Map.LoadOrStore(key, nc)
	c := lv.(*collection)
	c.Add(value)

	if isDup {
		d.collectionPool.Put(nc)
		return nil, true
	}

	t := d.timerPool.Get().(*time.Timer)
	t.Reset(d.deduplicationWindow)
	<-t.C

	dups = c.values

	go func() {
		t.Reset(d.cooldownWindow)
		<-t.C
		d.timerPool.Put(t)

		d.Map.Delete(key)
		c.Reset()
		d.collectionPool.Put(c)
	}()

	return dups, false
}

func newDeduplicator(deduplicationWindow, cooldownWindow time.Duration) *deduplicator {
	d := &deduplicator{
		Map:                 &sync.Map{},
		deduplicationWindow: deduplicationWindow,
		cooldownWindow:      cooldownWindow,
		timerPool:           &sync.Pool{},
		collectionPool:      &sync.Pool{},
	}

	d.collectionPool.New = func() interface{} { return newCollection() }
	d.timerPool.New = func() interface{} {
		t := time.NewTimer(time.Duration(0))
		if !t.Stop() {
			<-t.C
		}
		return t
	}
	return d
}
