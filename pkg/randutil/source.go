// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package randutil

import (
	"math/rand"
	"sync"
)

// LockedSource is a rand.Source, which is safe for concurrent use. Adapted from the non-exported
// lockedSource from stdlib rand.
type LockedSource struct {
	mu  sync.Mutex
	src rand.Source
	s64 rand.Source64 // non-nil if src is source64
}

// Int63 returns a non-negative pseudo-random 63-bit integer as an int64.
func (r *LockedSource) Int63() (n int64) {
	r.mu.Lock()
	n = r.src.Int63()
	r.mu.Unlock()
	return
}

// Uint64 returns a pseudo-random 64-bit value as a uint64.
func (r *lockedSource) Uint64() (n uint64) {
	if r.s64 != nil {
		r.lk.Lock()
		n = r.s64.Uint64()
		r.lk.Unlock()
		return
	}
	return uint64(r.Int63())>>31 | uint64(r.Int63())<<32
}

// Seed uses the provided seed value to initialize the generator to a deterministic state.
// Seed should not be called concurrently with any other Rand method.
func (r *LockedSource) Seed(seed int64) {
	r.mu.Lock()
	r.src.Seed(seed)
	r.mu.Unlock()
}

// NewLockedSource returns a rand.Source, which is safe for concurrent use.
func NewLockedSource(src rand.Source) *LockedSource {
	return &LockedSource{
		src: src,
	}
}
