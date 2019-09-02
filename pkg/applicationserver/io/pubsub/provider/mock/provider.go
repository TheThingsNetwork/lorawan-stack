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

// Package mock implements a mock PubSub provider using the mempubsub driver.
package mock

import (
	"context"
	"time"

	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/mempubsub"
)

// ConnectionWithError is an connection wrapped with an error.
type ConnectionWithError struct {
	*Connection
	error
}

// Impl is the mock provider implementation.
type Impl struct {
	OpenConnectionCh chan *ConnectionWithError
	ShutdownCh       chan *ConnectionWithError
}

// Connection is a set of mempubsub topics.
type Connection struct {
	impl *Impl
	provider.Target

	Push    *pubsub.Topic
	Replace *pubsub.Topic

	UplinkMessage  *pubsub.Subscription
	JoinAccept     *pubsub.Subscription
	DownlinkAck    *pubsub.Subscription
	DownlinkNack   *pubsub.Subscription
	DownlinkSent   *pubsub.Subscription
	DownlinkFailed *pubsub.Subscription
	DownlinkQueued *pubsub.Subscription
	LocationSolved *pubsub.Subscription
}

func (c *Connection) ApplicationPubSubIdentifiers() *ttnpb.ApplicationPubSubIdentifiers {
	if pb, ok := c.Target.(*ttnpb.ApplicationPubSub); ok {
		return &pb.ApplicationPubSubIdentifiers
	}
	return nil
}

// Shutdown implements provider.Shutdowner.
func (c *Connection) Shutdown(ctx context.Context) (err error) {
	defer func() {
		c.impl.ShutdownCh <- &ConnectionWithError{
			Connection: c,
			error:      err,
		}
	}()
	for _, topic := range []interface{ Shutdown(context.Context) error }{
		c.Push,
		c.Replace,

		c.UplinkMessage,
		c.JoinAccept,
		c.DownlinkAck,
		c.DownlinkNack,
		c.DownlinkSent,
		c.DownlinkFailed,
		c.DownlinkQueued,
		c.LocationSolved,
	} {
		if topic != nil {
			if err = topic.Shutdown(ctx); err != nil && !errors.IsCanceled(err) {
				return err
			}
		}
	}
	return nil
}

// OpenConnection implements provider.Provider using the mempubsub package.
func (i *Impl) OpenConnection(ctx context.Context, target provider.Target) (pc *provider.Connection, err error) {
	conn := &Connection{
		impl:   i,
		Target: target,
	}
	pc = &provider.Connection{
		ProviderConnection: conn,
	}
	defer func() {
		i.OpenConnectionCh <- &ConnectionWithError{
			Connection: conn,
			error:      err,
		}
	}()
	for _, t := range []struct {
		topic        **pubsub.Topic
		subscription **pubsub.Subscription
		subject      string
	}{
		{
			topic:        &conn.Push,
			subscription: &pc.Subscriptions.Push,
		},
		{
			topic:        &conn.Replace,
			subscription: &pc.Subscriptions.Replace,
		},
		{
			topic:        &pc.Topics.UplinkMessage,
			subscription: &conn.UplinkMessage,
		},
		{
			topic:        &pc.Topics.JoinAccept,
			subscription: &conn.JoinAccept,
		},
		{
			topic:        &pc.Topics.DownlinkAck,
			subscription: &conn.DownlinkAck,
		},
		{
			topic:        &pc.Topics.DownlinkNack,
			subscription: &conn.DownlinkNack,
		},
		{
			topic:        &pc.Topics.DownlinkSent,
			subscription: &conn.DownlinkSent,
		},
		{
			topic:        &pc.Topics.DownlinkFailed,
			subscription: &conn.DownlinkFailed,
		},
		{
			topic:        &pc.Topics.DownlinkQueued,
			subscription: &conn.DownlinkQueued,
		},
		{
			topic:        &pc.Topics.LocationSolved,
			subscription: &conn.LocationSolved,
		},
	} {
		*t.topic = mempubsub.NewTopic()
		*t.subscription = mempubsub.NewSubscription(*t.topic, 5*time.Minute)
	}
	return pc, nil
}

func init() {
	impl := &Impl{
		OpenConnectionCh: make(chan *ConnectionWithError, 10),
		ShutdownCh:       make(chan *ConnectionWithError, 10),
	}
	for _, p := range []ttnpb.ApplicationPubSub_Provider{
		&ttnpb.ApplicationPubSub_NATS{},
		&ttnpb.ApplicationPubSub_MQTT{},
	} {
		provider.RegisterProvider(p, impl)
	}
}
