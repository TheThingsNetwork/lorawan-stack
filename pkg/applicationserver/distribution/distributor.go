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
	"sync"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewDistributor creates a Distributor on top of the provided PubSub.
// The underlying pub/sub subscriptions can timeout if there are no active subscribers.
// A timeout of 0 means the underlying subscriptions never timeout.
func NewDistributor(ctx context.Context, pubsub PubSub, timeout time.Duration) *Distributor {
	return &Distributor{
		ctx:     ctx,
		pubsub:  pubsub,
		timeout: timeout,
	}
}

// Distributor routes upstream traffic through an underlying PubSub.
// Multiple subscribers to the same application share the underlying
// PubSub subscription.
type Distributor struct {
	ctx     context.Context
	pubsub  PubSub
	sets    sync.Map
	timeout time.Duration
}

// SendUp sends traffic to the underlying Pub/Sub.
func (d *Distributor) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return d.pubsub.SendUp(ctx, up)
}

type distributorSet struct {
	set *SubscriptionSet

	init    chan struct{}
	initErr error
}

func (d *Distributor) loadOrCreateSet(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*SubscriptionSet, error) {
	uid := unique.ID(ctx, ids)
	s := &distributorSet{
		init: make(chan struct{}),
	}
	if existing, loaded := d.sets.LoadOrStore(uid, s); loaded {
		exists := existing.(*distributorSet)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-exists.init:
		}
		if exists.initErr != nil {
			return nil, exists.initErr
		}
		return exists.set, nil
	}

	var err error
	defer func() {
		close(s.init)
		if err != nil {
			s.initErr = err
			d.sets.Delete(uid)
		}
	}()

	ctx = log.NewContextWithField(d.ctx, "application_uid", uid)
	ctx, err = unique.WithContext(ctx, uid)
	if err != nil {
		return nil, err
	}

	set := NewSubscriptionSet(ctx, d.timeout)
	go func() {
		<-set.Context().Done()
		d.sets.Delete(uid)
	}()
	go func() {
		ctx := set.Context()
		if err := d.pubsub.Subscribe(ctx, ids, set.SendUp); err != nil {
			log.FromContext(ctx).WithError(err).Warn("Pub/Sub subscription failed")
			set.Cancel(err)
		}
	}()
	s.set = set

	return set, nil
}

// Subscribe creates a subscription in the associated subscription set.
func (d *Distributor) Subscribe(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	s, err := d.loadOrCreateSet(ctx, ids)
	if err != nil {
		return nil, err
	}
	return s.Subscribe(ctx, protocol, &ids)
}
