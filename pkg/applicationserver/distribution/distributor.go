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

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/errorcontext"
	"go.thethings.network/lorawan-stack/v3/pkg/log"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// NewDistributor creates a Distributor on top of the provided PubSub.
func NewDistributor(ctx context.Context, pubsub PubSub) *Distributor {
	return &Distributor{
		ctx:    ctx,
		pubsub: pubsub,
	}
}

// Distributor routes upstream traffic through an underlying PubSub.
// Multiple subscribers to the same application share the underlying
// PubSub subscription.
type Distributor struct {
	ctx    context.Context
	pubsub PubSub
	sets   sync.Map
}

// SendUp sends traffic to the underlying Pub/Sub.
func (d *Distributor) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	return d.pubsub.SendUp(ctx, up)
}

func (d *Distributor) loadOrCreateSet(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*set, error) {
	uid := unique.ID(ctx, ids)
	s := &set{
		init:          make(chan struct{}),
		subscribeCh:   make(chan *io.Subscription),
		unsubscribeCh: make(chan *io.Subscription),
		upCh:          make(chan *io.ContextualApplicationUp),
	}
	if existing, loaded := d.sets.LoadOrStore(uid, s); loaded {
		exists := existing.(*set)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-exists.init:
		}
		return exists, nil
	}

	var err error
	defer func() {
		close(s.init)
		if err != nil {
			d.sets.Delete(uid)
		}
	}()

	ctx = log.NewContextWithField(d.ctx, "application_uid", uid)
	ctx, err = unique.WithContext(ctx, uid)
	if err != nil {
		return nil, err
	}
	s.ctx, s.cancel = errorcontext.New(ctx)

	go func() {
		if err := d.pubsub.Subscribe(s.ctx, ids, s.SendUp); err != nil {
			log.FromContext(s.ctx).WithError(err).Warn("Pub/Sub subscription failed")
			s.cancel(err)
		}
	}()

	go s.run()

	return s, nil
}

// Subscribe creates a subscription in the associated subscription set.
func (d *Distributor) Subscribe(ctx context.Context, protocol string, ids ttnpb.ApplicationIdentifiers) (*io.Subscription, error) {
	s, err := d.loadOrCreateSet(ctx, ids)
	if err != nil {
		return nil, err
	}
	return s.Subscribe(ctx, protocol, ids)
}
