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

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/formatters"
	mock_server "go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider"
	mock_provider "go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/provider/mock"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub/redis"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	cloudpubsub "gocloud.dev/pubsub"
)

type messageWithError struct {
	*cloudpubsub.Message
	error
}

func TestPubSub(t *testing.T) {
	a := assertions.New(t)
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	redisClient, flush := test.NewRedis(t, "pubsub_test")
	defer flush()
	defer redisClient.Close()
	registry := &redis.PubSubRegistry{
		Redis: redisClient,
	}
	ids := ttnpb.ApplicationPubSubIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		PubSubID:               registeredPubSubID,
	}

	_, err := registry.Set(ctx, ids, nil, func(_ *ttnpb.ApplicationPubSub) (*ttnpb.ApplicationPubSub, []string, error) {
		return &ttnpb.ApplicationPubSub{
				ApplicationPubSubIdentifiers: ttnpb.ApplicationPubSubIdentifiers{
					ApplicationIdentifiers: registeredApplicationID,
					PubSubID:               registeredPubSubID,
				},
				Attributes: map[string]string{
					mock_provider.MockAckDeadline: timeout.String(),
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
			},
			[]string{
				"base_topic",
				"downlink_ack",
				"downlink_failed",
				"downlink_nack",
				"downlink_queued",
				"downlink_sent",
				"downlink_push",
				"downlink_replace",
				"format",
				"attributes",
				"ids",
				"join_accept",
				"location_solved",
				"uplink_message",
			}, nil
	})
	if err != nil {
		t.Fatalf("Failed to set pubsub in registry: %s", err)
	}

	mockProvider, err := provider.GetProvider(ttnpb.ApplicationPubSub_AWSSNSSQS)
	a.So(mockProvider, should.NotBeNil)
	a.So(err, should.BeNil)
	mockImpl := mockProvider.(*mock_provider.Impl)

	io := mock_server.NewServer()
	c := component.MustNew(test.GetLogger(t), &component.Config{})
	_, err = pubsub.Start(c, io, registry)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	test.Must(nil, c.Start())
	defer c.Close()

	sub := <-io.Subscriptions()
	conn := <-mockImpl.OpenConnectionCh

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
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{
							SessionKeyID: []byte{0x11},
							FPort:        42,
							FCnt:         42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.UplinkMessage,
			},
			{
				Name: "JoinAccept",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: &ttnpb.ApplicationJoinAccept{
							SessionKeyID: []byte{0x22},
						},
					},
				},
				Subscription: conn.JoinAccept,
			},
			{
				Name: "DownlinkMessage/Ack",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkAck{
						DownlinkAck: &ttnpb.ApplicationDownlink{
							SessionKeyID: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkAck,
			},
			{
				Name: "DownlinkMessage/Nack",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkNack{
						DownlinkNack: &ttnpb.ApplicationDownlink{
							SessionKeyID: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkNack,
			},
			{
				Name: "DownlinkMessage/Sent",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkSent{
						DownlinkSent: &ttnpb.ApplicationDownlink{
							SessionKeyID: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkSent,
			},
			{
				Name: "DownlinkMessage/Queued",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkQueued{
						DownlinkQueued: &ttnpb.ApplicationDownlink{
							SessionKeyID: []byte{0x22},
							FCnt:         42,
							FPort:        42,
							FRMPayload:   []byte{0x1, 0x2, 0x3},
						},
					},
				},
				Subscription: conn.DownlinkQueued,
			},
			{
				Name: "DownlinkMessage/Failed",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_DownlinkFailed{
						DownlinkFailed: &ttnpb.ApplicationDownlinkFailed{
							ApplicationDownlink: ttnpb.ApplicationDownlink{
								SessionKeyID: []byte{0x22},
								FCnt:         42,
								FPort:        42,
								FRMPayload:   []byte{0x1, 0x2, 0x3},
							},
							Error: ttnpb.ErrorDetails{
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
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_LocationSolved{
						LocationSolved: &ttnpb.ApplicationLocation{
							Location: ttnpb.Location{
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
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				ctx, cancel := context.WithCancel(ctx)
				defer cancel()

				err := sub.SendUp(ctx, tc.Message)
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
					EndDeviceIdentifiers: registeredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x1, 0x1, 0x1},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x1, 0x1, 0x1},
					},
				},
			},
			{
				Name:  "ValidReplace",
				Topic: conn.Replace,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIdentifiers: registeredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x2, 0x2, 0x2},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2},
					},
				},
			},
			{
				Name:  "InvalidPush",
				Topic: conn.Push,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIdentifiers: unregisteredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x3, 0x3, 0x3},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
					},
				},
			},
			{
				Name:  "InvalidReplace",
				Topic: conn.Replace,
				Message: &ttnpb.DownlinkQueueRequest{
					EndDeviceIdentifiers: unregisteredDeviceID,
					Downlinks: []*ttnpb.ApplicationDownlink{
						{
							FPort:      42,
							FRMPayload: []byte{0x4, 0x4, 0x4},
						},
					},
				},
				Expected: []*ttnpb.ApplicationDownlink{
					{
						FPort:      42,
						FRMPayload: []byte{0x2, 0x2, 0x2}, // Do not expect a change.
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
