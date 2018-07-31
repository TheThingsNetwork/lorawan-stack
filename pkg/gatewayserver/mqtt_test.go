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
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/gatewayserver"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

const mqttConnectionTimeout = 3 * time.Second

var (
	registeredGatewayUID = "registered-gateway"
	registeredGatewayEUI = types.EUI64{0xAA, 0xEE, 0x00, 0x00, 0x00, 0x00, 0x00, 0x00}
)

// TODO: Refactor TestMQTTConnection/TestUDP(/TestLink?)
func TestMQTTConnection(t *testing.T) {
	a := assertions.New(t)

	logger := test.GetLogger(t)
	ctx := log.NewContext(test.Context(), logger)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	store, err := test.NewFrequencyPlansStore()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer store.Destroy()

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
			FrequencyPlans: config.FrequencyPlans{
				StoreDirectory: store.Directory(),
			},
			GRPC: config.GRPC{
				AllowInsecureForCredentials: true,
			},
		},
	})
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

		select {
		case msg := <-ns.messageReceived:
			if msg != "StartServingGateway" {
				t.Fatal("Expected Gateway Server to call StartServingGateway on the Network Server, instead received", msg)
			}
		case <-time.After(nsReceptionTimeout):
			t.Fatal("The Gateway Server never called the Network Server's StartServingGateway to handle the join-request. This might be due to an unexpected error in the GatewayServer.handleMQTTConnection() function.")
		}
	})
	if !ok {
		t.FailNow()
	}

	t.Run("v3API", func(t *testing.T) {
		ok := t.Run("Uplink", func(t *testing.T) {
			a := assertions.New(t)

			uplinkTopic := topic.Join([]string{
				gatewayserver.V3TopicPrefix,
				registeredGatewayUID,
				gatewayserver.UplinkTopicSuffix,
			})
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

		t.Run("Disconnect", func(t *testing.T) {
			client.Disconnect(0)

			select {
			case msg := <-ns.messageReceived:
				if msg != "StopServingGateway" {
					t.Fatal("Expected Gateway Server to call StopServingGateway on the Network Server, instead received", msg)
				}
			case <-time.After(nsReceptionTimeout):
				t.Fatal("The Gateway Server never called the Network Server's StopServingGateway to handle the join-request. This might be due to an unexpected error in the GatewayServer.handleMQTTConnection() function.")
			}
		})
	})
}
