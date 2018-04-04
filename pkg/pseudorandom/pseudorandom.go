// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package pseudorandom

import (
	"math/rand"
	"sync"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/random"
)

// TTNRandom is used as a wrapper around math/rand.
type TTNRandom struct {
	mu sync.Mutex
	random.Interface
}

// New returns a new Random, in most cases you can also just use the global funcs.
func New(seed int64) random.Interface {
	return &TTNRandom{
		Interface: &random.TTNRandom{
			Source: rand.New(rand.NewSource(seed)),
		},
	}
}

var global = New(time.Now().UnixNano())

// Intn returns random int with max n. This func uses the global TTNRandom.
func Intn(n int) int { return global.Intn(n) }

// Intn returns random int with max n.
func (r *TTNRandom) Intn(n int) int {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.Intn(n)
}

// String returns random string of length n. This func uses the global TTNRandom.
func String(n int) string { return global.String(n) }

// String returns random string of length n.
func (r *TTNRandom) String(n int) string {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.String(n)
}

// Bytes generates a random byte slice of length n. This func uses the global TTNRandom.
func Bytes(n int) []byte { return global.Bytes(n) }

// Bytes generates a random byte slice of length n.
func (r *TTNRandom) Bytes(n int) []byte {
	r.mu.Lock()
	defer r.mu.Unlock()
	return r.Interface.Bytes(n)
}

// FillBytes fills the byte slice with random bytes. This func uses the global TTNRandom.
func FillBytes(p []byte) { global.FillBytes(p) }

// FillBytes fills the byte slice with random bytes.
func (r *TTNRandom) FillBytes(p []byte) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.Interface.FillBytes(p)
}
