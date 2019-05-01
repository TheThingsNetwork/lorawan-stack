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

package io

import (
	"sort"
	"sync"
	"time"
)

type rtts struct {
	count int
	mu    sync.RWMutex
	items []time.Duration
}

func newRTTs(count int) *rtts {
	return &rtts{
		count: count,
		items: make([]time.Duration, 0, count+1),
	}
}

// Record records the given round-trip time.
func (r *rtts) Record(d time.Duration) {
	r.mu.Lock()
	r.items = append(r.items, d)
	if len(r.items) > r.count {
		r.items = append(r.items[:0], r.items[len(r.items)-r.count:]...)
	}
	r.mu.Unlock()
}

// Last returns the last measured round-trip time.
func (r *rtts) Last() (time.Duration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.items) == 0 {
		return 0, false
	}
	return r.items[len(r.items)-1], true
}

// Stats returns the min, max, median and number of recorded round-trip times.
func (r *rtts) Stats() (min, max, median time.Duration, count int) {
	r.mu.RLock()
	sorted := append(make([]time.Duration, 0, len(r.items)), r.items...)
	r.mu.RUnlock()
	if len(sorted) == 0 {
		return
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i] < sorted[j] })
	min = sorted[0]
	max = sorted[len(sorted)-1]
	if l := len(sorted); l%2 == 0 {
		median = (sorted[l/2-1] + sorted[l/2]) / 2
	} else {
		median = sorted[l/2]
	}
	count = len(sorted)
	return
}
