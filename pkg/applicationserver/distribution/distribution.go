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

package distribution

import (
	"context"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// PubSub sends upstream traffic from publishers to subscribers.
type PubSub interface {
	// Publish publishes traffic to the subscribers.
	Publish(context.Context, *ttnpb.ApplicationUp) error
	// Subscribe to the traffic of a specific application.
	Subscribe(context.Context, *ttnpb.ApplicationIdentifiers, func(context.Context, *ttnpb.ApplicationUp) error) error
}

// Distributor sends upstream traffic from publishers to subscribers.
type Distributor interface {
	// Publish publishes traffic to the subscribers.
	Publish(context.Context, *ttnpb.ApplicationUp) error
	// Subscribe to the traffic of a specific application.
	Subscribe(context.Context, string, *ttnpb.ApplicationIdentifiers) (*io.Subscription, error)
}

// RequestDecoupler decouples the security information found in a context
// from the lifetime of the context.
type RequestDecoupler interface {
	FromRequestContext(ctx context.Context) context.Context
}
