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

	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

var defaultPubSub = NewPubSub(DefaultBufferSize)

// SetDefaultPubSub sets pubsub used by the package to ps.
func SetDefaultPubSub(ps PubSub) {
	defaultPubSub = ps
}

// DefaultPubSub returns the default PubSub.
func DefaultPubSub() PubSub {
	return defaultPubSub
}

// Subscribe adds an event handler to the default event pubsub.
// The name can be a glob in order to catch multiple event types.
// The handler must be non-blocking.
func Subscribe(name string, hdl Handler) error {
	return defaultPubSub.Subscribe(name, hdl)
}

// Unsubscribe removes an event handler from the default event pubsub.
func Unsubscribe(name string, hdl Handler) {
	defaultPubSub.Unsubscribe(name, hdl)
}

// Publish emits events on the default event pubsub.
func Publish(evts ...Event) {
	for _, evt := range evts {
		defaultPubSub.Publish(local(evt).withCaller())
	}
}

// PublishEvent creates an event and emits it on the default event pubsub.
// Event names are dot-separated for namespacing.
// Event identifiers identify the TTN entities that are related to the event.
// System events have nil identifiers.
// Event data will in most cases be marshaled to JSON, but ideally is a proto message.
func PublishEvent(ctx context.Context, name string, identifiers CombinedIdentifiers, data interface{}, visibility ...ttnpb.Right) {
	defaultPubSub.Publish(local(New(ctx, name, identifiers, data, visibility...)).withCaller())
}
