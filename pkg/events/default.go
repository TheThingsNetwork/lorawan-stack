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

type noopPubSub struct{}

func (noopPubSub) Publish(Event) {}

func (noopPubSub) Subscribe(context.Context, []string, []*ttnpb.EntityIdentifiers, Handler) error {
	return nil
}

var defaultPubSub PubSub = &noopPubSub{}

// SetDefaultPubSub sets pubsub used by the package to ps.
func SetDefaultPubSub(ps PubSub) {
	defaultPubSub = ps
}

// DefaultPubSub returns the default PubSub.
func DefaultPubSub() PubSub {
	return defaultPubSub
}

// Subscribe subscribes on the default PubSub.
func Subscribe(ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, hdl Handler) error {
	return defaultPubSub.Subscribe(ctx, names, ids, hdl)
}

// Publish emits events on the default event pubsub.
func Publish(evts ...Event) {
	for _, evt := range evts {
		defaultPubSub.Publish(local(evt).withCaller())
		publishes.WithLabelValues(evt.Name()).Inc()
	}
}
