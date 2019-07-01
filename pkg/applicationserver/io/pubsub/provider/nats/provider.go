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

	"github.com/nats-io/go-nats"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"gocloud.dev/pubsub"
	"gocloud.dev/pubsub/natspubsub"
)

const (
	// NATSServerAttribute is the server URL for the NATS server.
	NATSServerAttribute = "nats-server"
)

var (
	errAttributeNotFound = errors.DefineNotFound("attribute_not_found", "attribute `{attribute}` not found")
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
func (impl) OpenConnection(ctx context.Context, pb *ttnpb.ApplicationPubSub) (pc *provider.Connection, err error) {
	var serverURL string
	var ok bool
	if serverURL, ok = pb.Attributes[NATSServerAttribute]; !ok {
		return nil, errAttributeNotFound.WithAttributes("attribute", NATSServerAttribute)
	}
	var conn *nats.Conn
	if conn, err = nats.Connect(serverURL); err != nil {
		return nil, err
	}
	pc = &provider.Connection{
		ProviderConnection: &connection{
			Conn: conn,
		},
	}
	for _, t := range []struct {
		topic   **pubsub.Topic
		subject string
	}{
		{
			topic:   &pc.Topics.UplinkMessage,
			subject: pb.GetUplinkMessage().GetTopic(),
		},
		{
			topic:   &pc.Topics.JoinAccept,
			subject: pb.GetJoinAccept().GetTopic(),
		},
		{
			topic:   &pc.Topics.DownlinkAck,
			subject: pb.GetDownlinkAck().GetTopic(),
		},
		{
			topic:   &pc.Topics.DownlinkNack,
			subject: pb.GetDownlinkNack().GetTopic(),
		},
		{
			topic:   &pc.Topics.DownlinkSent,
			subject: pb.GetDownlinkSent().GetTopic(),
		},
		{
			topic:   &pc.Topics.DownlinkFailed,
			subject: pb.GetDownlinkFailed().GetTopic(),
		},
		{
			topic:   &pc.Topics.DownlinkQueued,
			subject: pb.GetDownlinkQueued().GetTopic(),
		},
		{
			topic:   &pc.Topics.LocationSolved,
			subject: pb.GetLocationSolved().GetTopic(),
		},
	} {
		if *t.topic, err = natspubsub.OpenTopic(
			conn,
			combineSubjects(pb.BaseTopic, t.subject),
			&natspubsub.TopicOptions{},
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	for _, s := range []struct {
		subscription **pubsub.Subscription
		subject      string
	}{
		{
			subscription: &pc.Subscriptions.Push,
			subject:      pb.GetDownlinkPush().GetTopic(),
		},
		{
			subscription: &pc.Subscriptions.Replace,
			subject:      pb.GetDownlinkReplace().GetTopic(),
		},
	} {
		if *s.subscription, err = natspubsub.OpenSubscription(
			conn,
			combineSubjects(pb.BaseTopic, s.subject),
			&natspubsub.SubscriptionOptions{},
		); err != nil {
			conn.Close()
			return nil, err
		}
	}
	return pc, nil
}

func init() {
	provider.RegisterProvider(ttnpb.ApplicationPubSub_NATS, impl{})
}
