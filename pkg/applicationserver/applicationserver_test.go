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

package applicationserver_test

import (
	"bytes"
	"context"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	mqttserver "github.com/TheThingsIndustries/mystique/pkg/server"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	pbtypes "github.com/gogo/protobuf/types"
	nats_server "github.com/nats-io/nats-server/v2/server"
	nats_test_server "github.com/nats-io/nats-server/v2/test"
	nats_client "github.com/nats-io/nats.go"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver"
	iopubsubredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/redis"
	iowebredis "go.thethings.network/lorawan-stack/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/config"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/unique"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

type connChannels struct {
	up          chan *ttnpb.ApplicationUp
	downPush    chan *ttnpb.DownlinkQueueRequest
	downReplace chan *ttnpb.DownlinkQueueRequest
	downErr     chan error
}

func TestApplicationServer(t *testing.T) {
	a := assertions.New(t)

	// This application will be added to the Entity Registry and to the link registry of the Application Server so that it
	// links automatically on start to the Network Server.
	registeredApplicationID := ttnpb.ApplicationIdentifiers{ApplicationID: "foo-app"}
	registeredApplicationKey := "secret"
	registeredApplicationFormatter := ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
	registeredApplicationWebhookID := ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		WebhookID:              "test",
	}
	registeredApplicationPubSubID := ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		PubSubID:               "test",
	}

	// This device gets registered in the device registry of the Application Server.
	registeredDevice := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			DeviceID:               "foo-device",
			JoinEUI:                eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			DevEUI:                 eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		},
		VersionIDs: &ttnpb.EndDeviceVersionIdentifiers{
			BrandID:         "thethingsproducts",
			ModelID:         "thethingsnode",
			HardwareVersion: "1.0",
			FirmwareVersion: "1.1",
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter:   ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
			DownFormatter: ttnpb.PayloadFormatter_FORMATTER_REPOSITORY,
		},
	}

	// This device does not get registered in the device registry of the Application Server and will be created on join
	// and on uplink.
	unregisteredDeviceID := ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		DeviceID:               "bar-device",
		JoinEUI:                eui64Ptr(types.EUI64{0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
		DevEUI:                 eui64Ptr(types.EUI64{0x24, 0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
	}

	deviceRepositoryData := map[string][]byte{
		"brands.yml": []byte(`version: '3'
brands:
thethingsproducts:
  name: The Things Products
  url: https://www.thethingsnetwork.org`),
		"thethingsproducts/devices.yml": []byte(`version: '3'
devices:
  thethingsnode:
    name: The Things Node`),
		"thethingsproducts/thethingsnode/versions.yml": []byte(`version: '3'
hardware_versions:
  '1.0':
    - firmware_version: 1.1
      payload_format:
        up:
          type: javascript
          parameter: decoder.js
        down:
          type: javascript
          parameter: encoder.js`),
		"thethingsproducts/thethingsnode/1.0/decoder.js": []byte(`function Decoder(payload, f_port) {
	var sum = 0;
	for (i = 0; i < payload.length; i++) {
		sum += payload[i];
	}
	return {
		sum: sum
	};
}`),
		"thethingsproducts/thethingsnode/1.0/encoder.js": []byte(`function Encoder(payload, f_port) {
	var res = [];
	for (i = 0; i < payload.sum; i++) {
		res[i] = 1;
	}
	return res;
}`)}

	ctx := test.Context()
	is, isAddr := startMockIS(ctx)
	js, jsAddr := startMockJS(ctx)
	ns, nsAddr := startMockNS(ctx, func(md rpcmetadata.MD) bool {
		return md.AuthType == "Bearer" && md.AuthValue == registeredApplicationKey
	})

	// Register the application in the Entity Registry.
	is.add(ctx, registeredApplicationID, registeredApplicationKey)

	// Register some sessions in the Join Server. Sometimes the keys are sent by the Network Server as part of the
	// join-accept, and sometimes they are not sent by the Network Server so the Application Server gets them from the
	// Join Server.
	js.add(ctx, *registeredDevice.DevEUI, []byte{0x11}, ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.
		EncryptedKey: []byte{0xa8, 0x11, 0x8f, 0x80, 0x2e, 0xbf, 0x8, 0xdc, 0x62, 0x37, 0xc3, 0x4, 0x63, 0xa2, 0xfa, 0xcb, 0xf8, 0x87, 0xaa, 0x31, 0x90, 0x23, 0x85, 0xc1},
		KEKLabel:     "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, []byte{0x22}, ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
		EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
		KEKLabel:     "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, []byte{0x33}, ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
		EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
		KEKLabel:     "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, []byte{0x44}, ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}
		EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
		KEKLabel:     "test",
	})
	js.add(ctx, *registeredDevice.DevEUI, []byte{0x55}, ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
		EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
		KEKLabel:     "test",
	})

	devsRedisClient, devsFlush := test.NewRedis(t, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient}

	linksRedisClient, linksFlush := test.NewRedis(t, "applicationserver_test", "links")
	defer linksFlush()
	defer linksRedisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: linksRedisClient}
	_, err := linkRegistry.Set(ctx, registeredApplicationID, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{
			APIKey: registeredApplicationKey,
			DefaultFormatters: &ttnpb.MessagePayloadFormatters{
				UpFormatter:   registeredApplicationFormatter,
				DownFormatter: registeredApplicationFormatter,
			},
		}, []string{"api_key", "default_formatters"}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set link in registry: %s", err)
	}

	webhooksRedisClient, webhooksFlush := test.NewRedis(t, "applicationserver_test", "webhooks")
	defer webhooksFlush()
	defer webhooksRedisClient.Close()
	webhookRegistry := iowebredis.WebhookRegistry{Redis: webhooksRedisClient}

	pubsubRedisClient, pubsubFlush := test.NewRedis(t, "applicationserver_test", "pubsub")
	defer pubsubFlush()
	defer pubsubRedisClient.Close()
	pubsubRegistry := iopubsubredis.PubSubRegistry{Redis: pubsubRedisClient}

	natsServer := nats_test_server.RunServer(&nats_server.Options{
		Host:           "127.0.0.1",
		Port:           4124,
		NoLog:          true,
		NoSigs:         true,
		MaxControlLine: 256,
	})
	defer natsServer.Shutdown()
	time.Sleep(Timeout)

	mqttServer := mqttserver.New(ctx)
	mqttLis, err := mqttnet.Listen("tcp", ":0")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	defer mqttLis.Close()
	go func() {
		for {
			conn, err := mqttLis.Accept()
			if err != nil {
				return
			}
			go mqttServer.Handle(conn)
		}
	}()
	time.Sleep(Timeout)

	c := component.MustNew(test.GetLogger(t), &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9184",
				AllowInsecureForCredentials: true,
			},
			HTTP: config.HTTP{
				Listen: ":8099",
			},
			Cluster: config.Cluster{
				IdentityServer: isAddr,
				JoinServer:     jsAddr,
				NetworkServer:  nsAddr,
			},
			DeviceRepository: config.DeviceRepositoryConfig{
				Static: deviceRepositoryData,
			},
			KeyVault: config.KeyVault{
				Static: map[string][]byte{
					"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			},
		},
	})
	config := &applicationserver.Config{
		LinkMode: "all",
		Devices:  deviceRegistry,
		Links:    linkRegistry,
		MQTT: applicationserver.MQTTConfig{
			Listen: ":1883",
		},
		Webhooks: applicationserver.WebhooksConfig{
			Registry:  webhookRegistry,
			Target:    "direct",
			Timeout:   Timeout,
			QueueSize: 1,
		},
		PubSub: applicationserver.PubSubConfig{
			Registry: pubsubRegistry,
		},
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_APPLICATION_SERVER)

	test.Must(nil, c.Start())
	defer c.Close()
	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_JOIN_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	// Delay for the AS-NS link to establish.
	time.Sleep(Timeout)

	for _, ptc := range []struct {
		Protocol         string
		ValidAuth        func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool
		Connect          func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error
		SkipCheckDownErr bool
	}{
		{
			Protocol: "grpc",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewAppAsClient(as.LoopbackConn())
				stream, err := client.Subscribe(ctx, &ids, creds)
				if err != nil {
					return err
				}
				errCh := make(chan error, 1)
				// Read upstream.
				go func() {
					for {
						msg, err := stream.Recv()
						if err != nil {
							errCh <- err
							return
						}
						chs.up <- msg
					}
				}()
				// Write downstream.
				go func() {
					for {
						var err error
						select {
						case <-ctx.Done():
							return
						case req := <-chs.downPush:
							_, err = client.DownlinkQueuePush(ctx, req, creds)
						case req := <-chs.downReplace:
							_, err = client.DownlinkQueueReplace(ctx, req, creds)
						}
						chs.downErr <- err
					}
				}()
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
		},
		{
			Protocol: "mqtt",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker("tcp://0.0.0.0:1883")
				clientOpts.SetUsername(unique.ID(ctx, ids))
				clientOpts.SetPassword(key)
				client := mqtt.NewClient(clientOpts)
				if token := client.Connect(); !token.WaitTimeout(Timeout) {
					return errors.New("connect timeout")
				} else if token.Error() != nil {
					return token.Error()
				}
				defer client.Disconnect(uint(Timeout / time.Millisecond))
				errCh := make(chan error, 1)
				// Write downstream.
				go func() {
					for {
						var req *ttnpb.DownlinkQueueRequest
						var topicFmt string
						select {
						case <-ctx.Done():
							return
						case req = <-chs.downPush:
							topicFmt = "v3/%v/devices/%v/down/push"
						case req = <-chs.downReplace:
							topicFmt = "v3/%v/devices/%v/down/replace"
						}
						msg := &ttnpb.ApplicationDownlinks{
							Downlinks: req.Downlinks,
						}
						buf, err := jsonpb.TTN().Marshal(msg)
						if err != nil {
							chs.downErr <- err
							continue
						}
						token := client.Publish(fmt.Sprintf(topicFmt, unique.ID(ctx, req.ApplicationIdentifiers), req.DeviceID), 1, false, buf)
						token.Wait()
						chs.downErr <- token.Error()
					}
				}()
				// Read upstream.
				token := client.Subscribe(fmt.Sprintf("v3/%v/devices/#", unique.ID(ctx, ids)), 1, func(_ mqtt.Client, raw mqtt.Message) {
					msg := &ttnpb.ApplicationUp{}
					if err := jsonpb.TTN().Unmarshal(raw.Payload(), msg); err != nil {
						errCh <- err
						return
					}
					chs.up <- msg
				})
				if token.Wait() && token.Error() != nil {
					return token.Error()
				}
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			SkipCheckDownErr: true, // There is no direct error response in MQTT.
		},
		{
			Protocol: "pubsub/nats",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				// Configure pubsub.
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewApplicationPubSubRegistryClient(as.LoopbackConn())
				req := &ttnpb.SetApplicationPubSubRequest{
					ApplicationPubSub: ttnpb.ApplicationPubSub{
						ApplicationPubSubIdentifiers: registeredApplicationPubSubID,
						Provider: &ttnpb.ApplicationPubSub_NATS{
							NATS: &ttnpb.ApplicationPubSub_NATSProvider{
								ServerURL: "nats://localhost:4124",
							},
						},
						Format:    "json",
						BaseTopic: "foo.bar",
						DownlinkPush: &ttnpb.ApplicationPubSub_Message{
							Topic: "down.downlink.push",
						},
						DownlinkReplace: &ttnpb.ApplicationPubSub_Message{
							Topic: "down.downlink.replace",
						},
						UplinkMessage: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.uplink.message",
						},
						JoinAccept: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.join.accept",
						},
						DownlinkAck: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlink.ack",
						},
						DownlinkNack: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlink.nack",
						},
						DownlinkSent: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlink.sent",
						},
						DownlinkFailed: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlnk.failed",
						},
						DownlinkQueued: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlink.queued",
						},
						LocationSolved: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.location.solved",
						},
					},
					FieldMask: pbtypes.FieldMask{
						Paths: []string{
							"base_topic",
							"downlink_ack",
							"downlink_failed",
							"downlink_nack",
							"downlink_queued",
							"downlink_sent",
							"downlink_push",
							"downlink_replace",
							"format",
							"provider",
							"join_accept",
							"location_solved",
							"uplink_message",
						},
					},
				}
				if _, err := client.Set(ctx, req, creds); err != nil {
					return err
				}
				nc, err := nats_client.Connect("nats://localhost:4124")
				if err != nil {
					return err
				}
				defer nc.Close()
				// Write downstream.
				go func() {
					for {
						var req *ttnpb.DownlinkQueueRequest
						var subject string
						select {
						case <-ctx.Done():
							return
						case req = <-chs.downPush:
							subject = "foo.bar.down.downlink.push"
						case req = <-chs.downReplace:
							subject = "foo.bar.down.downlink.replace"
						}
						buf, err := jsonpb.TTN().Marshal(req)
						if err != nil {
							chs.downErr <- err
							continue
						}
						chs.downErr <- nc.Publish(subject, buf)
						err = nc.FlushTimeout(Timeout)
						if err != nil {
							chs.downErr <- err
						}
					}
				}()
				errCh := make(chan error, 1)
				// Read upstream.
				subscription, err := nc.Subscribe("foo.bar.up.>", func(raw *nats_client.Msg) {
					msg := &ttnpb.ApplicationUp{}
					err = jsonpb.TTN().Unmarshal(raw.Data, msg)
					if err != nil {
						errCh <- err
						return
					}
					chs.up <- msg
				})
				if err != nil {
					return err
				}
				defer subscription.Unsubscribe()
				select {
				case err = <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			SkipCheckDownErr: true, // There is no direct error response in PubSub.
		},
		{
			Protocol: "pubsub/mqtt",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				// Configure pubsub.
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewApplicationPubSubRegistryClient(as.LoopbackConn())
				req := &ttnpb.SetApplicationPubSubRequest{
					ApplicationPubSub: ttnpb.ApplicationPubSub{
						ApplicationPubSubIdentifiers: registeredApplicationPubSubID,
						Provider: &ttnpb.ApplicationPubSub_MQTT{
							MQTT: &ttnpb.ApplicationPubSub_MQTTProvider{
								ServerURL:    fmt.Sprintf("tcp://%v", mqttLis.Addr()),
								PublishQoS:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
								SubscribeQoS: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
							},
						},
						Format:    "json",
						BaseTopic: "foo/bar",
						DownlinkPush: &ttnpb.ApplicationPubSub_Message{
							Topic: "down/downlink/push",
						},
						DownlinkReplace: &ttnpb.ApplicationPubSub_Message{
							Topic: "down/downlink/replace",
						},
						UplinkMessage: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/uplink/message",
						},
						JoinAccept: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/join/accept",
						},
						DownlinkAck: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlink/ack",
						},
						DownlinkNack: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlink/nack",
						},
						DownlinkSent: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlink/sent",
						},
						DownlinkFailed: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlnk/failed",
						},
						DownlinkQueued: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlink/queued",
						},
						LocationSolved: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/location/solved",
						},
					},
					FieldMask: pbtypes.FieldMask{
						Paths: []string{
							"base_topic",
							"downlink_ack",
							"downlink_failed",
							"downlink_nack",
							"downlink_queued",
							"downlink_sent",
							"downlink_push",
							"downlink_replace",
							"format",
							"provider",
							"join_accept",
							"location_solved",
							"uplink_message",
						},
					},
				}
				if _, err := client.Set(ctx, req, creds); err != nil {
					return err
				}
				clientOpts := mqtt.NewClientOptions()
				clientOpts.AddBroker(mqttLis.Addr().String())
				mqttClient := mqtt.NewClient(clientOpts)
				a.So(client, should.NotBeNil)
				if token := mqttClient.Connect(); !token.WaitTimeout(Timeout) {
					return errors.New("connect timeout")
				} else if token.Error() != nil {
					return token.Error()
				}
				defer mqttClient.Disconnect(uint(Timeout / time.Millisecond))

				// Write downstream.
				go func() {
					for {
						var req *ttnpb.DownlinkQueueRequest
						var topic string
						select {
						case <-ctx.Done():
							return
						case req = <-chs.downPush:
							topic = "foo/bar/down/downlink/push"
						case req = <-chs.downReplace:
							topic = "foo/bar/down/downlink/replace"
						}
						buf, err := jsonpb.TTN().Marshal(req)
						if err != nil {
							chs.downErr <- err
							continue
						}
						token := mqttClient.Publish(topic, 1, false, buf)
						token.Wait()
						chs.downErr <- token.Error()
					}
				}()
				errCh := make(chan error, 1)
				// Read upstream.
				token := mqttClient.Subscribe("foo/bar/up/#", 1, func(_ mqtt.Client, raw mqtt.Message) {
					msg := &ttnpb.ApplicationUp{}
					if err := jsonpb.TTN().Unmarshal(raw.Payload(), msg); err != nil {
						errCh <- err
						return
					}
					chs.up <- msg
				})
				if token.Wait() && token.Error() != nil {
					return token.Error()
				}
				select {
				case err := <-errCh:
					return err
				case <-ctx.Done():
					return ctx.Err()
				}
			},
			SkipCheckDownErr: true, // There is no direct error response in PubSub.
		},
		{
			Protocol: "webhooks",
			ValidAuth: func(ctx context.Context, ids ttnpb.ApplicationIdentifiers, key string) bool {
				return ids == registeredApplicationID && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				// Start web server to read upstream.
				webhookTarget := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
					buf, err := ioutil.ReadAll(req.Body)
					if !a.So(err, should.BeNil) {
						t.FailNow()
					}
					msg := &ttnpb.ApplicationUp{}
					if err := jsonpb.TTN().Unmarshal(buf, msg); !a.So(err, should.BeNil) {
						t.FailNow()
					}
					chs.up <- msg
					res.WriteHeader(http.StatusAccepted)
				}))
				defer webhookTarget.Close()
				// Configure webhook.
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewApplicationWebhookRegistryClient(as.LoopbackConn())
				req := &ttnpb.SetApplicationWebhookRequest{
					ApplicationWebhook: ttnpb.ApplicationWebhook{
						ApplicationWebhookIdentifiers: registeredApplicationWebhookID,
						BaseURL:                       webhookTarget.URL,
						Format:                        "json",
						UplinkMessage:                 &ttnpb.ApplicationWebhook_Message{Path: ""},
						JoinAccept:                    &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkAck:                   &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkNack:                  &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkQueued:                &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkSent:                  &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkFailed:                &ttnpb.ApplicationWebhook_Message{Path: ""},
						LocationSolved:                &ttnpb.ApplicationWebhook_Message{Path: ""},
					},
					FieldMask: pbtypes.FieldMask{
						Paths: []string{
							"base_url",
							"format",
							"uplink_message",
							"join_accept",
							"downlink_ack",
							"downlink_nack",
							"downlink_queued",
							"downlink_sent",
							"downlink_failed",
							"location_solved",
						},
					},
				}
				if _, err := client.Set(ctx, req, creds); err != nil {
					return err
				}
				// Write downstream.
				go func() {
					for {
						var data *ttnpb.DownlinkQueueRequest
						var action string
						select {
						case data = <-chs.downPush:
							action = "push"
						case data = <-chs.downReplace:
							action = "replace"
						}
						buf, err := jsonpb.TTN().Marshal(&ttnpb.ApplicationDownlinks{Downlinks: data.Downlinks})
						if err != nil {
							chs.downErr <- err
							continue
						}
						url := fmt.Sprintf("http://127.0.0.1:8099/api/v3/as/applications/%s/webhooks/%s/devices/%s/down/%s",
							data.ApplicationID, registeredApplicationWebhookID.WebhookID, data.DeviceID, action,
						)
						req, err := http.NewRequest(http.MethodPost, url, bytes.NewReader(buf))
						if err != nil {
							chs.downErr <- err
							continue
						}
						req.Header.Set("Content-Type", "application/json")
						req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", key))
						res, err := http.DefaultClient.Do(req)
						if err == nil && (res.StatusCode < 200 || res.StatusCode > 299) {
							err = errors.FromHTTPStatusCode(res.StatusCode)
						}
						chs.downErr <- err
					}
				}()
				<-ctx.Done()
				return ctx.Err()
			},
		},
	} {
		t.Run(fmt.Sprintf("Authenticate/%v", ptc.Protocol), func(t *testing.T) {
			for _, ctc := range []struct {
				Name string
				ID   ttnpb.ApplicationIdentifiers
				Key  string
			}{
				{
					Name: "ValidIDAndKey",
					ID:   registeredApplicationID,
					Key:  registeredApplicationKey,
				},
				{
					Name: "InvalidKey",
					ID:   registeredApplicationID,
					Key:  "invalid-key",
				},
				{
					Name: "InvalidIDAndKey",
					ID:   ttnpb.ApplicationIdentifiers{ApplicationID: "invalid-application"},
					Key:  "invalid-key",
				},
			} {
				t.Run(ctc.Name, func(t *testing.T) {
					ctx, cancel := context.WithDeadline(ctx, time.Now().Add(Timeout))
					chs := &connChannels{
						up:          make(chan *ttnpb.ApplicationUp, 1),
						downPush:    make(chan *ttnpb.DownlinkQueueRequest),
						downReplace: make(chan *ttnpb.DownlinkQueueRequest),
						downErr:     make(chan error, 1),
					}
					err := ptc.Connect(ctx, t, ctc.ID, ctc.Key, chs)
					cancel()
					if errors.IsDeadlineExceeded(err) {
						if !ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
							t.Fatal("Expected link error due to invalid auth")
						}
					} else if ptc.ValidAuth(ctx, ctc.ID, ctc.Key) {
						t.Fatalf("Expected deadline exceeded with valid auth, but have %v", err)
					}
				})
			}
		})

		t.Run(fmt.Sprintf("Traffic/%v", ptc.Protocol), func(t *testing.T) {
			ctx, cancel := context.WithCancel(ctx)
			chs := &connChannels{
				up:          make(chan *ttnpb.ApplicationUp, 1),
				downPush:    make(chan *ttnpb.DownlinkQueueRequest),
				downReplace: make(chan *ttnpb.DownlinkQueueRequest),
				downErr:     make(chan error, 1),
			}

			wg := &sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := ptc.Connect(ctx, t, registeredApplicationID, registeredApplicationKey, chs)
				if !errors.IsCanceled(err) {
					t.Fatalf("Expected context canceled, but have %v", err)
				}
			}()
			// Wait for connection to establish.
			time.Sleep(Timeout)

			t.Run("Upstream", func(t *testing.T) {
				ns.reset()
				devsFlush()
				deviceRegistry.Set(ctx, registeredDevice.EndDeviceIdentifiers, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					return registeredDevice, []string{"ids", "version_ids", "formatters"}, nil
				})

				for _, tc := range []struct {
					Name          string
					IDs           ttnpb.EndDeviceIdentifiers
					ResetQueue    []*ttnpb.ApplicationDownlink
					Message       *ttnpb.ApplicationUp
					ExpectTimeout bool
					AssertUp      func(t *testing.T, up *ttnpb.ApplicationUp)
					AssertDevice  func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink)
				}{
					{
						Name: "RegisteredDevice/JoinAccept",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x11, 0x11, 0x11, 0x11}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x11},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x11, 0x11, 0x11, 0x11}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: []byte{0x11},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.BeNil)
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x11, 0x11, 0x11, 0x11},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x11},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0xa8, 0x11, 0x8f, 0x80, 0x2e, 0xbf, 0x8, 0xdc, 0x62, 0x37, 0xc3, 0x4, 0x63, 0xa2, 0xfa, 0xcb, 0xf8, 0x87, 0xaa, 0x31, 0x90, 0x23, 0x85, 0xc1},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 0,
								StartedAt:     dev.PendingSession.StartedAt,
							})
							a.So(queue, should.BeEmpty)
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KEKLabel:     "test",
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: []byte{0x22},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.BeNil)
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x22, 0x22, 0x22, 0x22},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 0,
								StartedAt:     dev.PendingSession.StartedAt,
							})
							a.So(queue, should.BeEmpty)
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/PendingSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x22},
									FPort:        22,
									FCnt:         22,
									FRMPayload:   []byte{0x01},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: []byte{0x22},
										FPort:        22,
										FCnt:         22,
										FRMPayload:   []byte{0xc1},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 193, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey/WithQueue",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KEKLabel:     "test",
									},
									InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyID: []byte{0x22},
											FPort:        11,
											FCnt:         11,
											FRMPayload:   []byte{0x69, 0x65, 0x9f, 0x8f},
										},
										{
											SessionKeyID: []byte{0x22},
											FPort:        22,
											FCnt:         22,
											FRMPayload:   []byte{0xb, 0x8f, 0x94, 0xe6},
										},
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID: []byte{0x33},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x22, 0x22, 0x22, 0x22},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 2,
								StartedAt:     dev.Session.StartedAt,
							})
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x33, 0x33, 0x33, 0x33},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 0,
								StartedAt:     dev.PendingSession.StartedAt,
							})
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x22},
									FPort:        11,
									FCnt:         1,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x22},
									FPort:        22,
									FCnt:         2,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/CurrentSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0xca, 0xa9, 0x42},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x01, 0x02, 0x03},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 6, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Queued",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkQueued{
								DownlinkQueued: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkQueued{
									DownlinkQueued: &ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Sent",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkSent{
								DownlinkSent: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkSent{
									DownlinkSent: &ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Failed",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									ApplicationDownlink: ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
									},
									Error: ttnpb.ErrorDetails{
										Name: "test",
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkFailed{
									DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
										ApplicationDownlink: ttnpb.ApplicationDownlink{
											SessionKeyID: []byte{0x33},
											FPort:        42,
											FCnt:         42,
											FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
										},
										Error: ttnpb.ErrorDetails{
											Name: "test",
										},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Ack",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkAck{
								DownlinkAck: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkAck{
									DownlinkAck: &ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Nack",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						ResetQueue: []*ttnpb.ApplicationDownlink{ // Pop the first item; it will be inserted because of the nack.
							{
								SessionKeyID: []byte{0x33},
								FPort:        22,
								FCnt:         2,
								FRMPayload:   []byte{0x92, 0xfe, 0x93, 0xf5},
							},
						},
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkNack{
								DownlinkNack: &ttnpb.ApplicationDownlink{
									SessionKeyID: []byte{0x33},
									FPort:        11,
									FCnt:         1,
									FRMPayload:   []byte{0x5f, 0x38, 0x7c, 0xb0},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkNack{
									DownlinkNack: &ttnpb.ApplicationDownlink{
										SessionKeyID: []byte{0x33},
										FPort:        11,
										FCnt:         1,
										FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{ // The nacked item is inserted first.
									SessionKeyID: []byte{0x33},
									FPort:        11,
									FCnt:         2,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x33},
									FPort:        22,
									FCnt:         3,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey/WithQueue/WithPendingSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KEKLabel:     "test",
									},
									PendingSession: true,
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyID:   []byte{0x44},
										PendingSession: true,
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x33, 0x33, 0x33, 0x33},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 3,
								StartedAt:     dev.Session.StartedAt,
							})
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x44, 0x44, 0x44, 0x44},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 0,
								StartedAt:     dev.PendingSession.StartedAt,
							})
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x33},
									FPort:        11,
									FCnt:         2,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x33},
									FPort:        22,
									FCnt:         3,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/PendingSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x44},
									FPort:        24,
									FCnt:         24,
									FRMPayload:   []byte{0x14, 0x4e, 0x3c},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: []byte{0x44},
										FPort:        24,
										FCnt:         24,
										FRMPayload:   []byte{0x64, 0x64, 0x64},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 300, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x44, 0x44, 0x44, 0x44},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 2,
								StartedAt:     dev.Session.StartedAt,
							})
							a.So(dev.PendingSession, should.BeNil)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x44},
									FPort:        11,
									FCnt:         1,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x44},
									FPort:        22,
									FCnt:         2,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkQueueInvalidated/KnownSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyID: []byte{0x44},
											FPort:        11,
											FCnt:         11,
											FRMPayload:   []byte{0x65, 0x98, 0xa7, 0xfc},
										},
										{
											SessionKeyID: []byte{0x44},
											FPort:        22,
											FCnt:         22,
											FRMPayload:   []byte{0x1b, 0x4b, 0x97, 0xb9},
										},
									},
									LastFCntDown: 42,
								},
							},
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session.LastAFCntDown, should.Equal, 44)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x44},
									FPort:        11,
									FCnt:         43,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x44},
									FPort:        22,
									FCnt:         44,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkQueueInvalidated/UnknownSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyID: []byte{0x44},
											FPort:        11,
											FCnt:         11,
											FRMPayload:   []byte{0x65, 0x98, 0xa7, 0xfc},
										},
										{
											SessionKeyID: []byte{0x11, 0x22, 0x33, 0x44},
											FPort:        12,
											FCnt:         12,
											FRMPayload:   []byte{0xff, 0xff, 0xff, 0xff},
										},
										{
											SessionKeyID: []byte{0x44},
											FPort:        22,
											FCnt:         22,
											FRMPayload:   []byte{0x1b, 0x4b, 0x97, 0xb9},
										},
									},
									LastFCntDown: 84,
								},
							},
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session.LastAFCntDown, should.Equal, 86)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x44},
									FPort:        11,
									FCnt:         85,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x44},
									FPort:        22,
									FCnt:         86,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/KnownSession",
						IDs:  registeredDevice.EndDeviceIdentifiers,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x55},
									FPort:        42,
									FCnt:         42,
									FRMPayload:   []byte{0xd1, 0x43, 0x6a},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIdentifiers: withDevAddr(registeredDevice.EndDeviceIdentifiers, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										SessionKeyID: []byte{0x55},
										FPort:        42,
										FCnt:         42,
										FRMPayload:   []byte{0x2a, 0x2a, 0x2a},
										DecodedPayload: &pbtypes.Struct{
											Fields: map[string]*pbtypes.Value{
												"sum": {
													Kind: &pbtypes.Value_NumberValue{
														NumberValue: 126, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIDs: up.CorrelationIDs,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x55, 0x55, 0x55, 0x55},
								SessionKeys: ttnpb.SessionKeys{
									SessionKeyID: []byte{0x55},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
										KEKLabel:     "test",
									},
								},
								LastAFCntDown: 2,
								StartedAt:     dev.Session.StartedAt,
							})
							a.So(dev.PendingSession, should.BeNil)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyID: []byte{0x55},
									FPort:        11,
									FCnt:         1,
									FRMPayload:   []byte{0x1, 0x1, 0x1, 0x1},
								},
								{
									SessionKeyID: []byte{0x55},
									FPort:        22,
									FCnt:         2,
									FRMPayload:   []byte{0x2, 0x2, 0x2, 0x2},
								},
							})
						},
					},
					{
						Name:          "UnregisteredDevice/JoinAccept",
						IDs:           unregisteredDeviceID,
						ExpectTimeout: true,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyID: []byte{0x55},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
										EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
										KEKLabel:     "test",
									},
								},
							},
						},
					},
					{
						Name:          "UnregisteredDevice/UplinkMessage",
						IDs:           unregisteredDeviceID,
						ExpectTimeout: true,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIdentifiers: withDevAddr(unregisteredDeviceID, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									SessionKeyID: []byte{0x55},
									FPort:        11,
									FCnt:         11,
									FRMPayload:   []byte{0xaa, 0x64, 0xb7, 0x7},
								},
							},
						},
					},
				} {
					tcok := t.Run(tc.Name, func(t *testing.T) {
						if tc.ResetQueue != nil {
							_, err := ns.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
								EndDeviceIdentifiers: tc.IDs,
								Downlinks:            tc.ResetQueue,
							})
							if err != nil {
								t.Fatalf("Unexpected error when resetting queue: %v", err)
							}
						}
						ns.upCh <- tc.Message
						select {
						case msg := <-chs.up:
							if tc.ExpectTimeout {
								t.Fatalf("Expected Timeout but got %v", msg)
							} else {
								if tc.AssertUp != nil {
									tc.AssertUp(t, msg)
								} else {
									t.Fatalf("Expected no upstream message but got %v", msg)
								}
							}
						case <-time.After(Timeout):
							if !tc.ExpectTimeout && tc.AssertUp != nil {
								t.Fatal("Expected upstream timeout")
							}
						}
						if tc.AssertDevice != nil {
							dev, err := deviceRegistry.Get(ctx, tc.Message.EndDeviceIdentifiers, []string{"session", "pending_session"})
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							queue, err := as.DownlinkQueueList(ctx, tc.IDs)
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							tc.AssertDevice(t, dev, queue)
						}
					})
					if !tcok {
						t.FailNow()
					}
				}
			})

			t.Run("Downstream", func(t *testing.T) {
				ns.reset()
				devsFlush()
				deviceRegistry.Set(ctx, registeredDevice.EndDeviceIdentifiers, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					dev := *registeredDevice
					dev.Session = &ttnpb.Session{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff},
						SessionKeys: ttnpb.SessionKeys{
							SessionKeyID: []byte{0x11},
							AppSKey: &ttnpb.KeyEnvelope{
								EncryptedKey: []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KEKLabel:     "test",
							},
						},
					}
					return &dev, []string{"ids", "version_ids", "session", "formatters"}, nil
				})
				t.Run("UnregisteredDevice/Push", func(t *testing.T) {
					a := assertions.New(t)
					chs.downPush <- &ttnpb.DownlinkQueueRequest{
						EndDeviceIdentifiers: unregisteredDeviceID,
						Downlinks: []*ttnpb.ApplicationDownlink{
							{
								FPort:      11,
								FRMPayload: []byte{0x1, 0x1, 0x1},
							},
						},
					}
					time.Sleep(Timeout)
					select {
					case err := <-chs.downErr:
						if !ptc.SkipCheckDownErr && a.So(err, should.NotBeNil) {
							a.So(errors.IsNotFound(err), should.BeTrue)
						}
					default:
						t.Fatal("Expected downlink error")
					}
					select {
					case up := <-chs.up:
						a.So(up.Up, should.HaveSameTypeAs, &ttnpb.ApplicationUp_DownlinkFailed{})
					default:
						t.Fatal("Expected upstream event")
					}
				})
				t.Run("RegisteredDevice/Push", func(t *testing.T) {
					a := assertions.New(t)
					for _, items := range [][]*ttnpb.ApplicationDownlink{
						{
							{
								FPort:      11,
								FRMPayload: []byte{0x1, 0x1, 0x1},
							},
							{
								FPort:      22,
								FRMPayload: []byte{0x2, 0x2, 0x2},
							},
						},
						{
							{
								FPort: 33,
								DecodedPayload: &pbtypes.Struct{
									Fields: map[string]*pbtypes.Value{
										"sum": {
											Kind: &pbtypes.Value_NumberValue{
												NumberValue: 6, // Payload formatter returns a byte slice with this many 1s.
											},
										},
									},
								},
							},
						},
					} {
						chs.downPush <- &ttnpb.DownlinkQueueRequest{
							EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
							Downlinks:            items,
						}
						time.Sleep(Timeout)
						select {
						case err := <-chs.downErr:
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
						default:
							t.Fatal("Expected downlink error")
						}
						for i := 0; i < len(items); i++ {
							select {
							case up := <-chs.up:
								a.So(up.Up, should.HaveSameTypeAs, &ttnpb.ApplicationUp_DownlinkQueued{})
							default:
								t.Fatalf("Expected upstream event")
							}
						}
					}
					res, err := as.DownlinkQueueList(ctx, registeredDevice.EndDeviceIdentifiers)
					if a.So(err, should.BeNil) && a.So(res, should.HaveLength, 3) {
						a.So(res, should.Resemble, []*ttnpb.ApplicationDownlink{
							{
								SessionKeyID:   []byte{0x11},
								FPort:          11,
								FCnt:           1,
								FRMPayload:     []byte{0x1, 0x1, 0x1},
								CorrelationIDs: res[0].CorrelationIDs,
							},
							{
								SessionKeyID:   []byte{0x11},
								FPort:          22,
								FCnt:           2,
								FRMPayload:     []byte{0x2, 0x2, 0x2},
								CorrelationIDs: res[1].CorrelationIDs,
							},
							{
								SessionKeyID:   []byte{0x11},
								FPort:          33,
								FCnt:           3,
								FRMPayload:     []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1},
								CorrelationIDs: res[2].CorrelationIDs,
							},
						})
					}
				})
				t.Run("RegisteredDevice/Replace", func(t *testing.T) {
					a := assertions.New(t)
					chs.downReplace <- &ttnpb.DownlinkQueueRequest{
						EndDeviceIdentifiers: registeredDevice.EndDeviceIdentifiers,
						Downlinks: []*ttnpb.ApplicationDownlink{
							{
								FPort:      11,
								FRMPayload: []byte{0x1, 0x1, 0x1},
							},
							{
								FPort:      22,
								FRMPayload: []byte{0x2, 0x2, 0x2},
							},
						},
					}
					time.Sleep(Timeout)
					select {
					case err := <-chs.downErr:
						if !a.So(err, should.BeNil) {
							t.FailNow()
						}
					default:
						t.Fatal("Expected downlink error")
					}
					for i := 0; i < 2; i++ {
						select {
						case up := <-chs.up:
							a.So(up.Up, should.HaveSameTypeAs, &ttnpb.ApplicationUp_DownlinkQueued{})
						default:
							t.Fatalf("Expected upstream event")
						}
					}
					res, err := as.DownlinkQueueList(ctx, registeredDevice.EndDeviceIdentifiers)
					if a.So(err, should.BeNil) && a.So(res, should.HaveLength, 2) {
						a.So(res, should.Resemble, []*ttnpb.ApplicationDownlink{
							{
								SessionKeyID:   []byte{0x11},
								FPort:          11,
								FCnt:           4,
								FRMPayload:     []byte{0x1, 0x1, 0x1},
								CorrelationIDs: res[0].CorrelationIDs,
							},
							{
								SessionKeyID:   []byte{0x11},
								FPort:          22,
								FCnt:           5,
								FRMPayload:     []byte{0x2, 0x2, 0x2},
								CorrelationIDs: res[1].CorrelationIDs,
							},
						})
					}
				})
			})

			cancel()
			wg.Wait()
		})
	}
}
