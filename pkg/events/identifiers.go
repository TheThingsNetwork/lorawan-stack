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

package events

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/unique"
)

type CombinedIdentifiers interface {
	CombinedIdentifiers() *ttnpb.CombinedIdentifiers
}

// IdentifierFilter can be used as a layer on top of a PubSub to filter events
// based on the identifiers they contain.
type IdentifierFilter interface {
	Handler
	Subscribe(ctx context.Context, ids CombinedIdentifiers, handler Handler)
	Unsubscribe(ctx context.Context, ids CombinedIdentifiers, handler Handler)
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

func (f *identifierFilter) Subscribe(ctx context.Context, ids CombinedIdentifiers, handler Handler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, entityIDs := range ids.CombinedIdentifiers().GetEntityIdentifiers() {
		uid := unique.ID(ctx, entityIDs)
		switch entityIDs.Identifiers().(type) {
		case *ttnpb.ApplicationIdentifiers:
			f.applicationIDs[uid] = append(f.applicationIDs[uid], handler)
		case *ttnpb.ClientIdentifiers:
			f.clientIDs[uid] = append(f.clientIDs[uid], handler)
		case *ttnpb.EndDeviceIdentifiers:
			f.deviceIDs[uid] = append(f.deviceIDs[uid], handler)
		case *ttnpb.GatewayIdentifiers:
			f.gatewayIDs[uid] = append(f.gatewayIDs[uid], handler)
		case *ttnpb.OrganizationIdentifiers:
			f.organizationIDs[uid] = append(f.organizationIDs[uid], handler)
		case *ttnpb.UserIdentifiers:
			f.userIDs[uid] = append(f.userIDs[uid], handler)
		}
	}
}

func removeHandler(handlers []Handler, toRemove Handler) []Handler {
	updated := make([]Handler, 0, len(handlers))
	for _, registered := range handlers {
		if registered == toRemove {
			continue
		}
		updated = append(updated, registered)
	}
	return updated
}

func (f *identifierFilter) Unsubscribe(ctx context.Context, ids CombinedIdentifiers, handler Handler) {
	f.mu.Lock()
	defer f.mu.Unlock()
	for _, entityIDs := range ids.CombinedIdentifiers().GetEntityIdentifiers() {
		uid := unique.ID(ctx, entityIDs)
		switch entityIDs.Identifiers().(type) {
		case *ttnpb.ApplicationIdentifiers:
			f.applicationIDs[uid] = removeHandler(f.applicationIDs[uid], handler)
			if len(f.applicationIDs[uid]) == 0 {
				delete(f.applicationIDs, uid)
			}
		case *ttnpb.ClientIdentifiers:
			f.clientIDs[uid] = removeHandler(f.clientIDs[uid], handler)
			if len(f.clientIDs[uid]) == 0 {
				delete(f.clientIDs, uid)
			}
		case *ttnpb.EndDeviceIdentifiers:
			f.deviceIDs[uid] = removeHandler(f.deviceIDs[uid], handler)
			if len(f.deviceIDs[uid]) == 0 {
				delete(f.deviceIDs, uid)
			}
		case *ttnpb.GatewayIdentifiers:
			f.gatewayIDs[uid] = removeHandler(f.gatewayIDs[uid], handler)
			if len(f.gatewayIDs[uid]) == 0 {
				delete(f.gatewayIDs, uid)
			}
		case *ttnpb.OrganizationIdentifiers:
			f.organizationIDs[uid] = removeHandler(f.organizationIDs[uid], handler)
			if len(f.organizationIDs[uid]) == 0 {
				delete(f.organizationIDs, uid)
			}
		case *ttnpb.UserIdentifiers:
			f.userIDs[uid] = removeHandler(f.userIDs[uid], handler)
			if len(f.userIDs[uid]) == 0 {
				delete(f.userIDs, uid)
			}
		}
	}
}

func (f *identifierFilter) Notify(evt Event) {
	var matched []Handler
	f.mu.RLock()
	for _, entityIDs := range evt.Identifiers() {
		switch ids := entityIDs.Identifiers().(type) {
		case *ttnpb.ApplicationIdentifiers:
			matched = append(matched, f.applicationIDs[unique.ID(evt.Context(), ids)]...)
		case *ttnpb.ClientIdentifiers:
			matched = append(matched, f.clientIDs[unique.ID(evt.Context(), ids)]...)
		case *ttnpb.EndDeviceIdentifiers:
			matched = append(matched, f.deviceIDs[unique.ID(evt.Context(), ids)]...)
			matched = append(matched, f.applicationIDs[unique.ID(evt.Context(), ids.ApplicationIdentifiers)]...)
		case *ttnpb.GatewayIdentifiers:
			matched = append(matched, f.gatewayIDs[unique.ID(evt.Context(), ids)]...)
		case *ttnpb.OrganizationIdentifiers:
			matched = append(matched, f.organizationIDs[unique.ID(evt.Context(), ids)]...)
		case *ttnpb.UserIdentifiers:
			matched = append(matched, f.userIDs[unique.ID(evt.Context(), ids)]...)
		}
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
