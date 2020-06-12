// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

// Package randutil provides pseudo-random number generator utilities.
package randutil

import (
	"math/rand"
	"sync"
)

// LockedRand wraps rand.Rand with a mutex.
type LockedRand struct {
	mu   *sync.Mutex
	rand *rand.Rand
}

// Read is like rand.Rand.Read, but safe for concurrent use.
func (r *LockedRand) Read(b []byte) (int, error) {
	r.mu.Lock()
	n, err := r.rand.Read(b)
	r.mu.Unlock()
	return n, err
}

// Int31 calls r.Rand.Int31 and is safe for concurrent use.
func (r *LockedRand) Int31() int32 {
	r.mu.Lock()
	v := r.rand.Int31()
	r.mu.Unlock()
	return v
}

// Int63 calls r.Rand.Int63 and is safe for concurrent use.
func (r *LockedRand) Int63() int64 {
	r.mu.Lock()
	v := r.rand.Int63()
	r.mu.Unlock()
	return v
}

// Intn calls r.Rand.Intn and is safe for concurrent use.
func (r *LockedRand) Intn(n int) int {
	r.mu.Lock()
	v := r.rand.Intn(n)
	r.mu.Unlock()
	return v
}

// Uint32 calls r.Rand.Uint32 and is safe for concurrent use.
func (r *LockedRand) Uint32() uint32 {
	r.mu.Lock()
	v := r.rand.Uint32()
	r.mu.Unlock()
	return v
}

// Float32 calls r.Rand.Float32 and is safe for concurrent use.
func (r *LockedRand) Float32() float32 {
	r.mu.Lock()
	v := r.rand.Float32()
	r.mu.Unlock()
	return v
}

// Float64 calls r.Rand.Float64 and is safe for concurrent use.
func (r *LockedRand) Float64() float64 {
	r.mu.Lock()
	v := r.rand.Float64()
	r.mu.Unlock()
	return v
}

// NewLockedRand returns a new rand.Rand, which is safe for concurrent use.
func NewLockedRand(src rand.Source) *LockedRand {
	return &LockedRand{
		mu:   &sync.Mutex{},
		rand: rand.New(src),
	}
}
