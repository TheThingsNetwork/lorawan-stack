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
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NewLocalDistributor creates a Distributor that routes the traffic locally.
// The underlying subscription sets can timeout if there are no active subscribers.
// A timeout of 0 means the underlying subscriptions never timeout.
func NewLocalDistributor(ctx context.Context, rd RequestDecoupler, timeout time.Duration, broadcastOpts []io.SubscriptionOption, mapOpts []io.SubscriptionOption) Distributor {
	return &localDistributor{
		broadcast:     newSubscriptionSet(ctx, rd, 0, broadcastOpts...),
		subscriptions: newSubscriptionMap(ctx, rd, timeout, noSetup, mapOpts...),
	}
}

type localDistributor struct {
	broadcast     *subscriptionSet
	subscriptions *subscriptionMap
}

// Publish publishes traffic to the underlying subscriptions.
func (d *localDistributor) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	if err := d.broadcast.Publish(ctx, up); err != nil {
		return err
	}
	set, err := d.subscriptions.Load(ctx, up.EndDeviceIds.ApplicationIds)
	if err != nil && !errors.IsNotFound(err) {
		return err
	} else if err == nil {
		return set.Publish(ctx, up)
	}
	return nil
}

// Subscribe creates a subscription in the associated subscription set. If the identifiers are nil,
// the subscription receives all of the traffic sent to the Distributor.
func (d *localDistributor) Subscribe(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	if ids == nil {
		return d.broadcast.Subscribe(ctx, protocol, ids)
	}
	s, err := d.subscriptions.LoadOrCreate(ctx, ids)
	if err != nil {
		return nil, err
	}
	return s.Subscribe(ctx, protocol, ids)
}
