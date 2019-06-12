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
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	. "go.thethings.network/lorawan-stack/pkg/applicationserver/io/pubsub"
	"go.thethings.network/lorawan-stack/pkg/jsonpb"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
	"gocloud.dev/pubsub"
	_ "gocloud.dev/pubsub/mempubsub"
)

var (
	registeredApplicationID   = ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}
	unregisteredApplicationID = ttnpb.ApplicationIdentifiers{ApplicationID: "no-app"}
	registeredDeviceID        = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: registeredApplicationID,
		DeviceID:               "test-device",
	}
	unregisteredDeviceID = ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: unregisteredApplicationID,
		DeviceID:               "no-device",
	}
	keys = ttnpb.NewPopulatedSessionKeys(test.Randy, false)

	timeout = (1 << 8) * test.Delay
)

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx, cancelCtx := context.WithCancel(ctx)
	defer cancelCtx()

	as := mock.NewServer()
	subs, err := Start(ctx, as, formatters.JSON, []string{"mem://cloud_test"}, []string{"mem://cloud_test"})
	a.So(err, should.BeNil)

	topic, err := pubsub.OpenTopic(ctx, "mem://cloud_test")
	a.So(err, should.BeNil)
	defer topic.Shutdown(ctx)

	t.Run("Upstream", func(t *testing.T) {
		for _, tc := range []struct {
			Name    string
			Message *ttnpb.ApplicationUp
		}{
			{
				Name: "Uplink",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_UplinkMessage{
						UplinkMessage: &ttnpb.ApplicationUplink{FRMPayload: []byte{0x2, 0x2, 0x2}},
					},
				},
			},
			{
				Name: "JoinAccept",
				Message: &ttnpb.ApplicationUp{
					EndDeviceIdentifiers: registeredDeviceID,
					Up: &ttnpb.ApplicationUp_JoinAccept{
						JoinAccept: &ttnpb.ApplicationJoinAccept{
							AppSKey:      keys.AppSKey,
							SessionKeyID: keys.SessionKeyID,
						},
					},
				},
			},
		} {
			t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)

				upCh := make(chan *ttnpb.ApplicationUp)

				subscription, err := pubsub.OpenSubscription(ctx, "mem://cloud_test")
				a.So(err, should.BeNil)
				defer subscription.Shutdown(ctx)

				go func() {
					msg, err := subscription.Receive(ctx)
					a.So(err, should.BeNil)

					up := &ttnpb.ApplicationUp{}
					err = jsonpb.TTN().Unmarshal(msg.Body, up)
					a.So(err, should.BeNil)
					upCh <- up
				}()

				err = subs[0].SendUp(ctx, tc.Message)
				a.So(err, should.BeNil)

				select {
				case up := <-upCh:
					a.So(up, should.Resemble, tc.Message)

				case <-time.After(timeout):
					t.Fatal("Receive expected upstream timeout")
				}
			})
		}
	})

	t.Run("Downstream", func(t *testing.T) {
		for _, tc := range []struct {
			Name     string
			IDs      ttnpb.EndDeviceIdentifiers
			Message  *ttnpb.DownlinkQueueOperation
			Expected []*ttnpb.ApplicationDownlink
		}{
			{
				Name: "PushEmptyQueue",
				IDs:  registeredDeviceID,
				Message: &ttnpb.DownlinkQueueOperation{
					Operation:            ttnpb.DownlinkQueueOperation_PUSH,
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
				Name: "ReplaceEmptyQueue",
				IDs:  registeredDeviceID,
				Message: &ttnpb.DownlinkQueueOperation{
					Operation:            ttnpb.DownlinkQueueOperation_REPLACE,
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
				Name: "PushInvalidDevice",
				IDs:  registeredDeviceID,
				Message: &ttnpb.DownlinkQueueOperation{
					Operation:            ttnpb.DownlinkQueueOperation_PUSH,
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
		} {
			tcok := t.Run(tc.Name, func(t *testing.T) {
				a := assertions.New(t)
				buf, err := jsonpb.TTN().Marshal(tc.Message)
				a.So(err, should.BeNil)
				if err := topic.Send(ctx, &pubsub.Message{Body: buf}); !a.So(err, should.BeNil) {
					t.FailNow()
				}
				<-time.After(timeout)
				res, err := as.DownlinkQueueList(ctx, tc.IDs)
				a.So(err, should.BeNil)
				a.So(res, should.Resemble, tc.Expected)
			})
			if !tcok {
				t.FailNow()
			}
		}
	})
}
