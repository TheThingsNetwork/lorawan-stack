// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

type builder struct {
	definition *definition
	options    []Option
}

func (b *builder) With(options ...Option) Builder {
	extended := &builder{
		definition: b.definition,
	}
	extended.options = append(extended.options, b.options...)
	extended.options = append(extended.options, options...)
	return extended
}

func (b *builder) New(ctx context.Context, opts ...Option) Event {
	evt := &event{
		ctx: ctx,
		innerEvent: ttnpb.Event{
			Name:           b.definition.name,
			Time:           time.Now().UTC(),
			Origin:         hostname,
			CorrelationIDs: CorrelationIDsFromContext(ctx),
		},
	}
	for _, opt := range b.options {
		opt.applyTo(evt)
	}
	for _, opt := range opts {
		opt.applyTo(evt)
	}
	return evt
}

func (b *builder) NewWithIdentifiersAndData(ctx context.Context, ids CombinedIdentifiers, data interface{}) Event {
	e := local(b.New(ctx))
	if ids != nil {
		e.innerEvent.Identifiers = ids.CombinedIdentifiers().GetEntityIdentifiers()
	}
	if data != nil {
		e.data = data
	}
	return e
}

func (b *builder) BindData(data interface{}) Builder {
	return b.With(WithData(data))
}

// Builder is the interface for building events from definitions.
type Builder interface {
	With(opts ...Option) Builder
	New(ctx context.Context, opts ...Option) Event

	// Convenience function for legacy code. Same as New(ctx, WithIdentifiers(ids), WithData(data)).
	NewWithIdentifiersAndData(ctx context.Context, ids CombinedIdentifiers, data interface{}) Event
	// Convenience function for legacy code. Same as With(WithData(data)).
	BindData(data interface{}) Builder
}

// Builders makes it easier to create multiple events at once.
type Builders []Builder

// New returns new events for each builder in the list.
func (bs Builders) New(ctx context.Context, opts ...Option) []Event {
	events := make([]Event, len(bs))
	for i, b := range bs {
		events[i] = b.New(ctx, opts...)
	}
	return events
}
