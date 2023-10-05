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

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// PubSub interface combines the Publisher and Subscriber interfaces.
type PubSub interface {
	Publisher
	Subscriber
}

// Publisher interface lets you publish events.
type Publisher interface {
	// Publish emits an event on the default event pubsub.
	Publish(evs ...Event)
}

// Subscriber interface lets you subscribe to events.
type Subscriber interface {
	// Subscribe to events that match the names and identifiers.
	// The subscription continues until the context is canceled.
	// Events matching _any_ of the names or identifiers will get sent to the handler.
	// The handler must be non-blocking.
	Subscribe(ctx context.Context, names []string, identifiers []*ttnpb.EntityIdentifiers, hdl Handler) error
}

// Subscription is the interface for PubSub subscriptions.
type Subscription interface {
	// Match returns whether the event matches the subscription.
	Match(Event) bool
	// Notify notifies the subscription of a new matching event.
	Notify(Event)
}

// PublishFunc is a function which implements Publisher.
type PublishFunc func(...Event)

// Publish implements events.Publisher.
func (f PublishFunc) Publish(evs ...Event) { f(evs...) }
