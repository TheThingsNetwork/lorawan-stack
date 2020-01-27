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

package mqtt

import (
	"context"
	"fmt"
	"io/ioutil"
	"testing"
	"time"

	paho_mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub"
)

func TestOpenConnection(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	ca, err := ioutil.ReadFile("testdata/rootCA.pem")
	a.So(err, should.BeNil)
	clientCert, err := ioutil.ReadFile("testdata/clientcert.pem")
	a.So(err, should.BeNil)
	clientKey, err := ioutil.ReadFile("testdata/clientkey.pem")
	a.So(err, should.BeNil)
	serverCert, err := ioutil.ReadFile("testdata/servercert.pem")
	a.So(err, should.BeNil)
	serverKey, err := ioutil.ReadFile("testdata/serverkey.pem")
	a.So(err, should.BeNil)

	clientTLSConfig, err := createTLSConfig(ca, clientCert, clientKey)
	a.So(err, should.BeNil)
	serverTLSConfig, err := createTLSConfig(ca, serverCert, serverKey)
	a.So(err, should.BeNil)

	lis, tlsLis, err := startMQTTServer(ctx, serverTLSConfig)
	a.So(err, should.BeNil)
	a.So(lis, should.NotBeNil)
	a.So(tlsLis, should.NotBeNil)
	defer lis.Close()
	defer tlsLis.Close()

	pb := &ttnpb.ApplicationPubSub{
		ApplicationPubSubIdentifiers: ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "app1",
			},
			PubSubID: "ps1",
		},
		Provider: &ttnpb.ApplicationPubSub_MQTT{
			MQTT: &ttnpb.ApplicationPubSub_MQTTProvider{},
		},
		BaseTopic: "app1/ps1",
		DownlinkPush: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/push",
		},
		DownlinkReplace: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/replace",
		},
		UplinkMessage: &ttnpb.ApplicationPubSub_Message{
			Topic: "uplink/message",
		},
		JoinAccept: &ttnpb.ApplicationPubSub_Message{
			Topic: "join/accept",
		},
		DownlinkAck: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/ack",
		},
		DownlinkNack: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/nack",
		},
		DownlinkSent: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/sent",
		},
		DownlinkFailed: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlnk/failed",
		},
		DownlinkQueued: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/queued",
		},
		LocationSolved: &ttnpb.ApplicationPubSub_Message{
			Topic: "location/solved",
		},
	}

	impl, err := provider.GetProvider(&ttnpb.ApplicationPubSub{
		Provider: &ttnpb.ApplicationPubSub_MQTT{},
	})
	a.So(impl, should.NotBeNil)
	a.So(err, should.BeNil)

	// Invalid attributes - no server provided.
	{
		conn, err := impl.OpenConnection(ctx, pb)
		a.So(conn, should.BeNil)
		a.So(err, should.NotBeNil)
	}

	// Valid attributes - connection established.
	for _, utc := range []struct {
		name         string
		provider     ttnpb.ApplicationPubSub_Provider
		createClient func(*testing.T, *assertions.Assertion) paho_mqtt.Client
	}{
		{
			name: "TCP",
			provider: &ttnpb.ApplicationPubSub_MQTT{
				MQTT: &ttnpb.ApplicationPubSub_MQTTProvider{
					ServerURL:    fmt.Sprintf("tcp://%v", lis.Addr()),
					SubscribeQoS: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
					PublishQoS:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
				},
			},
			createClient: func(t *testing.T, a *assertions.Assertion) paho_mqtt.Client {
				clientOpts := paho_mqtt.NewClientOptions()
				clientOpts.AddBroker(lis.Addr().String())
				client := paho_mqtt.NewClient(clientOpts)
				a.So(client, should.NotBeNil)
				token := client.Connect()
				if !token.WaitTimeout(timeout) {
					t.Fatal("Connection timeout")
				}
				if !a.So(token.Error(), should.BeNil) {
					t.FailNow()
				}
				return client
			},
		},
		{
			name: "TCP+TLS",
			provider: &ttnpb.ApplicationPubSub_MQTT{
				MQTT: &ttnpb.ApplicationPubSub_MQTTProvider{
					ServerURL:    fmt.Sprintf("tcps://%v", tlsLis.Addr()),
					SubscribeQoS: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
					PublishQoS:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,

					UseTLS:        true,
					TLSCA:         ca,
					TLSClientCert: clientCert,
					TLSClientKey:  clientKey,
				},
			},
			createClient: func(t *testing.T, a *assertions.Assertion) paho_mqtt.Client {
				clientOpts := paho_mqtt.NewClientOptions()
				clientOpts.AddBroker(fmt.Sprintf("tcps://%v", tlsLis.Addr()))
				clientOpts.SetTLSConfig(clientTLSConfig)
				client := paho_mqtt.NewClient(clientOpts)
				a.So(client, should.NotBeNil)
				token := client.Connect()
				if !token.WaitTimeout(timeout) {
					t.Fatal("Connection timeout")
				}
				if !a.So(token.Error(), should.BeNil) {
					t.FailNow()
				}
				return client
			},
		},
	} {
		t.Run(utc.name, func(t *testing.T) {
			pb.Provider = utc.provider

			conn, err := impl.OpenConnection(ctx, pb)
			a.So(conn, should.NotBeNil)
			if !a.So(err, should.BeNil) {
				t.FailNow()
			}
			defer conn.Shutdown(ctx)

			t.Run("Downstream", func(t *testing.T) {
				for _, tc := range []struct {
					name          string
					topicName     string
					subscription  *pubsub.Subscription
					expectMessage bool
				}{
					{
						name:          "ValidPush",
						topicName:     "app1/ps1/downlink/push",
						subscription:  conn.Subscriptions.Push,
						expectMessage: true,
					},
					{
						name:          "ValidReplace",
						topicName:     "app1/ps1/downlink/replace",
						subscription:  conn.Subscriptions.Replace,
						expectMessage: true,
					},
					{
						name:          "InvalidPush",
						topicName:     "foo/bar",
						subscription:  conn.Subscriptions.Push,
						expectMessage: false,
					},
					{
						name:          "InvalidReplace",
						topicName:     "bar/foo",
						subscription:  conn.Subscriptions.Replace,
						expectMessage: false,
					},
				} {
					t.Run(tc.name, func(t *testing.T) {
						a := assertions.New(t)
						ctx, cancel := context.WithTimeout(ctx, timeout)
						defer cancel()

						client := utc.createClient(t, a)
						defer client.Disconnect(uint(timeout / time.Millisecond))

						token := client.Publish(tc.topicName, 2, false, "foobar")
						if !token.WaitTimeout(timeout) {
							t.Fatal("Publish timeout")
						}
						if !a.So(token.Error(), should.BeNil) {
							t.FailNow()
						}

						msg, err := tc.subscription.Receive(ctx)
						if tc.expectMessage {
							a.So(err, should.BeNil)
							a.So(msg, should.NotBeNil)

							a.So(msg.Body, should.Resemble, []byte("foobar"))
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
					name      string
					topicName string
					topic     *pubsub.Topic
				}{
					{
						name:      "ValidUplink",
						topicName: "app1/ps1/uplink/message",
						topic:     conn.Topics.UplinkMessage,
					},
					{
						name:      "ValidJoinAccept",
						topicName: "app1/ps1/join/accept",
						topic:     conn.Topics.JoinAccept,
					},
					{
						name:      "ValidDownlinkAck",
						topicName: "app1/ps1/downlink/ack",
						topic:     conn.Topics.DownlinkAck,
					},
					{
						name:      "ValidDownlinkNack",
						topicName: "app1/ps1/downlink/nack",
						topic:     conn.Topics.DownlinkNack,
					},
					{
						name:      "ValidDownlinkSent",
						topicName: "app1/ps1/downlink/sent",
						topic:     conn.Topics.DownlinkSent,
					},
					{
						name:      "ValidDownlinkQueued",
						topicName: "app1/ps1/downlink/queued",
						topic:     conn.Topics.DownlinkQueued,
					},
					{
						name:      "ValidLocationSolved",
						topicName: "app1/ps1/location/solved",
						topic:     conn.Topics.LocationSolved,
					},
				} {
					t.Run(tc.name, func(t *testing.T) {
						a := assertions.New(t)

						client := utc.createClient(t, a)
						defer client.Disconnect(uint(timeout / time.Millisecond))

						upCh := make(chan paho_mqtt.Message, 10)
						defer close(upCh)
						token := client.Subscribe(tc.topicName, 2, func(_ paho_mqtt.Client, msg paho_mqtt.Message) {
							upCh <- msg
						})
						if !token.WaitTimeout(timeout) {
							t.Fatal("Subscribe timeout")
						}
						if !a.So(token.Error(), should.BeNil) {
							t.FailNow()
						}

						err = tc.topic.Send(ctx, &pubsub.Message{
							Body: []byte("foobar"),
						})
						a.So(err, should.BeNil)

						select {
						case <-time.After(timeout):
							t.Fatal("Expected message never arrived")
						case msg := <-upCh:
							a.So(msg, should.NotBeNil)
							a.So(msg.Payload(), should.Resemble, []byte("foobar"))
						}
					})
				}
			})
		})
	}
}

func init() {
	timeout = (1 << 8) * test.Delay
}
