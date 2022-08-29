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

package pubsub_test

import (
	"context"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/formatters"
	mock_server "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider"
	mock_provider "go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/provider/mock"
	"go.thethings.network/lorawan-stack/v3/pkg/applicationserver/io/pubsub/redis"
	"go.thethings.network/lorawan-stack/v3/pkg/component"
	componenttest "go.thethings.network/lorawan-stack/v3/pkg/component/test"
	"go.thethings.network/lorawan-stack/v3/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	cloudpubsub "gocloud.dev/pubsub"
)

type messageWithError struct {
	*cloudpubsub.Message
	error
}

func TestPubSub(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)

	redisClient, flush := test.NewRedis(ctx, "pubsub_test")
	defer flush()
	defer redisClient.Close()
	registry := &redis.PubSubRegistry{
		Redis:   redisClient,
		LockTTL: test.Delay << 10,
	}
	if err := registry.Init(ctx); !a.So(err, should.BeNil) {
		t.FailNow()
	}
	ids := &ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIds: registeredApplicationID,
		PubSubId:       registeredPubSubID,
	}

	ps := &ttnpb.ApplicationPubSub{
		Ids: &ttnpb.ApplicationPubSubIdentifiers{
			ApplicationIds: registeredApplicationID,
			PubSubId:       registeredPubSubID,
		},
		Provider: &ttnpb.ApplicationPubSub_Nats{
			Nats: &ttnpb.ApplicationPubSub_NATSProvider{
				ServerUrl: "nats://localhost",
			},
		},
		Format:    "json",
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
	paths := []string{
		"base_topic",
		"downlink_ack",
		"downlink_failed",
		"downlink_nack",
		"downlink_queued",
		"downlink_queue_invalidated",
		"downlink_sent",
		"downlink_push",
		"downlink_replace",
		"format",
		"ids",
		"provider",
		"join_accept",
		"location_solved",
		"uplink_normalized",
		"uplink_message",
		"service_data",
	}

	_, err := registry.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
		return ps, paths, nil
	})
	if err != nil {
		t.Fatalf("Failed to set pubsub in registry: %s", err)
	}

	result, err := registry.List(ctx, ids.ApplicationIds, paths)
	a.So(err, should.BeNil)
	if a.So(len(result), should.Equal, 1) {
		a.So(result[0], should.Resemble, ps)
	}

	mockProvider, err := provider.GetProvider(&ttnpb.ApplicationPubSub{
		Provider: &ttnpb.ApplicationPubSub_Nats{},
	})
	a.So(mockProvider, should.NotBeNil)
	a.So(err, should.BeNil)
	mockImpl := mockProvider.(*mock_provider.Impl)

	c := componenttest.NewComponent(t, &component.Config{})
	io := mock_server.NewServer(c)
	_, err = pubsub.New(c, io, registry, make(pubsub.ProviderStatuses))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	componenttest.StartComponent(t, c)
	defer c.Close()

	sub := <-io.Subscriptions()
	conn := <-mockImpl.OpenConnectionCh

	//nolint:paralleltest
	t.Run("Upstream", func(t *testing.T) {
		ctx, cancel := context.WithCancel(ctx)
		defer cancel()

		for _, tc := range []struct {
			Name         string
			Message      *ttnpb.ApplicationUp
			Subscription *cloudpubsub.Subscription
		}{
			{
				Name: "UplinkMessage",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{
							SessionKeyId: []byte{0x11},
							FPort:        42,
							FCnt:         42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.UplinkMessage,
			},
			{
				Name: "UplinkNormalized",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkNormalized{
						UplinkNormalized: &ttnpb.ApplicationUplinkNormalized{
							SessionKeyId: []byte{0x11},
							FPort:        42,
							FCnt:         42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
							NormalizedPayload: &pbtypes.Struct{
								Fields: map[string]*pbtypes.Value{
									"air": {
										Kind: &pbtypes.Value_StructValue{
											StructValue: &pbtypes.Struct{
												Fields: map[string]*pbtypes.Value{
													"temperature": {
														Kind: &pbtypes.Value_NumberValue{
															NumberValue: 21.5,
														},
													},
												},
											},
										},
									},
								},
							},
						},
					},
				},
				Subscription: conn.UplinkNormalized,
			},
			{
				Name: "JoinAccept",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: &ttnpb.ApplicationJoinAccept{
							SessionKeyId: []byte{0x22},
						},
					},
				},
				Subscription: conn.JoinAccept,
			},
			{
				Name: "DownlinkMessage/Ack",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkAck{
						DownlinkAck: &ttnpb.ApplicationDownlink{
							SessionKeyId: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkAck,
			},
			{
				Name: "DownlinkMessage/Nack",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkNack{
						DownlinkNack: &ttnpb.ApplicationDownlink{
							SessionKeyId: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkNack,
			},
			{
				Name: "DownlinkMessage/Sent",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkSent{
						DownlinkSent: &ttnpb.ApplicationDownlink{
							SessionKeyId: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkSent,
			},
			{
				Name: "DownlinkMessage/Queued",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkQueued{
						DownlinkQueued: &ttnpb.ApplicationDownlink{
							SessionKeyId: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FrmPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkQueued,
			},
			{
				Name: "DownlinkMessage/QueueInvalidated",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkQueueInvalidated{
						DownlinkQueueInvalidated: &ttnpb.ApplicationInvalidatedDownlinks{
							Downlinks: []*ttnpb.ApplicationDownlink{
								{
									SessionKeyId: []byte{0x22},
									FCnt:         42,
									FPort:        42,
									FrmPayload:   []byte{0x1, 0x2, 0x3},
								},
							},
							LastFCntDown: 42,
							SessionKeyId: []byte{0x22},
						},
					},
				},
				Subscription: conn.DownlinkQueueInvalidated,
			},
			{
				Name: "DownlinkMessage/Failed",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							Downlink: &ttnpb.ApplicationDownlink{
								SessionKeyId: []byte{0x22},
								FCnt:         42,
								FPort:        42,
								FrmPayload:   []byte{0x1, 0x2, 0x3},
							},
							Error: &ttnpb.ErrorDetails{
								Name: "test",
							},
						},
					},
				},
				Subscription: conn.DownlinkFailed,
			},
			{
				Name: "LocationSolved",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_LocationSolved{
						LocationSolved: &ttnpb.ApplicationLocation{
							Location: &ttnpb.Location{
								Latitude:  10,
								Longitude: 20,
								Altitude:  30,
							},
							Service: "test",
						},
					},
				},
				Subscription: conn.LocationSolved,
			},
			{
				Name: "ServiceData",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIds: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_ServiceData{
						ServiceData: &ttnpb.ApplicationServiceData{
							Data: &pbtypes.Struct{
								Fields: map[string]*pbtypes.Value{
									"battery": {
										Kind: &pbtypes.Value_NumberValue{
											NumberValue: 42.0,
										},
									},
								},
							},
							Service: "test",
						},
					},
				},
				Subscription: conn.ServiceData,
			},
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				err := sub.Publish(ctx, tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				ch := make(chan messageWithError, 10)
				go func() {
					msg, err := tc.Subscription.Receive(ctx)
					ch <- messageWithError{msg, err}
				}()

				var msg messageWithError
				select {
				case msg = <-ch:
					a.So(msg.Message, should.NotBeNil)
					a.So(err, should.BeNil)
					msg.Ack()
				case <-time.After(timeout):
					t.Fatal("Expected message but nothing received")
				}
				expectedBody, err := formatters.JSON.FromUp(tc.Message)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}
				a.So(msg.Message.Body, should.Resemble, expectedBody)
			})
			if !tcok {
				t.FailNow()
			}
		}
	})

	//nolint:paralleltest
	t.Run("Downstream", func(t *testing.T) {
		for _, tc := range []struct {
			Name     string
			Topic    *cloudpubsub.Topic
			Message  *ttnpb.DownlinkQueueRequest
			Expected []*ttnpb.ApplicationDownlink
		}{
			{
				Name:  "ValidPush",
				Topic: conn.Push,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIds: registeredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x1, 0x1, 0x1},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FrmPayload: []byte{0x1, 0x1, 0x1},
					},
				},
			},
			{
				Name:  "ValidReplace",
				Topic: conn.Replace,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIds: registeredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x2, 0x2, 0x2},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FrmPayload: []byte{0x2, 0x2, 0x2},
					},
				},
			},
			{
				Name:  "InvalidPush",
				Topic: conn.Push,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIds: &unregisteredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x3, 0x3, 0x3},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FrmPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
					},
				},
			},
			{
				Name:  "InvalidReplace",
				Topic: conn.Replace,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIds: &unregisteredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FrmPayload: []byte{0x4, 0x4, 0x4},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FrmPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
					},
				},
			},
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				buf, err := jsonpb.TTN().Marshal(tc.Message)
				a.So(err, should.BeNil)

				if err := tc.Topic.Send(ctx, &cloudpubsub.Message{
					Body: buf,
				}); !a.So(err, should.BeNil) {
					t.FailNow()
				}

				time.Sleep(timeout)

				res, err := io.DownlinkQueueList(ctx, registeredDeviceID)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Expected)
			})
			if !tcok {
				t.FailNow()
			}
		}
	})
}
