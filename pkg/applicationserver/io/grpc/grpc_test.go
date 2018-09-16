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
	iotesting "go.thethings.network/lorawan-stack/pkg/applicationserver/io/testing"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
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

	as := iotesting.NewServer()
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
				err := srv.Subscribe(&tc.ID, stream)
				if tc.OK && !a.So(errors.IsCanceled(err), should.BeTrue) {
					t.Fatalf("Unexpected link error: %v", err)
				}
				if !tc.OK && !a.So(errors.IsCanceled(err), should.BeFalse) {
					t.FailNow()
				}
				wg.Done()
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

	as := iotesting.NewServer()
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

	var conn *io.Connection
	select {
	case conn = <-as.Connections():
	case <-time.After(timeout):
		t.Fatal("Connection timeout")
	}

	t.Run("Upstream", func(t *testing.T) {
		a := assertions.New(t)

		up := &ttnpb.ApplicationUp{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID: "foo-device",
			},
			Up: &ttnpb.ApplicationUp_UplinkMessage{
				UplinkMessage: &ttnpb.ApplicationUplink{
					FRMPayload: []byte{0x01, 0x02, 0x03},
				},
			},
		}
		if err := conn.SendUp(up); !a.So(err, should.BeNil) {
			t.FailNow()
		}

		select {
		case actual := <-upCh:
			a.So(actual, should.Resemble, up)
		case <-time.After(timeout):
			t.Fatal("Receive expected upstream message timeout")
		}
	})
}
