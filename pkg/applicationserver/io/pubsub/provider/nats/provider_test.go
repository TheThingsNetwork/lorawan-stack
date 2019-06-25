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

package nats_test

import (
	"context"
	"testing"
	"time"

	nats_server "github.com/nats-io/nats-server/test"
	nats_client "github.com/nats-io/nats.go"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider/nats"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub"
)

var timeout = (1 << 8) * test.Delay

func TestOpenConnection(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	natsServer := nats_server.RunDefaultServer()
	a.So(natsServer, should.NotBeNil)
	defer natsServer.Shutdown()
	time.Sleep(timeout)

	pb := &ttnpb.ApplicationPubSub{
		ApplicationPubSubIdentifiers: ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "app1",
			},
			PubSubID: "ps1",
		},
		Attributes: map[string]string{},
		BaseTopic:  "app1.ps1",
		DownlinkPush: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.push",
		},
		DownlinkReplace: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.replace",
		},
		UplinkMessage: &ttnpb.ApplicationPubSub_Message{
			Topic: "uplink.message",
		},
		JoinAccept: &ttnpb.ApplicationPubSub_Message{
			Topic: "join.accept",
		},
		DownlinkAck: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.ack",
		},
		DownlinkNack: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.nack",
		},
		DownlinkSent: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.sent",
		},
		DownlinkFailed: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlnk.failed",
		},
		DownlinkQueued: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.queued",
		},
		LocationSolved: &ttnpb.ApplicationPubSub_Message{
			Topic: "location.solved",
		},
	}

	impl, err := provider.GetProvider(ttnpb.ApplicationPubSub_NATS)
	a.So(impl, should.NotBeNil)
	a.So(err, should.BeNil)

	// Invalid attributes - no server provided.
	{
		conn, err := impl.OpenConnection(ctx, pb)
		a.So(conn, should.BeNil)
		a.So(err, should.NotBeNil)
	}

	pb.Attributes[nats.NATSServerAttribute] = "localhost"

	// Valid attributes - connection established.
	{
		conn, err := impl.OpenConnection(ctx, pb)
		a.So(conn, should.NotBeNil)
		a.So(err, should.BeNil)

		defer conn.Shutdown(ctx)

		t.Run("Downstream", func(t *testing.T) {
			for _, tc := range []struct {
				name          string
				subject       string
				subscription  *pubsub.Subscription
				expectMessage bool
			}{
				{
					name:          "ValidPush",
					subject:       "app1.ps1.downlink.push",
					subscription:  conn.Subscriptions.Push,
					expectMessage: true,
				},
				{
					name:          "ValidReplace",
					subject:       "app1.ps1.downlink.replace",
					subscription:  conn.Subscriptions.Replace,
					expectMessage: true,
				},
				{
					name:          "InvalidPush",
					subject:       "foo.bar",
					subscription:  conn.Subscriptions.Push,
					expectMessage: false,
				},
				{
					name:          "InvalidReplace",
					subject:       "bar.foo",
					subscription:  conn.Subscriptions.Replace,
					expectMessage: false,
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					natsClient, err := nats_client.Connect("localhost")
					a.So(err, should.BeNil)
					defer natsClient.Close()

					err = natsClient.Publish(tc.subject, []byte("foobar"))
					a.So(err, should.BeNil)

					msg, err := tc.subscription.Receive(ctx)
					if tc.expectMessage {
						a.So(err, should.BeNil)
						a.So(msg, should.NotBeNil)

						a.So(msg.Body, should.Resemble, []byte("foobar"))
					} else if err == nil {
						t.Fatal("Unexpected message received")
					}
				})
			}
		})
		t.Run("Upstream", func(t *testing.T) {
			for _, tc := range []struct {
				name    string
				subject string
				topic   *pubsub.Topic
			}{
				{
					name:    "ValidUplink",
					subject: "app1.ps1.uplink.message",
					topic:   conn.Topics.UplinkMessage,
				},
				{
					name:    "ValidJoinAccept",
					subject: "app1.ps1.join.accept",
					topic:   conn.Topics.JoinAccept,
				},
				{
					name:    "ValidDownlinkAck",
					subject: "app1.ps1.downlink.ack",
					topic:   conn.Topics.DownlinkAck,
				},
				{
					name:    "ValidDownlinkNack",
					subject: "app1.ps1.downlink.nack",
					topic:   conn.Topics.DownlinkNack,
				},
				{
					name:    "ValidDownlinkSent",
					subject: "app1.ps1.downlink.sent",
					topic:   conn.Topics.DownlinkSent,
				},
				{
					name:    "ValidDownlinkQueued",
					subject: "app1.ps1.downlink.queued",
					topic:   conn.Topics.DownlinkQueued,
				},
				{
					name:    "ValidLocationSolved",
					subject: "app1.ps1.location.solved",
					topic:   conn.Topics.LocationSolved,
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					ctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					natsClient, err := nats_client.Connect("localhost")
					a.So(err, should.BeNil)
					defer natsClient.Close()

					upCh := make(chan *nats_client.Msg)
					defer close(upCh)

					sub, err := natsClient.ChanSubscribe(tc.subject, upCh)
					a.So(sub, should.NotBeNil)
					a.So(err, should.BeNil)

					err = tc.topic.Send(ctx, &pubsub.Message{
						Body: []byte("foobar"),
					})
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected message never arrived")
					case msg := <-upCh:
						a.So(msg, should.NotBeNil)
						a.So(msg.Data, should.Resemble, []byte("foobar"))
					}
				})
			}
		})
	}
}
