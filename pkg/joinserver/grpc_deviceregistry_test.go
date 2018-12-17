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

package joinserver_test

import (
	"context"
	"log"
	"testing"

	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/component"
	"go.thethings.network/lorawan-stack/pkg/errors"
	. "go.thethings.network/lorawan-stack/pkg/joinserver"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestDeviceRegistry(t *testing.T) {
	errTest := errors.New("test")

	for _, tc := range []struct {
		Name string

		Context func(context.Context) context.Context

		GetByEUI func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.EndDevice, error)

		DeviceRequest  *ttnpb.GetEndDeviceRequest
		DeviceResponse *ttnpb.EndDevice

		ErrorAssertion func(*testing.T, error) bool
	}{
		{
			Name: "Registry error",
			GetByEUI: func(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				return nil, errTest
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "id-test",
					DevEUI:   &types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
			ErrorAssertion: func(t *testing.T, err error) bool {
				return assertions.New(t).So(err, should.EqualErrorOrDefinition, errTest)
			},
		},
		{
			Name: "Working retrieve",
			GetByEUI: func(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
				a := assertions.New(test.MustTFromContext(ctx))
				a.So(joinEUI, should.Resemble, types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				a.So(devEUI, should.Resemble, types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff})
				return &ttnpb.EndDevice{
					EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
						DeviceID: "id-test",
						DevEUI:   &types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					},
				}, nil
			},
			DeviceRequest: &ttnpb.GetEndDeviceRequest{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "id-test",
					DevEUI:   &types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
					JoinEUI:  &types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
			DeviceResponse: &ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "id-test",
					DevEUI:   &types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
				},
			},
		},
	} {
		t.Run(tc.Name, func(t *testing.T) {
			a := assertions.New(t)

			ctx := clusterauth.NewContext(test.ContextWithT(test.Context(), t), nil)
			if tc.Context != nil {
				ctx = tc.Context(ctx)
			}

			js := JsDeviceServer{
				JS: test.Must(New(
					component.MustNew(test.GetLogger(t), &component.Config{}),
					&Config{
						Keys:    &MockKeyRegistry{},
						Devices: &MockDeviceRegistry{GetByEUIFunc: tc.GetByEUI},
					},
				)).(*JoinServer),
			}
			res, err := js.Get(ctx, tc.DeviceRequest)
			if tc.ErrorAssertion != nil {
				if !a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					log.Fatalf("Received unexpected error: %s", err)
				}
				a.So(res, should.BeNil)
				return
			}

			a.So(err, should.BeNil)
			a.So(res, should.Resemble, tc.DeviceResponse)
		})
	}
}
