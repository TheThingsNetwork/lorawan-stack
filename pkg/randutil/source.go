// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package randutil

import (
	"math/rand"
	"sync"
)

type LockedSource struct {
	mu  sync.Mutex
	src rand.Source
}

func (r *LockedSource) Int63() (n int64) {
	r.mu.Lock()
	n = r.src.Int63()
	r.mu.Unlock()
	return
}

func (r *LockedSource) Seed(seed int64) {
	r.mu.Lock()
	r.src.Seed(seed)
	r.mu.Unlock()
}

func NewLockedSource(src rand.Source) *LockedSource {
	return &LockedSource{
		src: src,
	}
}
