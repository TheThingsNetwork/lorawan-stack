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

package grpc_test

import (
	"context"
	"fmt"
	"sync"
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io"
	. "go.thethings.network/lorawan-stack/pkg/applicationserver/io/grpc"
	"go.thethings.network/lorawan-stack/pkg/applicationserver/io/mock"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/log"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	registeredApplicationUID = "test-app"
	registeredApplicationID  = ttnpb.ApplicationIdentifiers{ApplicationID: "test-app"}
	registeredApplicationKey = "test-key"

	timeout = 10 * test.Delay
)

func TestAuthentication(t *testing.T) {
	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)

	as := mock.NewServer()
	srv := New(as)

	for _, tc := range []struct {
		ID  ttnpb.ApplicationIdentifiers
		Key string
		OK  bool
	}{
		{
			ID:  registeredApplicationID,
			Key: registeredApplicationKey,
			OK:  true,
		},
		{
			ID:  registeredApplicationID,
			Key: "invalid-key",
			OK:  false,
		},
		{
			ID:  ttnpb.ApplicationIdentifiers{ApplicationID: "invalid-application"},
			Key: "invalid-key",
			OK:  false,
		},
	} {
		t.Run(fmt.Sprintf("%v:%v", tc.ID.ApplicationID, tc.Key), func(t *testing.T) {
			a := assertions.New(t)

			ctx, cancelCtx := context.WithCancel(ctx)
			stream := &mockAppAsLinkServerStream{
				MockServerStream: &test.MockServerStream{
					MockStream: &test.MockStream{
						ContextFunc: contextWithKey(ctx, tc.Key),
					},
				},
			}

			wg := sync.WaitGroup{}
			wg.Add(1)
			go func() {
				defer wg.Done()
				err := srv.Subscribe(&tc.ID, stream)
				if tc.OK && !a.So(errors.IsCanceled(err), should.BeTrue) {
					t.Fatalf("Unexpected link error: %v", err)
				}
				if !tc.OK && !a.So(errors.IsCanceled(err), should.BeFalse) {
					t.FailNow()
				}
			}()

			cancelCtx()
			wg.Wait()
		})
	}
}

func TestTraffic(t *testing.T) {
	a := assertions.New(t)

	ctx := log.NewContext(test.Context(), test.GetLogger(t))
	ctx = newContextWithRightsFetcher(ctx)

	as := mock.NewServer()
	srv := New(as)

	upCh := make(chan *ttnpb.ApplicationUp)

	stream := &mockAppAsLinkServerStream{
		MockServerStream: &test.MockServerStream{
			MockStream: &test.MockStream{
				ContextFunc: contextWithKey(ctx, registeredApplicationKey),
			},
		},
		SendFunc: func(up *ttnpb.ApplicationUp) error {
			upCh <- up
			return nil
		},
	}

	go func() {
		if err := srv.Subscribe(&registeredApplicationID, stream); err != nil {
			if !a.So(errors.IsCanceled(err), should.BeTrue) {
				t.FailNow()
			}
		}
	}()

	var sub *io.Subscription
	select {
	case sub = <-as.Subscriptions():
	case <-time.After(timeout):
		t.Fatal("Subscription timeout")
	}

	t.Run("Upstream", func(t *testing.T) {
		a := assertions.New(t)

		up := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: registeredApplicationID,
				DeviceID:               "foo-device",
			},
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					FRMPayload: []byte{0x01, 0x02, 0x03},
				},
			},
		}
		if err := sub.SendUp(up); !a.So(err, should.BeNil) {
			t.FailNow()
		}

		select {
		case actual := <-upCh:
			a.So(actual, should.Resemble, up)
		case <-time.After(timeout):
			t.Fatal("Receive expected upstream message timeout")
		}
	})

	t.Run("Downstream", func(t *testing.T) {
		a := assertions.New(t)
		ids := ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: registeredApplicationID,
			DeviceID:               "foo-device",
		}

		// List: unauthorized.
		{
			_, err := srv.DownlinkQueueList(ctx, &ids)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// List: happy flow; no items.
		{
			res, err := srv.DownlinkQueueList(contextWithKey(ctx, registeredApplicationKey)(), &ids)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 0)
		}

		// Push: unauthorized.
		{
			_, err := srv.DownlinkQueuePush(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      1,
						FRMPayload: []byte{0x01, 0x01, 0x01},
					},
				},
			})
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Push and assert content: happy flow.
		{
			_, err := srv.DownlinkQueuePush(contextWithKey(ctx, registeredApplicationKey)(), &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						SessionKeyID:   []byte{0x11, 0x22, 0x33, 0x44}, // This gets discarded.
						FPort:          1,
						FCnt:           100, // This gets discarded.
						FRMPayload:     []byte{0x01, 0x01, 0x01},
						Confirmed:      true,
						CorrelationIDs: []string{"test"},
					},
					{
						FPort:      2,
						FRMPayload: []byte{0x02, 0x02, 0x02},
					},
				},
			})
			a.So(err, should.BeNil)
		}
		{
			_, err := srv.DownlinkQueuePush(contextWithKey(ctx, registeredApplicationKey)(), &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      3,
						FRMPayload: []byte{0x03, 0x03, 0x03},
					},
				},
			})
			a.So(err, should.BeNil)
		}
		{
			res, err := srv.DownlinkQueueList(contextWithKey(ctx, registeredApplicationKey)(), &ids)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 3)
			a.So(res.Downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{
				{
					FPort:          1,
					Confirmed:      true,
					FRMPayload:     []byte{0x01, 0x01, 0x01},
					CorrelationIDs: []string{"test"},
				},
				{
					FPort:      2,
					FRMPayload: []byte{0x02, 0x02, 0x02},
				},
				{
					FPort:      3,
					FRMPayload: []byte{0x03, 0x03, 0x03},
				},
			})
		}

		// Replace: unauthorized.
		{
			_, err := srv.DownlinkQueueReplace(ctx, &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      4,
						FRMPayload: []byte{0x04, 0x04, 0x04},
					},
				},
			})
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// Replace and assert content: happy flow.
		{
			_, err := srv.DownlinkQueueReplace(contextWithKey(ctx, registeredApplicationKey)(), &ttnpb.DownlinkQueueRequest{
				EndDeviceIdentifiers: ids,
				Downlinks: []*ttnpb.ApplicationDownlink{
					{
						FPort:      4,
						FCnt:       100, // This gets discarded.
						FRMPayload: []byte{0x04, 0x04, 0x04},
						Confirmed:  true,
					},
				},
			})
			a.So(err, should.BeNil)
		}
		{
			res, err := srv.DownlinkQueueList(contextWithKey(ctx, registeredApplicationKey)(), &ids)
			a.So(err, should.BeNil)
			a.So(res.Downlinks, should.HaveLength, 1)
			a.So(res.Downlinks, should.Resemble, []*ttnpb.ApplicationDownlink{
				{
					FPort:      4,
					FRMPayload: []byte{0x04, 0x04, 0x04},
					Confirmed:  true,
				},
			})
		}
	})
}
