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
	"os"
	"strconv"
	"testing"
	"time"

	paho_mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub"
)

type allEnabled struct{}

// Enabled implements provider.Enabler.
func (e *allEnabled) Enabled(context.Context, ttnpb.ApplicationPubSub_Provider) error {
	return nil
}

func TestOpenConnection(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	ca, err := os.ReadFile("testdata/rootCA.pem")
	a.So(err, should.BeNil)
	clientCert, err := os.ReadFile("testdata/clientcert.pem")
	a.So(err, should.BeNil)
	clientKey, err := os.ReadFile("testdata/clientkey.pem")
	a.So(err, should.BeNil)
	serverCert, err := os.ReadFile("testdata/servercert.pem")
	a.So(err, should.BeNil)
	serverKey, err := os.ReadFile("testdata/serverkey.pem")
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
		Ids: &ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIds: &ttnpb.ApplicationIdentifiers{
				ApplicationId: "app1",
			},
			PubSubId: "ps1",
		},
		Provider: &ttnpb.ApplicationPubSub_Mqtt{
			Mqtt: &ttnpb.ApplicationPubSub_MQTTProvider{},
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
		DownlinkQueueInvalidated: &ttnpb.ApplicationPubSub_Message{
			Topic: "downlink/invalidated",
		},
		LocationSolved: &ttnpb.ApplicationPubSub_Message{
			Topic: "location/solved",
		},
		ServiceData: &ttnpb.ApplicationPubSub_Message{
			Topic: "service/data",
		},
	}

	impl, err := provider.GetProvider(&ttnpb.ApplicationPubSub{
		Provider: &ttnpb.ApplicationPubSub_Mqtt{},
	})
	a.So(impl, should.NotBeNil)
	a.So(err, should.BeNil)

	// Invalid attributes - no server provided.
	{
		conn, err := impl.OpenConnection(ctx, pb, &allEnabled{})
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
			provider: &ttnpb.ApplicationPubSub_Mqtt{
				Mqtt: &ttnpb.ApplicationPubSub_MQTTProvider{
					ServerUrl:    fmt.Sprintf("tcp://%v", lis.Addr()),
					SubscribeQos: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
					PublishQos:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
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
			provider: &ttnpb.ApplicationPubSub_Mqtt{
				Mqtt: &ttnpb.ApplicationPubSub_MQTTProvider{
					ServerUrl:    fmt.Sprintf("tcps://%v", tlsLis.Addr()),
					SubscribeQos: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
					PublishQos:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,

					UseTls:        true,
					TlsCa:         ca,
					TlsClientCert: clientCert,
					TlsClientKey:  clientKey,
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
			client := utc.createClient(t, a)
			defer client.Disconnect(uint(timeout / time.Millisecond))

			unsubscribe := func(t *testing.T, topic string) {
				token := client.Unsubscribe(topic)
				if !token.WaitTimeout(timeout) {
					t.Fatal("Unsubscribe timeout")
				}
				if !a.So(token.Error(), should.BeNil) {
					t.FailNow()
				}
			}

			pb.Provider = utc.provider

			conn, err := impl.OpenConnection(ctx, pb, &allEnabled{})
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

						token := client.Publish(tc.topicName, 2, false, "foobar")
						if !token.WaitTimeout(timeout) {
							t.Fatal("Publish timeout")
						}
						if !a.So(token.Error(), should.BeNil) {
							t.FailNow()
						}

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
						name:      "ValidDownlinkQueueInvalidated",
						topicName: "app1/ps1/downlink/invalidated",
						topic:     conn.Topics.DownlinkQueueInvalidated,
					},
					{
						name:      "ValidLocationSolved",
						topicName: "app1/ps1/location/solved",
						topic:     conn.Topics.LocationSolved,
					},
					{
						name:      "ValidServiceData",
						topicName: "app1/ps1/service/data",
						topic:     conn.Topics.ServiceData,
					},
				} {
					t.Run(tc.name, func(t *testing.T) {
						a := assertions.New(t)

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
						defer unsubscribe(t, tc.topicName)

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
								a.So(msg.Payload(), should.Resemble, []byte("foobar"))
							}
						}
					})
				}
			})
		})
	}
}

func TestAdaptURLScheme(t *testing.T) {
	for i, tc := range []struct {
		ProvidedURL string
		ExpectedURL string
		ShouldError bool
	}{
		{
			ProvidedURL: "mqtt://localhost:1885",
			ExpectedURL: "tcp://localhost:1885",
			ShouldError: false,
		},
		{
			ProvidedURL: "mqtts://localhost:8885",
			ExpectedURL: "ssl://localhost:8885",
			ShouldError: false,
		},
		{
			ProvidedURL: "bar!://foo",
			ShouldError: true,
		},
	} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			a := assertions.New(t)
			result, err := adaptURLScheme(tc.ProvidedURL)
			if err != nil {
				if !tc.ShouldError {
					t.Fatal("Unexpected error", err)
				}
			} else {
				if tc.ShouldError {
					t.Fatal("Expected error but got nil")
				}
			}
			a.So(result, should.Equal, tc.ExpectedURL)
		})
	}
}

func init() {
	timeout = (1 << 8) * test.Delay
}
