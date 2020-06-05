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

type rttItem struct {
	d time.Duration
	t time.Time
}

type rtts struct {
	count int
	ttl   time.Duration
	mu    sync.RWMutex
	items []rttItem
}

func newRTTs(count int, ttl time.Duration) *rtts {
	return &rtts{
		count: count,
		ttl:   ttl,
		items: make([]rttItem, 0, count+1),
	}
}

// Record records the given round-trip time.
func (r *rtts) Record(d time.Duration, t time.Time) {
	r.mu.Lock()
	r.items = append(r.items, rttItem{d, t})
	if len(r.items) > r.count {
		r.items = append(r.items[:0], r.items[1:]...)
	}
	r.mu.Unlock()
}

// Last returns the last measured round-trip time.
func (r *rtts) Last(ref time.Time) (time.Duration, bool) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	if len(r.items) == 0 || ref.Sub(r.items[len(r.items)-1].t) > r.ttl {
		return 0, false
	}
	return r.items[len(r.items)-1].d, true
}

// Stats returns the min, max, median, requested percentile and number of recorded round-trip times.
func (r *rtts) Stats(percentile int, ref time.Time) (min, max, median, np time.Duration, count int) {
	r.mu.RLock()
	sorted := make([]rttItem, 0, len(r.items))
	for i, item := range r.items {
		if ref.Sub(item.t) <= r.ttl {
			sorted = append(sorted, r.items[i:]...)
			break
		}
	}
	r.mu.RUnlock()
	if len(sorted) == 0 {
		return
	}
	sort.Slice(sorted, func(i, j int) bool { return sorted[i].d < sorted[j].d })
	min = sorted[0].d
	max = sorted[len(sorted)-1].d
	if l := len(sorted); l%2 == 0 {
		median = (sorted[l/2-1].d + sorted[l/2].d) / 2
	} else {
		median = sorted[l/2].d
	}
	npi := int(float32(percentile)/100.0*float32(len(sorted))) - 1
	if npi < 0 {
		npi = 0
	}
	np = sorted[npi].d
	count = len(sorted)
	return
}
