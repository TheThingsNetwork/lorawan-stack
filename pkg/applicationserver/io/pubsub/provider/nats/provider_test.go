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

	nats_server "github.com/nats-io/nats-server/v2/server"
	nats_test_server "github.com/nats-io/nats-server/v2/test"
	nats_client "github.com/nats-io/nats.go"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub"
)

var timeout = (1 << 8) * test.Delay

type allEnabled struct{}

// Enabled implements provider.Enabler.
func (e *allEnabled) Enabled(context.Context, ttnpb.ApplicationPubSub_Provider) error {
	return nil
}

func TestOpenConnection(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	natsServer := nats_test_server.RunServer(&nats_server.Options{
		Host:           "127.0.0.1",
		Port:           4123,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	})
	a.So(natsServer, should.NotBeNil)
	defer natsServer.Shutdown()

	natsClient, err := nats_client.Connect("nats://localhost:4123")
	a.So(err, should.BeNil)
	defer natsClient.Close()

	pb := &ttnpb.ApplicationPubSub{
		Ids: &ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: "app1",
			},
			PubSubId: "ps1",
		},
		Provider: &ttnpb.ApplicationPubSub_Nats{
			Nats: &ttnpb.ApplicationPubSub_NATSProvider{},
		},
		BaseTopic: "app1.ps1",
		DownlinkPush: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.push",
		},
		DownlinkReplace: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.replace",
		},
		UplinkMessage: &ttnpb.ApplicationPubSub_Message{
			Topic: "uplink.message",
		},
		UplinkNormalized: &ttnpb.ApplicationPubSub_Message{
			Topic: "uplink.normalized",
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
		DownlinkQueueInvalidated: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink.invalidated",
		},
		LocationSolved: &ttnpb.ApplicationPubSub_Message{
			Topic: "location.solved",
		},
		ServiceData: &ttnpb.ApplicationPubSub_Message{
			Topic: "service.data",
		},
	}

	impl, err := provider.GetProvider(&ttnpb.ApplicationPubSub{
		Provider: &ttnpb.ApplicationPubSub_Nats{
			Nats: &ttnpb.ApplicationPubSub_NATSProvider{
				ServerUrl: "nats://invalid.local:4222",
			},
		},
	})
	a.So(impl, should.NotBeNil)
	a.So(err, should.BeNil)

	// Invalid attributes - invalid server.
	{
		conn, err := impl.OpenConnection(ctx, pb, &allEnabled{})
		a.So(conn, should.BeNil)
		a.So(err, should.NotBeNil)
	}

	pb.Provider = &ttnpb.ApplicationPubSub_Nats{
		Nats: &ttnpb.ApplicationPubSub_NATSProvider{
			ServerUrl: "nats://localhost:4123",
		},
	}

	// Valid attributes - connection established.
	{
		conn, err := impl.OpenConnection(ctx, pb, &allEnabled{})
		a.So(conn, should.NotBeNil)
		a.So(err, should.BeNil)
		defer conn.Shutdown(ctx)

		// Wait for subscriptions to connect, since they are not synchronous.
		time.Sleep(timeout)

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
					err = natsClient.Publish(tc.subject, []byte("foobar"))
					a.So(err, should.BeNil)

					ctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					msg, err := tc.subscription.Receive(ctx)
					if tc.expectMessage {
						a.So(err, should.BeNil)
						if a.So(msg, should.NotBeNil) {
							a.So(msg.Body, should.Resemble, []byte("foobar"))
						}
					} else if err == nil {
						t.Fatal("Unexpected message received")
					}
					if msg != nil {
						msg.Ack()
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
					name:    "ValidNormalizedUplink",
					subject: "app1.ps1.uplink.normalized",
					topic:   conn.Topics.UplinkNormalized,
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
					name:    "ValidDownlinkQueueInvalidated",
					subject: "app1.ps1.downlink.invalidated",
					topic:   conn.Topics.DownlinkQueueInvalidated,
				},
				{
					name:    "ValidLocationSolved",
					subject: "app1.ps1.location.solved",
					topic:   conn.Topics.LocationSolved,
				},
				{
					name:    "ValidServiceData",
					subject: "app1.ps1.service.data",
					topic:   conn.Topics.ServiceData,
				},
			} {
				t.Run(tc.name, func(t *testing.T) {
					upCh := make(chan *nats_client.Msg, 10)
					defer close(upCh)

					sub, err := natsClient.ChanSubscribe(tc.subject, upCh)
					a.So(sub, should.NotBeNil)
					a.So(err, should.BeNil)
					defer sub.Unsubscribe()

					// We have to sleep here since ChanSubscribe is not actually synced,
					// so it could be the case that we to topic.Send before the subscription,
					// was actually opened.
					time.Sleep(timeout)

					ctx, cancel := context.WithTimeout(ctx, timeout)
					defer cancel()

					err = tc.topic.Send(ctx, &pubsub.Message{
						Body: []byte("foobar"),
					})
					a.So(err, should.BeNil)

					select {
					case <-time.After(timeout):
						t.Fatal("Expected message never arrived")
					case msg := <-upCh:
						if a.So(msg, should.NotBeNil) {
							a.So(msg.Data, should.Resemble, []byte("foobar"))
						}
					}
				})
			}
		})
	}
}
