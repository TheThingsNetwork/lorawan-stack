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

package events

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

// IdentifierFilter can be used as a layer on top of a PubSub to filter events
// based on the identifiers they contain.
type IdentifierFilter interface {
	Handler
	Subscribe(ctx context.Context, ids ttnpb.Identifiers, handler Handler)
	Unsubscribe(ctx context.Context, ids ttnpb.Identifiers, handler Handler)
}

// NewIdentifierFilter returns a new IdentifierFilter (see interface).
func NewIdentifierFilter() IdentifierFilter {
	return &identifierFilter{
		applicationIDs:  make(map[string][]Handler),
		clientIDs:       make(map[string][]Handler),
		deviceIDs:       make(map[string][]Handler),
		gatewayIDs:      make(map[string][]Handler),
		organizationIDs: make(map[string][]Handler),
		userIDs:         make(map[string][]Handler),
	}
}

// identifierFilter is a simple implementation of the IdentifierFilter.
//
// It uses a single RWMutex to protect the maps as well as the slices that are
// contained within those maps. In the future it could be optimized by batching
// subscribes and unsubscribes into a single Lock/Unlock.
//
// Using a sync.Map will likely not help much here, as we also need to protect
// the slice from concurrent writes. Additionally it creates more complexity
// when cleaning up (deleting empty []Handler from the map).
type identifierFilter struct {
	mu              sync.RWMutex
	applicationIDs  map[string][]Handler
	clientIDs       map[string][]Handler
	deviceIDs       map[string][]Handler
	gatewayIDs      map[string][]Handler
	organizationIDs map[string][]Handler
	userIDs         map[string][]Handler
}

func (f *identifierFilter) Subscribe(ctx context.Context, ids ttnpb.Identifiers, handler Handler) {
	cids := ids.CombinedIdentifiers()
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, id := range cids.ApplicationIDs {
		uid := unique.ID(ctx, id)
		f.applicationIDs[uid] = append(f.applicationIDs[uid], handler)
	}
	for _, id := range cids.ClientIDs {
		uid := unique.ID(ctx, id)
		f.clientIDs[uid] = append(f.clientIDs[uid], handler)
	}
	for _, id := range cids.DeviceIDs {
		uid := unique.ID(ctx, id)
		f.deviceIDs[uid] = append(f.deviceIDs[uid], handler)
	}
	for _, id := range cids.GatewayIDs {
		uid := unique.ID(ctx, id)
		f.gatewayIDs[uid] = append(f.gatewayIDs[uid], handler)
	}
	for _, id := range cids.OrganizationIDs {
		uid := unique.ID(ctx, id)
		f.organizationIDs[uid] = append(f.organizationIDs[uid], handler)
	}
	for _, id := range cids.UserIDs {
		uid := unique.ID(ctx, id)
		f.userIDs[uid] = append(f.userIDs[uid], handler)
	}
}

func (f *identifierFilter) Unsubscribe(ctx context.Context, ids ttnpb.Identifiers, handler Handler) {
	cids := ids.CombinedIdentifiers()
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, id := range cids.ApplicationIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.applicationIDs[uid] {
			if registered == handler {
				f.applicationIDs[uid] = append(f.applicationIDs[uid][:i], f.applicationIDs[uid][i+1:]...)
				if len(f.applicationIDs[uid]) == 0 {
					delete(f.applicationIDs, uid)
				}
				break
			}
		}
	}
	for _, id := range cids.ClientIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.clientIDs[uid] {
			if registered == handler {
				f.clientIDs[uid] = append(f.clientIDs[uid][:i], f.clientIDs[uid][i+1:]...)
				if len(f.clientIDs[uid]) == 0 {
					delete(f.clientIDs, uid)
				}
				break
			}
		}
	}
	for _, id := range cids.DeviceIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.deviceIDs[uid] {
			if registered == handler {
				f.deviceIDs[uid] = append(f.deviceIDs[uid][:i], f.deviceIDs[uid][i+1:]...)
				if len(f.deviceIDs[uid]) == 0 {
					delete(f.deviceIDs, uid)
				}
				break
			}
		}
	}
	for _, id := range cids.GatewayIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.gatewayIDs[uid] {
			if registered == handler {
				f.gatewayIDs[uid] = append(f.gatewayIDs[uid][:i], f.gatewayIDs[uid][i+1:]...)
				if len(f.gatewayIDs[uid]) == 0 {
					delete(f.gatewayIDs, uid)
				}
				break
			}
		}
	}
	for _, id := range cids.OrganizationIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.organizationIDs[uid] {
			if registered == handler {
				f.organizationIDs[uid] = append(f.organizationIDs[uid][:i], f.organizationIDs[uid][i+1:]...)
				if len(f.organizationIDs[uid]) == 0 {
					delete(f.organizationIDs, uid)
				}
				break
			}
		}
	}
	for _, id := range cids.UserIDs {
		uid := unique.ID(ctx, id)
		for i, registered := range f.userIDs[uid] {
			if registered == handler {
				f.userIDs[uid] = append(f.userIDs[uid][:i], f.userIDs[uid][i+1:]...)
				if len(f.userIDs[uid]) == 0 {
					delete(f.userIDs, uid)
				}
				break
			}
		}
	}
}

func (f *identifierFilter) Notify(evt Event) {
	var matched []Handler
	ids := evt.Identifiers()
	if ids == nil {
		return
	}
	f.mu.RLock()
	for _, id := range ids.ApplicationIDs {
		matched = append(matched, f.applicationIDs[unique.ID(evt.Context(), id)]...)
	}
	for _, id := range ids.ClientIDs {
		matched = append(matched, f.clientIDs[unique.ID(evt.Context(), id)]...)
	}
	for _, id := range ids.DeviceIDs {
		matched = append(matched, f.deviceIDs[unique.ID(evt.Context(), id)]...)
		matched = append(matched, f.applicationIDs[unique.ID(evt.Context(), id.ApplicationIdentifiers)]...)
	}
	for _, id := range ids.GatewayIDs {
		matched = append(matched, f.gatewayIDs[unique.ID(evt.Context(), id)]...)
	}
	for _, id := range ids.OrganizationIDs {
		matched = append(matched, f.organizationIDs[unique.ID(evt.Context(), id)]...)
	}
	for _, id := range ids.UserIDs {
		matched = append(matched, f.userIDs[unique.ID(evt.Context(), id)]...)
	}
	f.mu.RUnlock()
	notified := make(map[Handler]struct{}, len(matched))
	for _, handler := range matched {
		if _, ok := notified[handler]; ok {
			continue
		}
		handler.Notify(evt)
		notified[handler] = struct{}{}
	}
}
