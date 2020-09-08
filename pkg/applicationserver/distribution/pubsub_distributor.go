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
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
)

// NewPubSubDistributor creates a Distributor on top of the provided PubSub.
// The underlying subscription sets can timeout if there are no active subscribers.
// A timeout of 0 means the underlying subscription sets never timeout.
func NewPubSubDistributor(ctx context.Context, timeout time.Duration, pubsub PubSub) Distributor {
	return &pubSubDistributor{
		pubsub:        pubsub,
		subscriptions: newSubscriptionMap(ctx, timeout, subscribeSetToPubSub(pubsub)),
	}
}

type pubSubDistributor struct {
	pubsub        PubSub
	subscriptions *subscriptionMap
}

// Publish publishes traffic to the underlying Pub/Sub.
func (d *pubSubDistributor) Publish(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return d.pubsub.Publish(ctx, up)
}

var errMissingIdentifiers = errors.DefineFailedPrecondition("missing_identifiers", "subscriptions without identifiers are not supported")

// Subscribe creates a subscription in the associated subscription set.
func (d *pubSubDistributor) Subscribe(ctx context.Context, protocol string, ids *ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	if ids == nil {
		return nil, errMissingIdentifiers.New()
	}
	s, err := d.subscriptions.LoadOrCreate(ctx, *ids)
	if err != nil {
		return nil, err
	}
	return s.Subscribe(ctx, protocol, ids)
}

func subscribeSetToPubSub(pubsub PubSub) func(*subscriptionSet, ttnpb.ApplicationIdentifiers) error {
	return func(set *subscriptionSet, ids ttnpb.ApplicationIdentifiers) error {
		go func() {
			ctx := set.Context()
			if err := pubsub.Subscribe(ctx, ids, set.Publish); err != nil {
				log.FromContext(ctx).WithError(err).Warn("Pub/Sub subscription failed")
				set.Cancel(err)
			}
		}()
		return nil
	}
}
