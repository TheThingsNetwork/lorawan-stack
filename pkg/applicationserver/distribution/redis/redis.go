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

package redis

import (
	"context"
	"sync"

	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io"
	"go.thethings.network/lorawan-stack/v3/pkg/log"

	ttnredis "go.thethings.network/lorawan-stack/v3/pkg/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
)

// New creates a Distributor using Redis Pub/Sub.
// This Distributor routes messages based on the application identifiers.
func New(ctx context.Context, cl *ttnredis.Client) distribution.Distributor {
	return &distributor{
		ctx:   ctx,
		redis: cl,
	}
}

type distributor struct {
	ctx   context.Context
	redis *ttnredis.Client

	sets sync.Map
}

func (d *distributor) uidUplinkKey(uid string) string {
	return d.redis.Key("uid", uid, "uplinks")
}

// SendUp publishes the uplink to Pub/Sub.
func (d *distributor) SendUp(ctx context.Context, up *ttnpb.ApplicationUp) error {
	s, err := ttnredis.MarshalProto(up)
	if err != nil {
		return err
	}
	uid := unique.ID(ctx, up.ApplicationIdentifiers)
	channel := d.uidUplinkKey(uid)
	if err = d.redis.Publish(channel, s).Err(); err != nil {
		return ttnredis.ConvertError(err)
	}
	return nil
}

func (d *distributor) loadOrCreateSet(ctx context.Context, ids ttnpb.ApplicationIdentifiers) (*set, error) {
	uid := unique.ID(ctx, ids)
	s := &set{
		init: make(chan struct{}),
	}
	if existing, loaded := d.sets.LoadOrStore(uid, s); loaded {
		exists := existing.(*set)
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-exists.init:
		}
		if exists.initErr != nil {
			return nil, exists.initErr
		}
		return exists, nil
	}

	var err error
	defer func() {
		s.initErr = err
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

	channel := d.uidUplinkKey(uid)
	ps := d.redis.Subscribe(channel)
	if err = s.setup(ctx, ps); err != nil {
		return nil, err
	}

	return s, nil
}

// Subscribe adds the subscription to the appropriate subscription set.
// The subscription is automatically removed when cancelled.
func (d *distributor) Subscribe(ctx context.Context, sub *io.Subscription) error {
	s, err := d.loadOrCreateSet(ctx, *sub.ApplicationIDs())
	if err != nil {
		return err
	}
	return s.Subscribe(ctx, sub)
}
