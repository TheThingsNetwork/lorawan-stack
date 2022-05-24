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

package events

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// Store extends PubSub implementations with storage of historical events.
type Store interface {
	PubSub
	// FindRelated finds events with matching correlation IDs.
	FindRelated(ctx context.Context, correlationID string) ([]Event, error)
	// FetchHistory fetches the tail (optional) of historical events matching the
	// given names (optional) and identifiers (mandatory) after the given time (optional).
	FetchHistory(
		ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int,
	) ([]Event, error)
	// SubscribeWithHistory is like FetchHistory, but after fetching historical events,
	// this continues sending live events until the context is done.
	SubscribeWithHistory(
		ctx context.Context, names []string, ids []*ttnpb.EntityIdentifiers, after *time.Time, tail int, hdl Handler,
	) error
}
