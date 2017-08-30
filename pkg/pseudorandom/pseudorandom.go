// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package pseudorandom

import (
	"math/rand"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/random"
)

// TTNRandom is used as a wrapper around math/rand
type TTNRandom struct {
	mu sync.Mutex
	random.Interface
}

// New returns a new Random, in most cases you can also just use the global funcs
func New(seed int64) random.Interface {
	return &TTNRandom{
		Interface: &random.TTNRandom{
			Source: rand.New(rand.NewSource(seed)),
		},
	}
}

var global = New(time.Now().UnixNano())

// Intn returns random int with max n
func Intn(n int) int { return global.Intn(n) }
func (r *TTNRandom) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.Intn(n)
}

// String returns random string of length n
func String(n int) string { return global.String(n) }
func (r *TTNRandom) String(n int) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.String(n)
}

// Bytes generates a random byte slice of length n
func Bytes(n int) []byte { return global.Bytes(n) }
func (r *TTNRandom) Bytes(n int) []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.Bytes(n)
}

// FillBytes fills the byte slice with random bytes. It does not use an intermediate buffer
func FillBytes(p []byte) { global.FillBytes(p) }
func (r *TTNRandom) FillBytes(p []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Interface.FillBytes(p)
}
