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

package blacklist

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/errors"
)

var errBlacklistedID = errors.DefineInvalidArgument("blacklisted_id", "the ID `{id}` can not be registered")

// New returns a new Blacklist for the given IDs.
func New(ids ...string) *Blacklist {
	b := &Blacklist{
		blacklisted: make(map[string]struct{}),
	}
	b.Add(ids...)
	return b
}

// Blacklist is a list of IDs that is blacklisted.
type Blacklist struct {
	mu          sync.RWMutex
	blacklisted map[string]struct{}
}

// Add an ID to the blacklist.
func (b *Blacklist) Add(ids ...string) {
	b.mu.Lock()
	for _, id := range ids {
		b.blacklisted[id] = struct{}{}
	}
	b.mu.Unlock()
}

// Contains returns whether the blacklist contains the given ID.
func (b *Blacklist) Contains(id string) bool {
	b.mu.RLock()
	_, found := b.blacklisted[id]
	b.mu.RUnlock()
	return found
}

// Blacklists contains multiple blacklists.
type Blacklists []*Blacklist

// Contains returns whether any of the blacklists contains the given ID.
func (b Blacklists) Contains(id string) bool {
	for _, b := range b {
		if b.Contains(id) {
			return true
		}
	}
	return false
}

// Check the given ID on the builtin blacklist as well as the blacklists that may
// be in the context.
func Check(ctx context.Context, id string) error {
	if builtin.Contains(id) {
		return errBlacklistedID
	}
	if FromContext(ctx).Contains(id) {
		return errBlacklistedID
	}
	return nil
}
