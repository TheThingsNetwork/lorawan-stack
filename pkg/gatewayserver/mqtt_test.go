// Copyright © 2018 The Things Network Foundation, The Things Industries B.V.
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

package gatewayserver_test

import (
	"context"
	"fmt"
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/pkg/unique"

	"github.com/TheThingsIndustries/mystique/pkg/topic"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const mqttConnectionTimeout = 3 * time.Second

// TODO: Refactor TestMQTTConnection/TestUDP(/TestLink?)
func TestMQTTConnection(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)
	ctx := log.NewContext(test.Context(), logger)
	ctx = clusterauth.NewContext(ctx, nil)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	gtwID, err := unique.ToGatewayID(registeredGatewayUID)
	a.So(err, should.BeNil)
	gtwID.EUI = &registeredGatewayEUI

	is, isAddr := StartMockIsGatewayServer(ctx, []ttnpb.Gateway{
		{
			GatewayIdentifiers: gtwID,
			FrequencyPlanID:    "EU_863_870",
			DisableTxDelay:     true,
		},
	})
	is.rights = []ttnpb.Right{ttnpb.RIGHT_GATEWAY_INFO, ttnpb.RIGHT_GATEWAY_LINK}
	logger.WithField("address", isAddr).Info("Started mock Identity Server")

	ns, nsAddr := StartMockGsNsServer(ctx)
	logger.WithField("address", nsAddr).Info("Started mock Network Server")

	mqttAddress := "127.0.0.1:9883"
	c := component.MustNew(logger, &component.Config{
		ServiceBase: config.ServiceBase{
			Cluster: config.Cluster{
				Name:           "test-gateway-server",
				IdentityServer: isAddr,
				NetworkServer:  nsAddr,
			},
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
	c.FrequencyPlans.Fetcher = test.FrequencyPlansFetcher
	gs, err := gatewayserver.New(c, gatewayserver.Config{
		MQTT: gatewayserver.MQTTConfig{
			Listen: mqttAddress,
		},
	})
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway Server could not be initialized:", err)
	}

	err = gs.Start()
	if !a.So(err, should.BeNil) {
		t.Fatal("Gateway Server could not start:", err)
	}

	gsStart := time.Now()
	for gs.GetPeer(ttnpb.PeerInfo_IDENTITY_SERVER, []string{}, nil) == nil || gs.GetPeer(ttnpb.PeerInfo_NETWORK_SERVER, []string{}, nil) == nil {
		if time.Since(gsStart) > nsReceptionTimeout {
			t.Fatal("Identity Server and Network Server were not initialized in time by the Gateway Server - timeout")
		}
		time.Sleep(2 * time.Millisecond)
	}

	clientOptions := mqtt.NewClientOptions()
	clientOptions.AddBroker(fmt.Sprintf("tcp://%s", mqttAddress))
	clientOptions.SetUsername(registeredGatewayUID)
	clientOptions.SetPassword("test")
	clientOptions.SetClientID("test-client")

	client := mqtt.NewClient(clientOptions)

	ok := t.Run("Connect", func(t *testing.T) {
		a := assertions.New(t)

		token := client.Connect()
		if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
			t.Fatal("CONNECT timed out")
		}
		if err := token.Error(); !a.So(err, should.BeNil) {
			t.Fatal("CONNECT returned an error")
		}

		// TODO: monitor cluster claim on IDs https://github.com/TheThingsIndustries/lorawan-stack/issues/941
	})
	if !ok {
		t.FailNow()
	}

	t.Run("v3API", func(t *testing.T) {
		status := ttnpb.NewPopulatedGatewayStatus(test.Randy, false)
		statusStart := time.Now()
		statusTopic := topic.Join([]string{
			gatewayserver.V3TopicPrefix,
			registeredGatewayUID,
			gatewayserver.StatusTopicSuffix,
		})
		ok = t.Run("Status", func(t *testing.T) {
			a := assertions.New(t)

			statusBytes, err := status.Marshal()
			if !a.So(err, should.BeNil) {
				t.Fatal("Could not marshal status")
			}

			token := client.Publish(statusTopic, 0x00, false, statusBytes)
			if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
				t.Fatal("PUBLISH timed out")
			}
			if err := token.Error(); !a.So(err, should.BeNil) {
				t.Fatal("PUBLISH returned an error")
			}
		})
		if !ok {
			t.FailNow()
		}

		uplinkStart := time.Now()
		uplinkTopic := topic.Join([]string{
			gatewayserver.V3TopicPrefix,
			registeredGatewayUID,
			gatewayserver.UplinkTopicSuffix,
		})
		ok := t.Run("Uplink", func(t *testing.T) {
			a := assertions.New(t)

			uplink := ttnpb.NewPopulatedUplinkMessage(test.Randy, false)
			uplinkBytes, err := uplink.Marshal()
			if !a.So(err, should.BeNil) {
				t.Fatal("Could not marshal uplink")
			}

			token := client.Publish(uplinkTopic, 0x00, false, uplinkBytes)
			if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
				t.Fatal("PUBLISH timed out")
			}
			if err := token.Error(); !a.So(err, should.BeNil) {
				t.Fatal("PUBLISH returned an error")
			}

			select {
			case msg := <-ns.messageReceived:
				if msg != "HandleUplink" {
					t.Fatal("Expected Gateway Server to call HandleUplink on the Network Server, instead received", msg)
				}
			case <-time.After(nsReceptionTimeout):
				t.Fatal("The Gateway Server never called the Network Server's HandleUplink to handle the join-request. This might be due to an unexpected error in the GatewayServer.handleMQTTConnection() function.")
			}
		})
		if !ok {
			t.FailNow()
		}

		newDownlink := make(chan bool)
		downlinksHandler := func(mqtt.Client, mqtt.Message) {
			newDownlink <- true
		}

		ok = t.Run("Subscription", func(t *testing.T) {
			a := assertions.New(t)

			for _, subscription := range []struct {
				success bool
				topic   string
			}{
				{
					success: false,
					topic:   registeredGatewayUID,
				},
				{
					success: false,
					topic: topic.Join([]string{
						"v2",
						registeredGatewayUID,
						gatewayserver.DownlinkTopicSuffix,
					}),
				},
				{
					success: true,
					topic: topic.Join([]string{
						gatewayserver.V3TopicPrefix,
						registeredGatewayUID,
						gatewayserver.DownlinkTopicSuffix,
					}),
				},
				{
					success: false,
					topic: topic.Join([]string{
						gatewayserver.V3TopicPrefix,
						registeredGatewayUID,
						gatewayserver.UplinkTopicSuffix,
					}),
				},
				{
					success: false,
					topic: topic.Join([]string{
						gatewayserver.V3TopicPrefix,
						"random-gateway",
						gatewayserver.DownlinkTopicSuffix,
					}),
				},
			} {
				token := client.Subscribe(subscription.topic, 0x00, downlinksHandler)
				if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
					t.Fatal("SUBSCRIBE timed out")
				}
				subToken := token.(*mqtt.SubscribeToken)
				err := token.Error()
				switch subscription.success {
				case true:
					if err != nil {
						t.Fatalf("SUBSCRIBE returned an error for %s, but it should have succeeded", subscription.topic)
					}
					qos := subToken.Result()[subscription.topic]
					if !a.So(qos, should.BeLessThanOrEqualTo, 0x02) {
						t.Fatalf("SUBSCRIBE returned an error code for %d %s", qos, subscription.topic)
					}
				case false:
					qos := subToken.Result()[subscription.topic]
					if err != nil && !a.So(qos, should.BeGreaterThan, 0x02) {
						t.Fatalf("SUBSCRIBE succeeded to subscribe to a topic for %s with QoS %d, but it should have failed", subscription.topic, qos)
					}
				}
			}
		})
		if !ok {
			t.FailNow()
		}

		downlinksStart := time.Now()
		ok = t.Run("Downlinks", func(t *testing.T) {
			a := assertions.New(t)

			downlink := ttnpb.NewPopulatedDownlinkMessage(test.Randy, false)
			downlink.TxMetadata.GatewayIdentifiers = ttnpb.GatewayIdentifiers{
				GatewayID: gtwID.GatewayID,
			}
			downlink.Settings.Frequency = 863000000
			downlink.Settings.SpreadingFactor = 7
			downlink.Settings.Bandwidth = 125000
			downlink.Settings.CodingRate = "4/5"
			_, err := gs.ScheduleDownlink(ctx, downlink)
			a.So(err, should.BeNil)

			select {
			case <-newDownlink:
			case <-time.After(mqttConnectionTimeout):
				t.Fatal("Downlink never received over MQTT")
			}
		})
		if !ok {
			t.FailNow()
		}

		ctx := rights.NewContextWithFetcher(ctx, rights.FetcherFunc(func(context.Context, ttnpb.Identifiers) ([]ttnpb.Right, error) {
			return []ttnpb.Right{ttnpb.RIGHT_GATEWAY_STATUS_READ}, nil
		}))
		t.Run("Statistics", func(t *testing.T) {
			a := assertions.New(t)
			observations, err := gs.GetGatewayObservations(ctx, &gtwID)
			a.So(err, should.BeNil)

			if !a.So(*observations.GetLastDownlinkReceivedAt(), should.HappenAfter, downlinksStart) {
				t.Fatal("Expected last downlink reception time to have been recorded, but it wasn't")
			}
			if !a.So(*observations.GetLastUplinkReceivedAt(), should.HappenAfter, uplinkStart) {
				t.Fatal("Expected last uplink reception time to have been recorded, but it wasn't")
			}
			if !a.So(*observations.GetLastStatusReceivedAt(), should.HappenAfter, statusStart) {
				t.Fatal("Expected last status reception time to have been recorded, but it wasn't")
			}
			a.So(pretty.Diff(observations.GetLastStatus(), status), should.BeEmpty)
		})

		t.Run("Invalid messages", func(t *testing.T) {
			a := assertions.New(t)

			invalidStatusSent := time.Now()
			token := client.Publish(statusTopic, 0x00, false, []byte{0x00, 0xaa})
			if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
				t.Fatal("PUBLISH timed out when sending the invalid uplink")
			}
			if err := token.Error(); !a.So(err, should.BeNil) {
				t.Fatal("PUBLISH returned an error")
			}

			invalidUplinkSent := time.Now()
			token = client.Publish(uplinkTopic, 0x00, false, []byte{0x00, 0xaa})
			if ok := token.WaitTimeout(mqttConnectionTimeout); !a.So(ok, should.BeTrue) {
				t.Fatal("PUBLISH timed out when sending the invalid uplink")
			}
			if err := token.Error(); !a.So(err, should.BeNil) {
				t.Fatal("PUBLISH returned an error")
			}

			observations, err := gs.GetGatewayObservations(ctx, &gtwID)
			a.So(err, should.BeNil)

			if !a.So(*observations.GetLastStatusReceivedAt(), should.HappenBefore, invalidStatusSent) {
				t.Fatal("Expected invalid status to not have been handled, but it was")
			}
			if !a.So(*observations.GetLastUplinkReceivedAt(), should.HappenBefore, invalidUplinkSent) {
				t.Fatal("Expected invalid uplink to not have been handled, but it was")
			}
		})

		t.Run("Disconnect", func(t *testing.T) {
			client.Disconnect(0)

			// TODO: monitor cluster claim on IDs https://github.com/TheThingsIndustries/lorawan-stack/issues/941
		})
	})
}
