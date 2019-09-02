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

// Package nats implements the NATS provider using the natspubsub driver.
package nats

import (
	"context"

	"github.com/nats-io/nats.go"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/natspubsub"
)

type impl struct {
}

type connection struct {
	*nats.Conn
}

// Shutdown implements provider.Shutdowner.
func (c *connection) Shutdown(_ context.Context) error {
	c.Close()
	return nil
}

// OpenConnection implements provider.Provider using the natspubsub package.
func (impl) OpenConnection(ctx context.Context, target provider.Target) (pc *provider.Connection, err error) {
	settings, ok := target.GetProvider().(*ttnpb.ApplicationPubSub_NATS)
	if !ok {
		panic("wrong provider type provided to OpenConnection")
	}
	var conn *nats.Conn
	if conn, err = nats.Connect(settings.NATS.ServerURL); err != nil {
		return nil, err
	}
	pc = &provider.Connection{
		ProviderConnection: &connection{
			Conn: conn,
		},
	}
	for _, t := range []struct {
		topic   **pubsub.Topic
		message *ttnpb.ApplicationPubSub_Message
	}{
		{
			topic:   &pc.Topics.UplinkMessage,
			message: target.GetUplinkMessage(),
		},
		{
			topic:   &pc.Topics.JoinAccept,
			message: target.GetJoinAccept(),
		},
		{
			topic:   &pc.Topics.DownlinkAck,
			message: target.GetDownlinkAck(),
		},
		{
			topic:   &pc.Topics.DownlinkNack,
			message: target.GetDownlinkNack(),
		},
		{
			topic:   &pc.Topics.DownlinkSent,
			message: target.GetDownlinkSent(),
		},
		{
			topic:   &pc.Topics.DownlinkFailed,
			message: target.GetDownlinkFailed(),
		},
		{
			topic:   &pc.Topics.DownlinkQueued,
			message: target.GetDownlinkQueued(),
		},
		{
			topic:   &pc.Topics.LocationSolved,
			message: target.GetLocationSolved(),
		},
	} {
		if t.message == nil {
			continue
		}
		if *t.topic, err = natspubsub.OpenTopic(
			conn,
			combineSubjects(target.GetBaseTopic(), t.message.GetTopic()),
			&natspubsub.TopicOptions{},
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	for _, s := range []struct {
		subscription **pubsub.Subscription
		message      *ttnpb.ApplicationPubSub_Message
	}{
		{
			subscription: &pc.Subscriptions.Push,
			message:      target.GetDownlinkPush(),
		},
		{
			subscription: &pc.Subscriptions.Replace,
			message:      target.GetDownlinkReplace(),
		},
	} {
		if s.message == nil {
			continue
		}
		if *s.subscription, err = natspubsub.OpenSubscription(
			conn,
			combineSubjects(target.GetBaseTopic(), s.message.GetTopic()),
			&natspubsub.SubscriptionOptions{},
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return pc, nil
}

func init() {
	provider.RegisterProvider(&ttnpb.ApplicationPubSub_NATS{}, impl{})
}
