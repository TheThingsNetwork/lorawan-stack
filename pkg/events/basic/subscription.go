// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package basic

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/events"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// Subscription is a basic implementation of a PubSub subscription.
type Subscription struct {
	ctx         context.Context
	names       []string
	identifiers []*ttnpb.EntityIdentifiers
	handler     events.Handler
}

// NewSubscription creates a new basic PubSub subscription.
func NewSubscription(ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl events.Handler) (*Subscription, error) {
	s := &Subscription{
		ctx:         ctx,
		names:       names,
		identifiers: identifiers,
		handler:     hdl,
	}
	return s, nil
}

func (s *Subscription) matchName(evt events.Event) bool {
	if len(s.names) == 0 {
		return true
	}
	for _, subName := range s.names {
		if subName == evt.Name() {
			return true
		}
	}
	return false
}

func (s *Subscription) matchIdentifiers(evt events.Event) bool {
	if len(s.identifiers) == 0 {
		return true
	}
	definition := events.GetDefinition(evt)
	for _, evtIDs := range evt.Identifiers() {
		evtEntityType := evtIDs.EntityType()
		for _, subIDs := range s.identifiers {
			subEntityType, subUID := subIDs.EntityType(), unique.ID(s.ctx, subIDs)
			if evtEntityType == subEntityType && unique.ID(evt.Context(), evtIDs) == subUID {
				return true
			}
			if evtEntityType == "end device" && subEntityType == "application" &&
				unique.ID(evt.Context(), evtIDs.GetDeviceIds().ApplicationIds) == unique.ID(s.ctx, subIDs) &&
				definition != nil && definition.PropagateToParent() {
				return true
			}
		}
	}
	return false
}

// Match returns whether the event matches the subscription.
func (s *Subscription) Match(evt events.Event) bool {
	if s == nil {
		return false
	}
	return s.matchName(evt) && s.matchIdentifiers(evt)
}

// Notify notifies the subscription of a new matching event.
func (s *Subscription) Notify(evt events.Event) {
	if s == nil {
		return
	}
	s.handler.Notify(evt)
}
