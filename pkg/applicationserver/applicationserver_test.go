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
	"io"
	"net/http"
	"net/http/httptest"
	"sync"
	"testing"
	"time"

	mqttnet "github.com/TheThingsIndustries/mystique/pkg/net"
	mqttserver "github.com/TheThingsIndustries/mystique/pkg/server"
	mqtt "github.com/eclipse/paho.mqtt.golang"
	nats_server "github.com/nats-io/nats-server/v2/server"
	nats_test_server "github.com/nats-io/nats-server/v2/test"
	nats_client "github.com/nats-io/nats.go"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver"
	distribredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/distribution/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages"
	asioapredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/packages/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub"
	iopubsubredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web"
	iowebredis "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/web/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/metadata"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/cluster"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/config"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/events"
	mockis "go.thethings.network/lorawan-stack/v3/pkg/identityserver/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/rpcmetadata"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
	"google.golang.org/protobuf/types/known/wrapperspb"
)

type connChannels struct {
	up          chan *ttnpb.ApplicationUp
	downPush    chan *ttnpb.DownlinkQueueRequest
	downReplace chan *ttnpb.DownlinkQueueRequest
	downErr     chan error
}

func TestApplicationServer(t *testing.T) {
	a, ctx := test.New(t)

	// This application will be added to the Entity Registry and to the link registry of the Application Server so that it
	// links automatically on start to the Network Server.
	registeredApplicationID := &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}
	registeredApplicationKey := "secret"
	registeredApplicationFormatter := ttnpb.PayloadFormatter_FORMATTER_CAYENNELPP
	registeredApplicationWebhookID := &ttnpb.ApplicationWebhookIdentifiers{
		ApplicationIds: registeredApplicationID,
		WebhookId:      "test",
	}
	registeredApplicationPubSubID := &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIds: registeredApplicationID,
		PubSubId:       "test",
	}

	// This device gets registered in the device registry of the Application Server.
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: registeredApplicationID,
			DeviceId:       "foo-device",
			JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		},
		VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
			BrandId:         "thethingsproducts",
			ModelId:         "thethingsnode",
			HardwareVersion: "1.0",
			FirmwareVersion: "1.1",
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter: ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
			UpFormatterParameter: `function decodeUplink(input) {
				var sum = 0;
				for (i = 0; i < input.bytes.length; i++) {
					sum += input.bytes[i];
				}
				return {
					data: {
						sum: sum
					}
				};
			}
			`,
			DownFormatter: ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
			DownFormatterParameter: `function encodeDownlink(input) {
				var bytes = [];
				for (i = 0; i < input.data.sum; i++) {
					bytes[i] = 1;
				}
				return {
					bytes: bytes,
					fPort: input.fPort
				};
			}

			function decodeDownlink(input) {
				var sum = 0;
				for (i = 0; i < input.bytes.length; i++) {
					sum += input.bytes[i];
				}
				return {
					data: {
						sum: sum
					}
				}
			}
			`,
		},
	}

	// This device does not get registered in the device registry of the Application Server and will be created on join
	// and on uplink.
	unregisteredDeviceID := &ttnpb.EndDeviceIdentifiers{
		ApplicationIds: registeredApplicationID,
		DeviceId:       "bar-device",
		JoinEui:        types.EUI64{0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		DevEui:         types.EUI64{0x24, 0x24, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
	}

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	js, jsAddr := startMockJS(ctx)
	nsConnChan := make(chan *mockNSASConn)
	ns, nsAddr := startMockNS(ctx, nsConnChan)

	// Register the application in the Entity Registry.
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	// Register some sessions in the Join Server. Sometimes the keys are sent by the Network Server as part of the
	// join-accept, and sometimes they are not sent by the Network Server so the Application Server gets them from the
	// Join Server.
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x11}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.
		EncryptedKey: []byte{0xa8, 0x11, 0x8f, 0x80, 0x2e, 0xbf, 0x8, 0xdc, 0x62, 0x37, 0xc3, 0x4, 0x63, 0xa2, 0xfa, 0xcb, 0xf8, 0x87, 0xaa, 0x31, 0x90, 0x23, 0x85, 0xc1},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x22}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
		EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x33}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
		EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x44}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}
		EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x55}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
		EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
		KekLabel:     "test",
	})

	devsRedisClient, devsFlush := test.NewRedis(ctx, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	linksRedisClient, linksFlush := test.NewRedis(ctx, "applicationserver_test", "links")
	defer linksFlush()
	defer linksRedisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: linksRedisClient, LockTTL: test.Delay << 10}
	if err := linkRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err := linkRegistry.Set(ctx, registeredApplicationID, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{
			DefaultFormatters: &ttnpb.MessagePayloadFormatters{
				UpFormatter:   registeredApplicationFormatter,
				DownFormatter: registeredApplicationFormatter,
			},
		}, []string{"default_formatters"}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set link in registry: %s", err)
	}

	webhooksRedisClient, webhooksFlush := test.NewRedis(ctx, "applicationserver_test", "webhooks")
	defer webhooksFlush()
	defer webhooksRedisClient.Close()
	webhookRegistry := iowebredis.WebhookRegistry{Redis: webhooksRedisClient, LockTTL: test.Delay << 10}
	if err := webhookRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	pubsubRedisClient, pubsubFlush := test.NewRedis(ctx, "applicationserver_test", "pubsub")
	defer pubsubFlush()
	defer pubsubRedisClient.Close()
	pubsubRegistry := iopubsubredis.PubSubRegistry{Redis: pubsubRedisClient, LockTTL: test.Delay << 10}
	if err := pubsubRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	distribRedisClient, distribFlush := test.NewRedis(ctx, "applicationserver_test", "traffic")
	defer distribFlush()
	defer distribRedisClient.Close()
	distribPubSub := distribredis.PubSub{Redis: distribRedisClient}

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

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9184",
				AllowInsecureForCredentials: true,
			},
			HTTP: config.HTTP{
				Listen: ":8099",
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
				JoinServer:     jsAddr,
				NetworkServer:  nsAddr,
			},
			KeyVault: config.KeyVault{
				Provider: "static",
				Static: map[string][]byte{
					"test": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			},
		},
	})
	config := &applicationserver.Config{
		Devices: deviceRegistry,
		Links:   linkRegistry,
		MQTT: config.MQTT{
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
		Distribution: applicationserver.DistributionConfig{
			Global: applicationserver.GlobalDistributorConfig{
				PubSub: distribPubSub,
			},
		},
		EndDeviceMetadataStorage: applicationserver.EndDeviceMetadataStorageConfig{
			Location: applicationserver.EndDeviceLocationStorageConfig{
				Registry: metadata.NewNoopEndDeviceLocationRegistry(),
			},
		},
		Downlinks: applicationserver.DownlinksConfig{
			ConfirmationConfig: applicationserver.ConfirmationConfig{
				DefaultRetryAttempts: 3,
				MaxRetryAttempts:     10,
			},
		},
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_APPLICATION_SERVER)

	componenttest.StartComponent(t, c)
	defer c.Close()

	select {
	case <-ctx.Done():
		return
	case nsConnChan <- &mockNSASConn{
		cc:   as.LoopbackConn(),
		auth: as.WithClusterAuth(),
	}:
	}

	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_JOIN_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	for _, ptc := range []struct {
		Protocol         string
		ValidAuth        func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool
		Connect          func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error
		SkipCheckDownErr bool
	}{
		{
			Protocol: "grpc",
			ValidAuth: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool {
				return unique.ID(ctx, ids) == unique.ID(ctx, registeredApplicationID) && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewAppAsClient(as.LoopbackConn())
				stream, err := client.Subscribe(ctx, ids, creds)
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
			ValidAuth: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool {
				return unique.ID(ctx, ids) == unique.ID(ctx, registeredApplicationID) && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
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
						token := client.Publish(fmt.Sprintf(topicFmt, unique.ID(ctx, req.EndDeviceIds.ApplicationIds), req.EndDeviceIds.DeviceId), 1, false, buf)
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
			ValidAuth: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool {
				return unique.ID(ctx, ids) == unique.ID(ctx, registeredApplicationID) && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				evCh := make(chan events.Event, EventsBufferSize)
				defer test.RedirectEvents(evCh)()
				// Configure pubsub.
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewApplicationPubSubRegistryClient(as.LoopbackConn())
				req := &ttnpb.SetApplicationPubSubRequest{
					Pubsub: &ttnpb.ApplicationPubSub{
						Ids: registeredApplicationPubSubID,
						Provider: &ttnpb.ApplicationPubSub_Nats{
							Nats: &ttnpb.ApplicationPubSub_NATSProvider{
								ServerUrl: "nats://localhost:4124",
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
						UplinkNormalized: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.uplink.normalized",
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
						DownlinkQueueInvalidated: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.downlink.invalidated",
						},
						LocationSolved: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.location.solved",
						},
						ServiceData: &ttnpb.ApplicationPubSub_Message{
							Topic: "up.service.data",
						},
					},
					FieldMask: ttnpb.FieldMask(
						"base_topic",
						"downlink_ack",
						"downlink_failed",
						"downlink_nack",
						"downlink_push",
						"downlink_queue_invalidated",
						"downlink_queued",
						"downlink_replace",
						"downlink_sent",
						"format",
						"join_accept",
						"location_solved",
						"provider",
						"service_data",
						"uplink_message",
						"uplink_normalized",
					),
				}
				if _, err := client.Set(ctx, req, creds); err != nil {
					return err
				}
				if !test.WaitEvent(ctx, evCh, "as.pubsub.start") {
					t.Fatal("Integration start timeout")
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
			SkipCheckDownErr: true, // There is no direct error response in pub/sub.
		},
		{
			Protocol: "pubsub/mqtt",
			ValidAuth: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool {
				return unique.ID(ctx, ids) == unique.ID(ctx, registeredApplicationID) && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				evCh := make(chan events.Event, EventsBufferSize)
				defer test.RedirectEvents(evCh)()
				// Configure pubsub.
				creds := grpc.PerRPCCredentials(rpcmetadata.MD{
					AuthType:      "Bearer",
					AuthValue:     key,
					AllowInsecure: true,
				})
				client := ttnpb.NewApplicationPubSubRegistryClient(as.LoopbackConn())
				req := &ttnpb.SetApplicationPubSubRequest{
					Pubsub: &ttnpb.ApplicationPubSub{
						Ids: registeredApplicationPubSubID,
						Provider: &ttnpb.ApplicationPubSub_Mqtt{
							Mqtt: &ttnpb.ApplicationPubSub_MQTTProvider{
								ServerUrl:    fmt.Sprintf("tcp://%v", mqttLis.Addr()),
								PublishQos:   ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
								SubscribeQos: ttnpb.ApplicationPubSub_MQTTProvider_AT_LEAST_ONCE,
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
						UplinkNormalized: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/uplink/normalized",
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
						DownlinkQueueInvalidated: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/downlink/invalidated",
						},
						LocationSolved: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/location/solved",
						},
						ServiceData: &ttnpb.ApplicationPubSub_Message{
							Topic: "up/service/data",
						},
					},
					FieldMask: ttnpb.FieldMask(
						"base_topic",
						"downlink_ack",
						"downlink_failed",
						"downlink_nack",
						"downlink_push",
						"downlink_queue_invalidated",
						"downlink_queued",
						"downlink_replace",
						"downlink_sent",
						"format",
						"join_accept",
						"location_solved",
						"provider",
						"service_data",
						"uplink_message",
						"uplink_normalized",
					),
				}
				if _, err := client.Set(ctx, req, creds); err != nil {
					return err
				}
				if !test.WaitEvent(ctx, evCh, "as.pubsub.start") {
					t.Fatal("Integration start timeout")
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
			SkipCheckDownErr: true, // There is no direct error response in pub/sub.
		},
		{
			Protocol: "webhooks",
			ValidAuth: func(ctx context.Context, ids *ttnpb.ApplicationIdentifiers, key string) bool {
				return unique.ID(ctx, ids) == unique.ID(ctx, registeredApplicationID) && key == registeredApplicationKey
			},
			Connect: func(ctx context.Context, t *testing.T, ids *ttnpb.ApplicationIdentifiers, key string, chs *connChannels) error {
				// Start web server to read upstream.
				webhookTarget := httptest.NewServer(http.HandlerFunc(func(res http.ResponseWriter, req *http.Request) {
					buf, err := io.ReadAll(req.Body)
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
					Webhook: &ttnpb.ApplicationWebhook{
						Ids:              registeredApplicationWebhookID,
						BaseUrl:          webhookTarget.URL,
						Format:           "json",
						UplinkMessage:    &ttnpb.ApplicationWebhook_Message{Path: ""},
						UplinkNormalized: &ttnpb.ApplicationWebhook_Message{Path: ""},
						JoinAccept:       &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkAck:      &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkNack:     &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkQueued:   &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkSent:     &ttnpb.ApplicationWebhook_Message{Path: ""},
						DownlinkFailed:   &ttnpb.ApplicationWebhook_Message{Path: ""},
						LocationSolved:   &ttnpb.ApplicationWebhook_Message{Path: ""},
						ServiceData:      &ttnpb.ApplicationWebhook_Message{Path: ""},
					},
					FieldMask: ttnpb.FieldMask(
						"base_url",
						"downlink_ack",
						"downlink_failed",
						"downlink_nack",
						"downlink_queued",
						"downlink_sent",
						"format",
						"join_accept",
						"location_solved",
						"service_data",
						"uplink_message",
						"uplink_normalized",
					),
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
							data.EndDeviceIds.ApplicationIds.ApplicationId, registeredApplicationWebhookID.WebhookId, data.EndDeviceIds.DeviceId, action,
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
				ID   *ttnpb.ApplicationIdentifiers
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
					ID:   &ttnpb.ApplicationIdentifiers{ApplicationId: "invalid-application"},
					Key:  "invalid-key",
				},
			} {
				t.Run(ctc.Name, func(t *testing.T) {
					ctx, cancel := context.WithDeadline(ctx, time.Now().Add(8*Timeout))
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
					t.Errorf("Expected context canceled, but have %v", err)
				}
			}()
			// Wait for connection to establish.
			time.Sleep(2 * Timeout)

			now := time.Now().UTC()

			t.Run("Upstream", func(t *testing.T) {
				ns.reset()
				devsFlush()
				deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					return registeredDevice, []string{"ids", "version_ids", "formatters"}, nil
				})

				for _, tc := range []struct {
					Name          string
					IDs           *ttnpb.EndDeviceIdentifiers
					ResetQueue    []*ttnpb.ApplicationDownlink
					Message       *ttnpb.ApplicationUp
					ExpectTimeout bool
					AssertUp      func(t *testing.T, up *ttnpb.ApplicationUp)
					AssertDevice  func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink)
				}{
					{
						Name: "RegisteredDevice/JoinAccept",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x11, 0x11, 0x11, 0x11}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x11},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x11, 0x11, 0x11, 0x11}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyId: []byte{0x11},
										ReceivedAt:   up.GetJoinAccept().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.BeNil)
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x11, 0x11, 0x11, 0x11}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x11},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0xa8, 0x11, 0x8f, 0x80, 0x2e, 0xbf, 0x8, 0xdc, 0x62, 0x37, 0xc3, 0x4, 0x63, 0xa2, 0xfa, 0xcb, 0xf8, 0x87, 0xaa, 0x31, 0x90, 0x23, 0x85, 0xc1},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 0,
							})
							a.So(queue, should.BeEmpty)
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KekLabel:     "test",
									},
									ReceivedAt: timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyId: []byte{0x22},
										ReceivedAt:   up.GetJoinAccept().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.BeNil)
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x22, 0x22, 0x22, 0x22}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 0,
							})
							a.So(queue, should.BeEmpty)
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/PendingSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x22},
									FPort:        22,
									FCnt:         22,
									FrmPayload:   []byte{0x01},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
										Settings: &ttnpb.TxSettings{
											DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
											Frequency: 868000000,
										},
										SessionKeyId: []byte{0x22},
										FPort:        22,
										FCnt:         22,
										FrmPayload:   []byte{0xc1},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 193, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
										VersionIds: registeredDevice.VersionIds,
										ReceivedAt: up.GetUplinkMessage().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey/WithQueue",
						IDs:  registeredDevice.Ids,
						ResetQueue: []*ttnpb.ApplicationDownlink{
							{
								SessionKeyId: []byte{0x22},
								FPort:        11,
								FCnt:         11,
								FrmPayload:   []byte{0x69, 0x65, 0x9f, 0x8f},
							},
							{
								SessionKeyId: []byte{0x22},
								FPort:        22,
								FCnt:         22,
								FrmPayload:   []byte{0xb, 0x8f, 0x94, 0xe6},
							},
						},
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KekLabel:     "test",
									},
									InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x22},
											FPort:        11,
											FCnt:         11,
											FrmPayload:   []byte{0x69, 0x65, 0x9f, 0x8f},
										},
										{
											SessionKeyId: []byte{0x22},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0xb, 0x8f, 0x94, 0xe6},
										},
									},
									ReceivedAt: timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyId: []byte{0x33},
										ReceivedAt:   up.GetJoinAccept().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x22, 0x22, 0x22, 0x22}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 0,
							})
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x33, 0x33, 0x33, 0x33}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 2,
							})
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x22},
									FPort:        11,
									FCnt:         11,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4,
												},
											},
										},
									},
								},
								{
									SessionKeyId: []byte{0x22},
									FPort:        22,
									FCnt:         22,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8,
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/CurrentSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FrmPayload:   []byte{0xca, 0xa9, 0x42},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
										Settings: &ttnpb.TxSettings{
											DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
											Frequency: 868000000,
										},
										SessionKeyId: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x01, 0x02, 0x03},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 6, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
										VersionIds: registeredDevice.VersionIds,
										ReceivedAt: up.GetUplinkMessage().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/QueueInvalidated",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x33},
											FPort:        42,
											FCnt:         42,
											FrmPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
										},
									},
									LastFCntDown: 42,
									SessionKeyId: []byte{0x33},
								},
							},
						},
						ExpectTimeout: true, // Payload encryption is carried out by AS; this message is not sent upstream.
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Sent",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkSent{
								DownlinkSent: &ttnpb.ApplicationDownlink{
									SessionKeyId: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FrmPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 370,
												},
											},
										},
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkSent{
									DownlinkSent: &ttnpb.ApplicationDownlink{
										SessionKeyId: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Failed",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkFailed{
								DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
									Downlink: &ttnpb.ApplicationDownlink{
										SessionKeyId: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
									},
									Error: &ttnpb.ErrorDetails{
										Name: "test",
									},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkFailed{
									DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
										Downlink: &ttnpb.ApplicationDownlink{
											SessionKeyId: []byte{0x33},
											FPort:        42,
											FCnt:         42,
											FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
											DecodedPayload: &structpb.Struct{
												Fields: map[string]*structpb.Value{
													"sum": {
														Kind: &structpb.Value_NumberValue{
															NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
														},
													},
												},
											},
										},
										Error: &ttnpb.ErrorDetails{
											Name: "test",
										},
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Ack",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkAck{
								DownlinkAck: &ttnpb.ApplicationDownlink{
									SessionKeyId: []byte{0x33},
									FPort:        42,
									FCnt:         42,
									FrmPayload:   []byte{0x50, 0xd, 0x40, 0xd5},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkAck{
									DownlinkAck: &ttnpb.ApplicationDownlink{
										SessionKeyId: []byte{0x33},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkMessage/Nack",
						IDs:  registeredDevice.Ids,
						ResetQueue: []*ttnpb.ApplicationDownlink{ // Pop the first item; it will be appended because of the nack.
							{
								SessionKeyId: []byte{0x33},
								FPort:        22,
								FCnt:         2,
								FrmPayload:   []byte{0x92, 0xfe, 0x93, 0xf5},
							},
						},
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
							Up: &ttnpb.ApplicationUp_DownlinkNack{
								DownlinkNack: &ttnpb.ApplicationDownlink{
									SessionKeyId: []byte{0x33},
									FPort:        11,
									FCnt:         1,
									FrmPayload:   []byte{0x5f, 0x38, 0x7c, 0xb0},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x33, 0x33, 0x33, 0x33}),
								Up: &ttnpb.ApplicationUp_DownlinkNack{
									DownlinkNack: &ttnpb.ApplicationDownlink{
										SessionKeyId: []byte{0x33},
										FPort:        11,
										FCnt:         1,
										FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
										ConfirmedRetry: &ttnpb.ApplicationDownlink_ConfirmedRetry{
											Attempt: 1,
										},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x33},
									FPort:        22,
									FCnt:         2,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
								{ // The nacked item is appended at the end of the queue.
									SessionKeyId: []byte{0x33},
									FPort:        11,
									FCnt:         44,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									ConfirmedRetry: &ttnpb.ApplicationDownlink_ConfirmedRetry{
										Attempt: 1,
									},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/JoinAccept/WithAppSKey/WithQueue/WithPendingSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KekLabel:     "test",
									},
									PendingSession: true,
									InvalidatedDownlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x33},
											FPort:        11,
											FCnt:         2,
											FrmPayload:   []byte{0x91, 0xfd, 0x90, 0xf6},
										},
										{
											SessionKeyId: []byte{0x33},
											FPort:        22,
											FCnt:         3,
											FrmPayload:   []byte{0x2f, 0x3f, 0x31, 0x2c},
										},
									},
									ReceivedAt: timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
								Up: &ttnpb.ApplicationUp_JoinAccept{
									JoinAccept: &ttnpb.ApplicationJoinAccept{
										SessionKeyId:   []byte{0x44},
										PendingSession: true,
										ReceivedAt:     up.GetJoinAccept().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x33, 0x33, 0x33, 0x33}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x33},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 44,
							})
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x44, 0x44, 0x44, 0x44}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 2,
							})
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x33},
									FPort:        22,
									FCnt:         2,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
								{
									SessionKeyId: []byte{0x33},
									FPort:        11,
									FCnt:         44,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									ConfirmedRetry: &ttnpb.ApplicationDownlink_ConfirmedRetry{
										Attempt: 1,
									},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/PendingSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x44},
									FPort:        24,
									FCnt:         24,
									FrmPayload:   []byte{0x14, 0x4e, 0x3c},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
										Settings: &ttnpb.TxSettings{
											DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
											Frequency: 868000000,
										},
										SessionKeyId: []byte{0x44},
										FPort:        24,
										FCnt:         24,
										FrmPayload:   []byte{0x64, 0x64, 0x64},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 300, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
										VersionIds: registeredDevice.VersionIds,
										ReceivedAt: up.GetUplinkMessage().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x44, 0x44, 0x44, 0x44}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x44},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 2,
							})
							a.So(dev.PendingSession, should.BeNil)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x44},
									FPort:        11,
									FCnt:         1,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
								{
									SessionKeyId: []byte{0x44},
									FPort:        22,
									FCnt:         2,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name:       "RegisteredDevice/DownlinkQueueInvalidated/KnownSession",
						IDs:        registeredDevice.Ids,
						ResetQueue: make([]*ttnpb.ApplicationDownlink, 0),
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x44},
											FPort:        11,
											FCnt:         11,
											FrmPayload:   []byte{0x65, 0x98, 0xa7, 0xfc},
										},
										{
											SessionKeyId: []byte{0x44},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0x1b, 0x4b, 0x97, 0xb9},
										},
									},
									LastFCntDown: 42,
									SessionKeyId: []byte{0x44},
								},
							},
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session.LastAFCntDown, should.Equal, 44)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x44},
									FPort:        11,
									FCnt:         43,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
								{
									SessionKeyId: []byte{0x44},
									FPort:        22,
									FCnt:         44,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/DownlinkQueueInvalidated/KnownSession/NoDownlinks",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks:    nil,
									LastFCntDown: 46,
									SessionKeyId: []byte{0x44},
								},
							},
						},
						ResetQueue: make([]*ttnpb.ApplicationDownlink, 0),
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session.LastAFCntDown, should.Equal, 46)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink(nil))
						},
					},
					{
						Name: "RegisteredDevice/DownlinkQueueInvalidated/UnknownSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x44, 0x44, 0x44, 0x44}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x44},
											FPort:        11,
											FCnt:         11,
											FrmPayload:   []byte{0x65, 0x98, 0xa7, 0xfc},
										},
										{
											SessionKeyId: []byte{0x11, 0x22, 0x33, 0x44},
											FPort:        12,
											FCnt:         12,
											FrmPayload:   []byte{0xff, 0xff, 0xff, 0xff},
										},
										{
											SessionKeyId: []byte{0x44},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0x1b, 0x4b, 0x97, 0xb9},
										},
									},
									LastFCntDown: 84,
									SessionKeyId: []byte{0x44},
								},
							},
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session.LastAFCntDown, should.Equal, 86)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x44},
									FPort:        11,
									FCnt:         85,
									FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 4, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
								{
									SessionKeyId: []byte{0x44},
									FPort:        22,
									FCnt:         86,
									FrmPayload:   []byte{0x2, 0x2, 0x2, 0x2},
									DecodedPayload: &structpb.Struct{
										Fields: map[string]*structpb.Value{
											"sum": {
												Kind: &structpb.Value_NumberValue{
													NumberValue: 8, // Payload formatter sums the bytes in FRMPayload.
												},
											},
										},
									},
								},
							})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/KnownSession",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x55},
									FPort:        42,
									FCnt:         42,
									FrmPayload:   []byte{0xd1, 0x43, 0x6a},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
										Settings: &ttnpb.TxSettings{
											DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
											Frequency: 868000000,
										},
										SessionKeyId: []byte{0x55},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x2a, 0x2a, 0x2a},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 126, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
										VersionIds: registeredDevice.VersionIds,
										ReceivedAt: up.GetUplinkMessage().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x55, 0x55, 0x55, 0x55}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x55},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 0,
							})
							a.So(dev.PendingSession, should.BeNil)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{})
						},
					},
					{
						Name: "RegisteredDevice/UplinkMessage/KnownSession/",
						IDs:  registeredDevice.Ids,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x55},
									FPort:        42,
									FCnt:         42,
									FrmPayload:   []byte{0xd1, 0x43, 0x6a},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.Resemble, &ttnpb.ApplicationUp{
								EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
								Up: &ttnpb.ApplicationUp_UplinkMessage{
									UplinkMessage: &ttnpb.ApplicationUplink{
										RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
										Settings: &ttnpb.TxSettings{
											DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
											Frequency: 868000000,
										},
										SessionKeyId: []byte{0x55},
										FPort:        42,
										FCnt:         42,
										FrmPayload:   []byte{0x2a, 0x2a, 0x2a},
										DecodedPayload: &structpb.Struct{
											Fields: map[string]*structpb.Value{
												"sum": {
													Kind: &structpb.Value_NumberValue{
														NumberValue: 126, // Payload formatter sums the bytes in FRMPayload.
													},
												},
											},
										},
										VersionIds: registeredDevice.VersionIds,
										ReceivedAt: up.GetUplinkMessage().ReceivedAt,
									},
								},
								CorrelationIds: up.CorrelationIds,
								ReceivedAt:     up.ReceivedAt,
							})
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice, queue []*ttnpb.ApplicationDownlink) {
							a := assertions.New(t)
							a.So(dev.Session, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x55, 0x55, 0x55, 0x55}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x55},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
										KekLabel:     "test",
									},
								},
								LastAFCntDown: 0,
							})
							a.So(dev.PendingSession, should.BeNil)
							a.So(queue, should.Resemble, []*ttnpb.ApplicationDownlink{})
						},
					},
					{
						Name:          "UnregisteredDevice/JoinAccept",
						IDs:           unregisteredDeviceID,
						ExpectTimeout: true,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(unregisteredDeviceID, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x55},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
										EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
										KekLabel:     "test",
									},
									ReceivedAt: timestamppb.New(now),
								},
							},
						},
					},
					{
						Name:          "UnregisteredDevice/UplinkMessage",
						IDs:           unregisteredDeviceID,
						ExpectTimeout: true,
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(unregisteredDeviceID, types.DevAddr{0x55, 0x55, 0x55, 0x55}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x55},
									FPort:        11,
									FCnt:         11,
									FrmPayload:   []byte{0xaa, 0x64, 0xb7, 0x7},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
					},
				} {
					tcok := t.Run(tc.Name, func(t *testing.T) {
						if tc.ResetQueue != nil {
							_, err := ns.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
								EndDeviceIds: tc.IDs,
								Downlinks:    tc.ResetQueue,
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
							dev, err := deviceRegistry.Get(ctx, tc.Message.EndDeviceIds, []string{"session", "pending_session"})
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
				deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					dev := registeredDevice
					dev.Session = &ttnpb.Session{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff}.Bytes(),
						Keys: &ttnpb.SessionKeys{
							SessionKeyId: []byte{0x11},
							AppSKey: &ttnpb.KeyEnvelope{
								EncryptedKey: []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KekLabel:     "test",
							},
						},
					}
					return dev, []string{"ids", "version_ids", "session", "formatters"}, nil
				})
				t.Run("UnregisteredDevice/Push", func(t *testing.T) {
					a := assertions.New(t)
					chs.downPush <- &ttnpb.DownlinkQueueRequest{
						EndDeviceIds: unregisteredDeviceID,
						Downlinks: []*ttnpb.ApplicationDownlink{
							{
								FPort:      11,
								FrmPayload: []byte{0x1, 0x1, 0x1},
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
				})
				t.Run("RegisteredDevice/Push", func(t *testing.T) {
					a := assertions.New(t)
					for _, items := range [][]*ttnpb.ApplicationDownlink{
						{
							{
								FPort:      11,
								FrmPayload: []byte{0x1, 0x1, 0x1},
							},
							{
								FPort:      22,
								FrmPayload: []byte{0x2, 0x2, 0x2},
							},
						},
						{
							{
								FPort: 33,
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 6, // Payload formatter returns a byte slice with this many 1s.
											},
										},
									},
								},
							},
						},
					} {
						chs.downPush <- &ttnpb.DownlinkQueueRequest{
							EndDeviceIds: registeredDevice.Ids,
							Downlinks:    items,
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
					res, err := as.DownlinkQueueList(ctx, registeredDevice.Ids)
					if a.So(err, should.BeNil) && a.So(res, should.HaveLength, 3) {
						a.So(res, should.Resemble, []*ttnpb.ApplicationDownlink{
							{
								SessionKeyId: []byte{0x11},
								FPort:        11,
								FCnt:         1,
								FrmPayload:   []byte{0x1, 0x1, 0x1},
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 3,
											},
										},
									},
								},
								CorrelationIds: res[0].CorrelationIds,
							},
							{
								SessionKeyId: []byte{0x11},
								FPort:        22,
								FCnt:         2,
								FrmPayload:   []byte{0x2, 0x2, 0x2},
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 6,
											},
										},
									},
								},
								CorrelationIds: res[1].CorrelationIds,
							},
							{
								SessionKeyId: []byte{0x11},
								FPort:        33,
								FCnt:         3,
								FrmPayload:   []byte{0x1, 0x1, 0x1, 0x1, 0x1, 0x1},
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 6,
											},
										},
									},
								},
								CorrelationIds: res[2].CorrelationIds,
							},
						})
					}
				})
				t.Run("RegisteredDevice/Replace", func(t *testing.T) {
					a := assertions.New(t)
					chs.downReplace <- &ttnpb.DownlinkQueueRequest{
						EndDeviceIds: registeredDevice.Ids,
						Downlinks: []*ttnpb.ApplicationDownlink{
							{
								FPort:      11,
								FrmPayload: []byte{0x1, 0x1, 0x1},
							},
							{
								FPort:      22,
								FrmPayload: []byte{0x2, 0x2, 0x2},
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
					res, err := as.DownlinkQueueList(ctx, registeredDevice.Ids)
					if a.So(err, should.BeNil) && a.So(res, should.HaveLength, 2) {
						a.So(res, should.Resemble, []*ttnpb.ApplicationDownlink{
							{
								SessionKeyId: []byte{0x11},
								FPort:        11,
								FCnt:         4,
								FrmPayload:   []byte{0x1, 0x1, 0x1},
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 3,
											},
										},
									},
								},
								CorrelationIds: res[0].CorrelationIds,
							},
							{
								SessionKeyId: []byte{0x11},
								FPort:        22,
								FCnt:         5,
								FrmPayload:   []byte{0x2, 0x2, 0x2},
								DecodedPayload: &structpb.Struct{
									Fields: map[string]*structpb.Value{
										"sum": {
											Kind: &structpb.Value_NumberValue{
												NumberValue: 6,
											},
										},
									},
								},
								CorrelationIds: res[1].CorrelationIds,
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

func TestSkipPayloadCrypto(t *testing.T) {
	a, ctx := test.New(t)

	// This application will be added to the Entity Registry and to the link registry of the Application Server so that it
	// links automatically on start to the Network Server.
	registeredApplicationID := &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}
	registeredApplicationKey := "secret"

	// This device gets registered in the device registry of the Application Server.
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: registeredApplicationID,
			DeviceId:       "foo-device",
			JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		},
	}

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	js, jsAddr := startMockJS(ctx)
	nsConnChan := make(chan *mockNSASConn)
	ns, nsAddr := startMockNS(ctx, nsConnChan)

	// Register the application in the Entity Registry.
	is.ApplicationRegistry().Add(ctx, registeredApplicationID, registeredApplicationKey, testRights...)

	// Register some sessions in the Join Server. Sometimes the keys are sent by the Network Server as part of the
	// join-accept, and sometimes they are not sent by the Network Server so the Application Server gets them from the
	// Join Server.
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x11}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.
		EncryptedKey: []byte{0xa8, 0x11, 0x8f, 0x80, 0x2e, 0xbf, 0x8, 0xdc, 0x62, 0x37, 0xc3, 0x4, 0x63, 0xa2, 0xfa, 0xcb, 0xf8, 0x87, 0xaa, 0x31, 0x90, 0x23, 0x85, 0xc1},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x22}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
		EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x33}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33, 0x33}
		EncryptedKey: []byte{0x5, 0x81, 0xe1, 0x15, 0x8a, 0xc3, 0x13, 0x68, 0x5e, 0x8d, 0x15, 0xc0, 0x11, 0x92, 0x14, 0x49, 0x9f, 0xa0, 0xc6, 0xf1, 0xdb, 0x95, 0xff, 0xbd},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x44}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44, 0x44}
		EncryptedKey: []byte{0x30, 0xcf, 0x47, 0x91, 0x11, 0x64, 0x53, 0x3f, 0xc3, 0xd5, 0xd8, 0x56, 0x5b, 0x71, 0xcb, 0xe7, 0x6d, 0x14, 0x2b, 0x2c, 0xf2, 0xc2, 0xd7, 0x7b},
		KekLabel:     "test",
	})
	js.add(ctx, types.MustEUI64(registeredDevice.Ids.DevEui).OrZero(), []byte{0x55}, &ttnpb.KeyEnvelope{
		// AppSKey is []byte{0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55, 0x55}
		EncryptedKey: []byte{0x56, 0x15, 0xaa, 0x22, 0xb7, 0x5f, 0xc, 0x24, 0x79, 0x6, 0x84, 0x68, 0x89, 0x0, 0xa6, 0x16, 0x4a, 0x9c, 0xef, 0xdb, 0xbf, 0x61, 0x6f, 0x0},
		KekLabel:     "test",
	})

	devsRedisClient, devsFlush := test.NewRedis(ctx, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	linksRedisClient, linksFlush := test.NewRedis(ctx, "applicationserver_test", "links")
	defer linksFlush()
	defer linksRedisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: linksRedisClient, LockTTL: test.Delay << 10}
	if err := linkRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err := linkRegistry.Set(ctx, registeredApplicationID, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{
			SkipPayloadCrypto: &wrapperspb.BoolValue{Value: true},
		}, []string{"skip_payload_crypto"}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set link in registry: %s", err)
	}

	distribRedisClient, distribFlush := test.NewRedis(ctx, "applicationserver_test", "traffic")
	defer distribFlush()
	defer distribRedisClient.Close()
	distribPubSub := distribredis.PubSub{Redis: distribRedisClient}

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":0",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
				JoinServer:     jsAddr,
				NetworkServer:  nsAddr,
			},
			KeyVault: config.KeyVault{
				Provider: "static",
				Static: map[string][]byte{
					"known": {0x00, 0x01, 0x02, 0x03, 0x04, 0x05, 0x06, 0x07, 0x08, 0x09, 0x0A, 0x0B, 0x0C, 0x0D, 0x0E, 0x0F},
				},
			},
		},
	})
	config := &applicationserver.Config{
		Devices: deviceRegistry,
		Links:   linkRegistry,
		Distribution: applicationserver.DistributionConfig{
			Global: applicationserver.GlobalDistributorConfig{
				PubSub: distribPubSub,
			},
		},
		EndDeviceMetadataStorage: applicationserver.EndDeviceMetadataStorageConfig{
			Location: applicationserver.EndDeviceLocationStorageConfig{
				Registry: metadata.NewNoopEndDeviceLocationRegistry(),
			},
		},
		Downlinks: applicationserver.DownlinksConfig{
			ConfirmationConfig: applicationserver.ConfirmationConfig{
				DefaultRetryAttempts: 3,
				MaxRetryAttempts:     10,
			},
		},
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_APPLICATION_SERVER)

	componenttest.StartComponent(t, c)
	defer c.Close()

	select {
	case <-ctx.Done():
		return
	case nsConnChan <- &mockNSASConn{
		cc:   as.LoopbackConn(),
		auth: as.WithClusterAuth(),
	}:
	}

	mustHavePeer(ctx, c, ttnpb.ClusterRole_NETWORK_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_JOIN_SERVER)
	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	chs := &connChannels{
		up:          make(chan *ttnpb.ApplicationUp, 1),
		downPush:    make(chan *ttnpb.DownlinkQueueRequest),
		downReplace: make(chan *ttnpb.DownlinkQueueRequest),
		downErr:     make(chan error, 1),
	}
	creds := grpc.PerRPCCredentials(rpcmetadata.MD{
		AuthType:      "Bearer",
		AuthValue:     registeredApplicationKey,
		AllowInsecure: true,
	})
	client := ttnpb.NewAppAsClient(as.LoopbackConn())
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	stream, err := client.Subscribe(ctx, registeredApplicationID, creds)
	if err != nil {
		t.Fatalf("Failed to subscribe: %v", err)
	}
	// Wait for connection to establish.
	time.Sleep(2 * Timeout)
	// Read upstream.
	go func() {
		for {
			msg, err := stream.Recv()
			if err != nil {
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

	for _, override := range []*wrapperspb.BoolValue{
		nil,
		{Value: true},
		{Value: false},
	} {
		hasOverride := map[bool]string{true: "Overrides", false: "NoOverride"}[override != nil]
		kekLabel := map[bool]string{true: "unknown", false: "known"}[override.GetValue()]

		t.Run(fmt.Sprintf("%v/%v", hasOverride, override.GetValue()), func(t *testing.T) {
			t.Run("Uplink", func(t *testing.T) {
				ns.reset()
				devsFlush()
				deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					dev := ttnpb.Clone(registeredDevice)
					dev.SkipPayloadCryptoOverride = override
					return dev, []string{
						"ids",
						"formatters",
						"skip_payload_crypto_override",
					}, nil
				})

				now := time.Now().UTC()

				for _, step := range []struct {
					Name         string
					Message      *ttnpb.ApplicationUp
					AssertUp     func(t *testing.T, up *ttnpb.ApplicationUp)
					AssertDevice func(t *testing.T, dev *ttnpb.EndDevice)
				}{
					{
						Name: "JoinAccept",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_JoinAccept{
								JoinAccept: &ttnpb.ApplicationJoinAccept{
									SessionKeyId: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KekLabel:     kekLabel,
									},
									ReceivedAt: timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.NotBeNil)
							if override.GetValue() {
								a.So(up, should.Resemble, &ttnpb.ApplicationUp{
									EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
									Up: &ttnpb.ApplicationUp_JoinAccept{
										JoinAccept: &ttnpb.ApplicationJoinAccept{
											SessionKeyId: []byte{0x22},
											AppSKey: &ttnpb.KeyEnvelope{
												// AppSKey is []byte{0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22, 0x22}
												EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
												KekLabel:     kekLabel,
											},
											ReceivedAt: up.GetJoinAccept().ReceivedAt,
										},
									},
									CorrelationIds: up.CorrelationIds,
									ReceivedAt:     up.ReceivedAt,
								})
							} else {
								a.So(up, should.Resemble, &ttnpb.ApplicationUp{
									EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
									Up: &ttnpb.ApplicationUp_JoinAccept{
										JoinAccept: &ttnpb.ApplicationJoinAccept{
											SessionKeyId: []byte{0x22},
											ReceivedAt:   up.GetJoinAccept().ReceivedAt,
										},
									},
									CorrelationIds: up.CorrelationIds,
									ReceivedAt:     up.ReceivedAt,
								})
							}
						},
						AssertDevice: func(t *testing.T, dev *ttnpb.EndDevice) {
							a := assertions.New(t)
							a.So(dev.Session, should.BeNil)
							a.So(dev.PendingSession, should.Resemble, &ttnpb.Session{
								DevAddr: types.DevAddr{0x22, 0x22, 0x22, 0x22}.Bytes(),
								Keys: &ttnpb.SessionKeys{
									SessionKeyId: []byte{0x22},
									AppSKey: &ttnpb.KeyEnvelope{
										EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
										KekLabel:     kekLabel,
									},
								},
								LastAFCntDown: 0,
							})
						},
					},
					{
						Name: "UplinkMessage/PendingSession",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_UplinkMessage{
								UplinkMessage: &ttnpb.ApplicationUplink{
									RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
									Settings: &ttnpb.TxSettings{
										DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
										Frequency: 868000000,
									},
									SessionKeyId: []byte{0x22},
									FPort:        22,
									FCnt:         22,
									FrmPayload:   []byte{0x01},
									ReceivedAt:   timestamppb.New(now),
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							a.So(up, should.NotBeNil)
							if override.GetValue() {
								a.So(up, should.Resemble, &ttnpb.ApplicationUp{
									EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
									Up: &ttnpb.ApplicationUp_UplinkMessage{
										UplinkMessage: &ttnpb.ApplicationUplink{
											RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
											Settings: &ttnpb.TxSettings{
												DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
												Frequency: 868000000,
											},
											SessionKeyId: []byte{0x22},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0x01},
											AppSKey: &ttnpb.KeyEnvelope{
												EncryptedKey: []byte{0x39, 0x11, 0x40, 0x98, 0xa1, 0x5d, 0x6f, 0x92, 0xd7, 0xf0, 0x13, 0x21, 0x5b, 0x5b, 0x41, 0xa8, 0x98, 0x2d, 0xac, 0x59, 0x34, 0x76, 0x36, 0x18},
												KekLabel:     kekLabel,
											},
											VersionIds: registeredDevice.VersionIds,
											ReceivedAt: up.GetUplinkMessage().ReceivedAt,
										},
									},
									CorrelationIds: up.CorrelationIds,
									ReceivedAt:     up.ReceivedAt,
								})
							} else {
								a.So(up, should.Resemble, &ttnpb.ApplicationUp{
									EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
									Up: &ttnpb.ApplicationUp_UplinkMessage{
										UplinkMessage: &ttnpb.ApplicationUplink{
											RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
											Settings: &ttnpb.TxSettings{
												DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
												Frequency: 868000000,
											},
											SessionKeyId: []byte{0x22},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0xc1},
											ReceivedAt:   up.GetUplinkMessage().ReceivedAt,
										},
									},
									CorrelationIds: up.CorrelationIds,
									ReceivedAt:     up.ReceivedAt,
								})
							}
						},
					},
					{
						Name: "DownlinkQueueInvalidation",
						Message: &ttnpb.ApplicationUp{
							EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
							Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
								DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
									Downlinks: []*ttnpb.ApplicationDownlink{
										{
											SessionKeyId: []byte{0x22},
											FPort:        22,
											FCnt:         22,
											FrmPayload:   []byte{0x01},
										},
									},
									SessionKeyId: []byte{0x22},
								},
							},
						},
						AssertUp: func(t *testing.T, up *ttnpb.ApplicationUp) {
							a := assertions.New(t)
							if override.GetValue() {
								a.So(up, should.Resemble, &ttnpb.ApplicationUp{
									EndDeviceIds: withDevAddr(registeredDevice.Ids, types.DevAddr{0x22, 0x22, 0x22, 0x22}),
									Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
										DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
											Downlinks: []*ttnpb.ApplicationDownlink{
												{
													SessionKeyId: []byte{0x22},
													FPort:        22,
													FCnt:         22,
													FrmPayload:   []byte{0x01},
												},
											},
											SessionKeyId: []byte{0x22},
										},
									},
									CorrelationIds: up.CorrelationIds,
									ReceivedAt:     up.ReceivedAt,
								})
							} else {
								a.So(up, should.BeNil)
							}
						},
					},
				} {
					stepok := t.Run(step.Name, func(t *testing.T) {
						ns.upCh <- step.Message
						select {
						case msg := <-chs.up:
							if step.AssertUp != nil {
								step.AssertUp(t, msg)
							} else {
								t.Fatalf("Expected no upstream message but got %v", msg)
							}
						case <-time.After(Timeout):
							if step.AssertUp != nil {
								step.AssertUp(t, nil)
							} else {
								t.Fatal("Expected upstream timeout")
							}
						}
						if step.AssertDevice != nil {
							dev, err := deviceRegistry.Get(ctx, step.Message.EndDeviceIds, []string{"session", "pending_session"})
							if !a.So(err, should.BeNil) {
								t.FailNow()
							}
							step.AssertDevice(t, dev)
						}
					})
					if !stepok {
						t.FailNow()
					}
				}
			})

			t.Run("Downlink", func(t *testing.T) {
				ns.reset()
				devsFlush()
				deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(_ *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
					dev := ttnpb.Clone(registeredDevice)
					dev.Session = &ttnpb.Session{
						DevAddr: types.DevAddr{0x42, 0xff, 0xff, 0xff}.Bytes(),
						Keys: &ttnpb.SessionKeys{
							SessionKeyId: []byte{0x11},
							AppSKey: &ttnpb.KeyEnvelope{
								EncryptedKey: []byte{0x1f, 0xa6, 0x8b, 0xa, 0x81, 0x12, 0xb4, 0x47, 0xae, 0xf3, 0x4b, 0xd8, 0xfb, 0x5a, 0x7b, 0x82, 0x9d, 0x3e, 0x86, 0x23, 0x71, 0xd2, 0xcf, 0xe5},
								KekLabel:     kekLabel,
							},
						},
					}
					dev.SkipPayloadCryptoOverride = override
					return dev, []string{
						"ids",
						"session",
						"skip_payload_crypto_override",
					}, nil
				})
				t.Run("Push", func(t *testing.T) {
					a := assertions.New(t)
					for _, items := range [][]*ttnpb.ApplicationDownlink{
						{
							{
								FPort:      11,
								FCnt:       1,
								FrmPayload: []byte{0x1, 0x1, 0x1},
							},
							{
								FPort:      22,
								FCnt:       2,
								FrmPayload: []byte{0x2, 0x2, 0x2},
							},
						},
					} {
						chs.downPush <- &ttnpb.DownlinkQueueRequest{
							EndDeviceIds: registeredDevice.Ids,
							Downlinks:    items,
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
					res, err := as.DownlinkQueueList(ctx, registeredDevice.Ids)
					if a.So(err, should.BeNil) && a.So(res, should.HaveLength, 2) {
						a.So(res, should.Resemble, []*ttnpb.ApplicationDownlink{
							{
								SessionKeyId:   []byte{0x11},
								FPort:          11,
								FCnt:           1,
								FrmPayload:     []byte{0x1, 0x1, 0x1},
								CorrelationIds: res[0].CorrelationIds,
							},
							{
								SessionKeyId:   []byte{0x11},
								FPort:          22,
								FCnt:           2,
								FrmPayload:     []byte{0x2, 0x2, 0x2},
								CorrelationIds: res[1].CorrelationIds,
							},
						})
					}
				})
			})
		})
	}
}

func TestLocationFromPayload(t *testing.T) {
	a, ctx := test.New(t)

	registeredApplicationID := ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}

	// This device gets registered in the device registry of the Application Server.
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &registeredApplicationID,
			DeviceId:       "foo-device",
			JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		},
		Session: &ttnpb.Session{
			DevAddr: types.DevAddr{0x11, 0x11, 0x11, 0x11}.Bytes(),
			Keys: &ttnpb.SessionKeys{
				SessionKeyId: []byte{0x11},
				AppSKey: &ttnpb.KeyEnvelope{
					Key: types.AES128Key{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(), //nolint:lll
				},
			},
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter: ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
			UpFormatterParameter: `function decodeUplink(input) {
				return {
					data: {
						lat: 4.85564,
						lng: 52.3456341,
						alt: 16,
						acc: 14
					}
				};
			}`,
		},
	}

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.EndDeviceRegistry().Add(ctx, registeredDevice)

	devsRedisClient, devsFlush := test.NewRedis(ctx, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err := deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(ed *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return registeredDevice, []string{"ids", "session", "formatters"}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set device in registry: %s", err)
	}

	linksRedisClient, linksFlush := test.NewRedis(ctx, "applicationserver_test", "links")
	defer linksFlush()
	defer linksRedisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: linksRedisClient, LockTTL: test.Delay << 10}
	if err := linkRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err = linkRegistry.Set(ctx, &registeredApplicationID, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{}, nil, nil
	})
	if err != nil {
		t.Fatalf("Failed to set link in registry: %s", err)
	}

	distribRedisClient, distribFlush := test.NewRedis(ctx, "applicationserver_test", "traffic")
	defer distribFlush()
	defer distribRedisClient.Close()
	distribPubSub := distribredis.PubSub{Redis: distribRedisClient}

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9189",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			HTTP: config.HTTP{
				Listen: ":8100",
			},
		},
	})
	config := &applicationserver.Config{
		Devices: deviceRegistry,
		Links:   linkRegistry,
		Distribution: applicationserver.DistributionConfig{
			Global: applicationserver.GlobalDistributorConfig{
				PubSub: distribPubSub,
			},
		},
		EndDeviceMetadataStorage: applicationserver.EndDeviceMetadataStorageConfig{
			Location: applicationserver.EndDeviceLocationStorageConfig{
				Registry: metadata.NewClusterEndDeviceLocationRegistry(c, (1<<4)*Timeout),
			},
		},
		Downlinks: applicationserver.DownlinksConfig{
			ConfirmationConfig: applicationserver.ConfirmationConfig{
				DefaultRetryAttempts: 3,
				MaxRetryAttempts:     10,
			},
		},
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_APPLICATION_SERVER)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	sub, err := as.Subscribe(ctx, "test", nil, false)
	a.So(err, should.BeNil)

	now := time.Now().UTC()
	err = as.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIds: registeredDevice.Ids,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
				Settings: &ttnpb.TxSettings{
					DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
					Frequency: 868000000,
				},
				SessionKeyId: []byte{0x11},
				FPort:        11,
				FCnt:         11,
				FrmPayload:   []byte{0x11},
				ReceivedAt:   timestamppb.New(now),
			},
		},
	})
	a.So(err, should.BeNil)

	assertLocation := func(loc *ttnpb.Location) {
		a.So(loc.Latitude, should.AlmostEqual, 4.85564, 0.00001)
		a.So(loc.Longitude, should.AlmostEqual, 52.3456341, 0.00001)
		a.So(loc.Altitude, should.Equal, 16)
		a.So(loc.Accuracy, should.Equal, 14)
	}

	assertApplicationlocation := func(loc *ttnpb.ApplicationLocation) {
		a.So(loc.Service, should.Equal, "frm-payload")
		assertLocation(loc.Location)
	}

	// The uplink message and the location solved message may come out of order.
	// Expect exactly two messages.
	var loc *ttnpb.ApplicationLocation
	for i := 0; i < 2; i++ {
		select {
		case msg := <-sub.Up():
			msgLoc := msg.ApplicationUp.GetLocationSolved()
			if msgLoc != nil {
				loc = msgLoc
			}
		case <-time.After(Timeout):
			t.Fatalf("Expected upstream message %d timed out", i)
		}
	}
	if loc == nil {
		t.Fatal("Expected location solved message")
	}

	assertApplicationlocation(loc)

	time.Sleep(Timeout)

	dev, ok := is.EndDeviceRegistry().Get(ctx, &ttnpb.GetEndDeviceRequest{EndDeviceIds: registeredDevice.Ids})
	if !a.So(ok, should.BeNil) {
		t.FailNow()
	}

	if loc, ok := dev.Locations["frm-payload"]; a.So(ok, should.BeTrue) {
		assertLocation(loc)
	}
}

func TestUplinkNormalized(t *testing.T) {
	a, ctx := test.New(t)

	registeredApplicationID := ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}

	// This device gets registered in the device registry of the Application Server.
	registeredDevice := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &registeredApplicationID,
			DeviceId:       "foo-device",
			JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
		},
		Session: &ttnpb.Session{
			DevAddr: types.DevAddr{0x11, 0x11, 0x11, 0x11}.Bytes(),
			Keys: &ttnpb.SessionKeys{
				SessionKeyId: []byte{0x11},
				AppSKey: &ttnpb.KeyEnvelope{
					Key: types.AES128Key{0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11, 0x11}.Bytes(), //nolint:lll
				},
			},
		},
		Formatters: &ttnpb.MessagePayloadFormatters{
			UpFormatter: ttnpb.PayloadFormatter_FORMATTER_JAVASCRIPT,
			UpFormatterParameter: `function decodeUplink(input) {
				return {
					data: {
						air: {
							temperature: 21.5,
						}
					}
				};
			}`,
		},
	}

	is, isAddr, closeIS := mockis.New(ctx)
	defer closeIS()
	is.EndDeviceRegistry().Add(ctx, registeredDevice)

	devsRedisClient, devsFlush := test.NewRedis(ctx, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err := deviceRegistry.Set(ctx, registeredDevice.Ids, nil, func(ed *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
		return registeredDevice, []string{"ids", "session", "formatters"}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set device in registry: %s", err)
	}

	linksRedisClient, linksFlush := test.NewRedis(ctx, "applicationserver_test", "links")
	defer linksFlush()
	defer linksRedisClient.Close()
	linkRegistry := &redis.LinkRegistry{Redis: linksRedisClient, LockTTL: test.Delay << 10}
	if err := linkRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	_, err = linkRegistry.Set(ctx, &registeredApplicationID, nil, func(_ *ttnpb.ApplicationLink) (*ttnpb.ApplicationLink, []string, error) {
		return &ttnpb.ApplicationLink{}, nil, nil
	})
	if err != nil {
		t.Fatalf("Failed to set link in registry: %s", err)
	}

	distribRedisClient, distribFlush := test.NewRedis(ctx, "applicationserver_test", "traffic")
	defer distribFlush()
	defer distribRedisClient.Close()
	distribPubSub := distribredis.PubSub{Redis: distribRedisClient}

	c := componenttest.NewComponent(t, &component.Config{
		ServiceBase: config.ServiceBase{
			GRPC: config.GRPC{
				Listen:                      ":9189",
				AllowInsecureForCredentials: true,
			},
			Cluster: cluster.Config{
				IdentityServer: isAddr,
			},
			HTTP: config.HTTP{
				Listen: ":8100",
			},
		},
	})
	config := &applicationserver.Config{
		Devices: deviceRegistry,
		Links:   linkRegistry,
		Distribution: applicationserver.DistributionConfig{
			Global: applicationserver.GlobalDistributorConfig{
				PubSub: distribPubSub,
			},
		},
		EndDeviceMetadataStorage: applicationserver.EndDeviceMetadataStorageConfig{
			Location: applicationserver.EndDeviceLocationStorageConfig{
				Registry: metadata.NewClusterEndDeviceLocationRegistry(c, (1<<4)*Timeout),
			},
		},
		Downlinks: applicationserver.DownlinksConfig{
			ConfirmationConfig: applicationserver.ConfirmationConfig{
				DefaultRetryAttempts: 3,
				MaxRetryAttempts:     10,
			},
		},
	}
	as, err := applicationserver.New(c, config)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	roles := as.Roles()
	a.So(len(roles), should.Equal, 1)
	a.So(roles[0], should.Equal, ttnpb.ClusterRole_APPLICATION_SERVER)

	componenttest.StartComponent(t, c)
	defer c.Close()

	mustHavePeer(ctx, c, ttnpb.ClusterRole_ENTITY_REGISTRY)

	sub, err := as.Subscribe(ctx, "test", nil, false)
	a.So(err, should.BeNil)

	now := time.Now().UTC()
	err = as.Publish(ctx, &ttnpb.ApplicationUp{
		EndDeviceIds: registeredDevice.Ids,
		Up: &ttnpb.ApplicationUp_UplinkMessage{
			UplinkMessage: &ttnpb.ApplicationUplink{
				RxMetadata: []*ttnpb.RxMetadata{{GatewayIds: &ttnpb.GatewayIdentifiers{GatewayId: "gtw"}}},
				Settings: &ttnpb.TxSettings{
					DataRate:  &ttnpb.DataRate{Modulation: &ttnpb.DataRate_Lora{Lora: &ttnpb.LoRaDataRate{}}},
					Frequency: 868000000,
				},
				SessionKeyId: []byte{0x11},
				FPort:        11,
				FCnt:         11,
				FrmPayload:   []byte{0x11},
				ReceivedAt:   timestamppb.New(now),
			},
		},
	})
	a.So(err, should.BeNil)

	// The uplink message and the normalized payload message may come out of order.
	// Expect exactly two messages.
	var normalized *ttnpb.ApplicationUplinkNormalized
	for i := 0; i < 2; i++ {
		select {
		case msg := <-sub.Up():
			if n := msg.GetUplinkNormalized(); n != nil {
				normalized = n
			}
		case <-time.After(Timeout):
			t.Fatalf("Expected upstream message %d timed out", i)
		}
	}
	if normalized == nil {
		t.Fatalf("Expected uplink normalized message")
	}
	a.So(normalized.NormalizedPayload, should.Resemble, &structpb.Struct{
		Fields: map[string]*structpb.Value{
			"air": {
				Kind: &structpb.Value_StructValue{
					StructValue: &structpb.Struct{
						Fields: map[string]*structpb.Value{
							"temperature": {
								Kind: &structpb.Value_NumberValue{
									NumberValue: 21.5,
								},
							},
						},
					},
				},
			},
		},
	})
}

func TestApplicationServerCleanup(t *testing.T) {
	a, ctx := test.New(t)

	app1 := &ttnpb.ApplicationIdentifiers{ApplicationId: "app-1"}
	app2 := &ttnpb.ApplicationIdentifiers{ApplicationId: "app-2"}
	app3 := &ttnpb.ApplicationIdentifiers{ApplicationId: "app-3"}
	app4 := &ttnpb.ApplicationIdentifiers{ApplicationId: "app-4"}
	webhookList := []*ttnpb.ApplicationWebhookIdentifiers{
		{
			ApplicationIds: app1,
			WebhookId:      "test-1",
		},
		{
			ApplicationIds: app3,
			WebhookId:      "test-2",
		},
		{
			ApplicationIds: app4,
			WebhookId:      "test-3",
		},
	}

	pubsubList := []*ttnpb.ApplicationPubSubIdentifiers{
		{
			ApplicationIds: app2,
			PubSubId:       "test-1",
		},
		{
			ApplicationIds: app3,
			PubSubId:       "test-2",
		},
		{
			ApplicationIds: app1,
			PubSubId:       "test-3",
		},
	}
	deviceList := []*ttnpb.EndDevice{
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app1,
				DeviceId:       "dev-1",
				JoinEui:        types.EUI64{0x41, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x41, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app1,
				DeviceId:       "dev-2",
				JoinEui:        types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app2,
				DeviceId:       "dev-3",
				JoinEui:        types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app2,
				DeviceId:       "dev-4",
				JoinEui:        types.EUI64{0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x44, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app4,
				DeviceId:       "dev-5",
				JoinEui:        types.EUI64{0x45, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x45, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: app4,
				DeviceId:       "dev-6",
				JoinEui:        types.EUI64{0x46, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
				DevEui:         types.EUI64{0x46, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}.Bytes(),
			},
		},
	}

	associationList := []*ttnpb.ApplicationPackageAssociationIdentifiers{
		{
			EndDeviceIds: deviceList[0].Ids,
			FPort:        1,
		},
		{
			EndDeviceIds: deviceList[0].Ids,
			FPort:        1,
		},
		{
			EndDeviceIds: deviceList[2].Ids,
			FPort:        1,
		},
		{
			EndDeviceIds: deviceList[5].Ids,
			FPort:        1,
		},
	}
	defaultAssociationList := []*ttnpb.ApplicationPackageDefaultAssociationIdentifiers{
		{
			ApplicationIds: app1,
			FPort:          1,
		},
		{
			ApplicationIds: app1,
			FPort:          2,
		},
		{
			ApplicationIds: app2,
			FPort:          3,
		},
		{
			ApplicationIds: app4,
			FPort:          4,
		},
	}

	devsRedisClient, devsFlush := test.NewRedis(ctx, "applicationserver_test", "devices")
	defer devsFlush()
	defer devsRedisClient.Close()
	deviceRegistry := &redis.DeviceRegistry{Redis: devsRedisClient, LockTTL: test.Delay << 10}
	if err := deviceRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	webhooksRedisClient, webhooksFlush := test.NewRedis(ctx, "applicationserver_test", "webhooks")
	defer webhooksFlush()
	defer webhooksRedisClient.Close()
	webhookRegistry := iowebredis.WebhookRegistry{Redis: webhooksRedisClient, LockTTL: test.Delay << 10}
	if err := webhookRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	pubsubRedisClient, pubsubFlush := test.NewRedis(ctx, "applicationserver_test", "pubsub")
	defer pubsubFlush()
	defer pubsubRedisClient.Close()
	pubsubRegistry := iopubsubredis.PubSubRegistry{Redis: pubsubRedisClient, LockTTL: test.Delay << 10}
	if err := pubsubRegistry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}

	applicationPackagesRedisClient, applicationPackagesFlush := test.NewRedis(ctx, "applicationserver_test", "applicationpackages")
	defer applicationPackagesFlush()
	defer applicationPackagesRedisClient.Close()
	applicationPackagesRegistry, err := asioapredis.NewApplicationPackagesRegistry(
		ctx, applicationPackagesRedisClient, test.Delay<<10,
	)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for _, dev := range deviceList {
		ret, err := deviceRegistry.Set(ctx, dev.Ids, []string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
		}, func(stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
			return dev, []string{
				"ids.application_ids",
				"ids.dev_eui",
				"ids.device_id",
				"ids.join_eui",
			}, nil
		})
		if !a.So(err, should.BeNil) || !a.So(ret, should.NotBeNil) {
			t.Fatalf("Failed to create device: %s", err)
		}
	}

	for _, webID := range webhookList {
		_, err := webhookRegistry.Set(ctx,
			webID,
			[]string{
				"ids.application_ids",
				"ids.webhook_id",
			},
			func(web *ttnpb.ApplicationWebhook) (*ttnpb.ApplicationWebhook, []string, error) {
				return &ttnpb.ApplicationWebhook{
						Ids:     webID,
						BaseUrl: "https://example.com",
						Format:  "json",
					},
					[]string{
						"ids.application_ids",
						"ids.webhook_id",
						"base_url",
						"format",
					}, nil
			})
		a.So(err, should.BeNil)
	}

	for _, pubsubID := range pubsubList {
		_, err := pubsubRegistry.Set(ctx, pubsubID,
			[]string{
				"ids.application_ids",
				"ids.pub_sub_id",
			},
			func(ps *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
				return &ttnpb.ApplicationPubSub{
						Ids: pubsubID,
						Provider: &ttnpb.ApplicationPubSub_Mqtt{
							Mqtt: &ttnpb.ApplicationPubSub_MQTTProvider{
								ServerUrl: "mqtt://example.com",
							},
						},
						Format: "json",
					},
					[]string{
						"ids.application_ids",
						"ids.pub_sub_id",
						"provider",
						"format",
					},
					nil
			})
		a.So(err, should.BeNil)
	}

	for _, associationID := range associationList {
		_, err := applicationPackagesRegistry.SetAssociation(ctx, associationID,
			[]string{
				"ids.end_device_ids.application_ids",
				"ids.end_device_ids.device_id",
				"ids.f_port",
			},
			func(as *ttnpb.ApplicationPackageAssociation) (*ttnpb.ApplicationPackageAssociation, []string, error) {
				return &ttnpb.ApplicationPackageAssociation{
						Ids:         associationID,
						PackageName: "example",
					},
					[]string{
						"ids.end_device_ids.application_ids",
						"ids.end_device_ids.device_id",
						"ids.f_port",
						"package_name",
					},
					nil
			})
		a.So(err, should.BeNil)
	}

	for _, defaultAssociationID := range defaultAssociationList {
		_, err := applicationPackagesRegistry.SetDefaultAssociation(ctx, defaultAssociationID,
			[]string{
				"ids.application_ids",
				"ids.f_port",
			},
			func(as *ttnpb.ApplicationPackageDefaultAssociation) (*ttnpb.ApplicationPackageDefaultAssociation, []string, error) {
				return &ttnpb.ApplicationPackageDefaultAssociation{
						Ids:         defaultAssociationID,
						PackageName: "example",
					},
					[]string{
						"ids.application_ids",
						"ids.f_port",
						"package_name",
					},
					nil
			})
		a.So(err, should.BeNil)
	}

	// Mock IS application and device sets
	isApplicationSet := map[string]struct{}{
		unique.ID(ctx, app3): {},
		unique.ID(ctx, app4): {},
	}
	isDeviceSet := map[string]struct{}{
		unique.ID(ctx, deviceList[4].Ids): {},
		unique.ID(ctx, deviceList[5].Ids): {},
	}

	// Test cleaner initialization (or just range to local set)
	pubsubCleaner := &pubsub.RegistryCleaner{
		PubSubRegistry: pubsubRegistry,
	}
	err = pubsubCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(pubsubCleaner.LocalSet, should.HaveLength, 3)

	webhookCleaner := &web.RegistryCleaner{
		WebRegistry: webhookRegistry,
	}
	err = webhookCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(webhookCleaner.LocalSet, should.HaveLength, 3)

	devCleaner := &applicationserver.RegistryCleaner{
		DevRegistry: deviceRegistry,
	}
	err = devCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(devCleaner.LocalSet, should.HaveLength, 6)

	packagesCleaner := &packages.RegistryCleaner{
		ApplicationPackagesRegistry: applicationPackagesRegistry,
	}
	err = packagesCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(packagesCleaner.LocalApplicationSet, should.HaveLength, 3)
	a.So(packagesCleaner.LocalDeviceSet, should.HaveLength, 3)

	// Test cleaning data
	err = pubsubCleaner.CleanData(ctx, isApplicationSet)
	a.So(err, should.BeNil)
	pubsubCleaner.RangeToLocalSet(ctx)
	a.So(pubsubCleaner.LocalSet, should.HaveLength, 1)

	err = webhookCleaner.CleanData(ctx, isApplicationSet)
	a.So(err, should.BeNil)
	webhookCleaner.RangeToLocalSet(ctx)
	a.So(webhookCleaner.LocalSet, should.HaveLength, 2)

	err = devCleaner.CleanData(ctx, isDeviceSet)
	a.So(err, should.BeNil)
	devCleaner.RangeToLocalSet(ctx)
	a.So(devCleaner.LocalSet, should.HaveLength, 2)

	err = packagesCleaner.CleanData(ctx, isDeviceSet, isApplicationSet)
	a.So(err, should.BeNil)
	packagesCleaner.RangeToLocalSet(ctx)
	a.So(packagesCleaner.LocalApplicationSet, should.HaveLength, 1)
	a.So(packagesCleaner.LocalDeviceSet, should.HaveLength, 1)
}
