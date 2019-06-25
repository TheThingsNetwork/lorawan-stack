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

package provider

import (
	"context"

	"go.thethings.network/lorawan-stack/pkg/errors"
	"gocloud.dev/pubsub"
)

// DownlinkSubscriptions contains the subscriptions for the push and replace queue operations.
type DownlinkSubscriptions struct {
	Push    *pubsub.Subscription
	Replace *pubsub.Subscription
}

// Shutdown shutdowns the active subscriptions.
func (ds *DownlinkSubscriptions) Shutdown(ctx context.Context) error {
	for _, sub := range []*pubsub.Subscription{
		ds.Push,
		ds.Replace,
	} {
		if sub != nil {
			if err := sub.Shutdown(ctx); err != nil && !errors.IsCanceled(err) {
				return err
			}
		}
	}
	return nil
}

// UplinkTopics contains the topics for the uplink messages.
type UplinkTopics struct {
	UplinkMessage  *pubsub.Topic
	JoinAccept     *pubsub.Topic
	DownlinkAck    *pubsub.Topic
	DownlinkNack   *pubsub.Topic
	DownlinkSent   *pubsub.Topic
	DownlinkFailed *pubsub.Topic
	DownlinkQueued *pubsub.Topic
	LocationSolved *pubsub.Topic
}

// Shutdown shutdowns the active topics.
func (ut *UplinkTopics) Shutdown(ctx context.Context) error {
	for _, topic := range []*pubsub.Topic{
		ut.UplinkMessage,
		ut.JoinAccept,
		ut.DownlinkAck,
		ut.DownlinkNack,
		ut.DownlinkSent,
		ut.DownlinkFailed,
		ut.DownlinkQueued,
		ut.LocationSolved,
	} {
		if topic != nil {
			if err := topic.Shutdown(ctx); err != nil && !errors.IsCanceled(err) {
				return err
			}
		}
	}
	return nil
}

// Shutdowner is an interface that contains a contextual shutdown method.
type Shutdowner interface {
	Shutdown(ctx context.Context) error
}

// ProviderConnection is an interface that represents a provider specific connection.
type ProviderConnection interface {
	Shutdowner
}

// Connection is a wrapper that wraps the topics and subscriptions with a ProviderConnection.
type Connection struct {
	Topics             UplinkTopics
	Subscriptions      DownlinkSubscriptions
	ProviderConnection ProviderConnection
}

// Shutdown shuts down the topics, subscriptions and the connections if required.
func (c *Connection) Shutdown(ctx context.Context) error {
	for _, s := range []Shutdowner{
		&c.Topics,
		&c.Subscriptions,
		c.ProviderConnection,
	} {
		if err := s.Shutdown(ctx); err != nil {
			return err
		}
	}
	return nil
}
