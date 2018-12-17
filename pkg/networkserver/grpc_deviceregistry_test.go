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

package networkserver_test

import (
	"context"
	"log"
	"testing"

	"github.com/smartystreets/assertions"
	clusterauth "go.thethings.network/lorawan-stack/pkg/auth/cluster"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

// MockDeviceRegistry implemets DeviceRegistry
type MockDeviceRegistry struct {
	GetByEUIFunc    func(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error)
	GetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error)
	RangeByAddrFunc func(devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error
	SetByIDFunc     func(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error)
}

func (r *MockDeviceRegistry) GetByEUI(ctx context.Context, joinEUI types.EUI64, devEUI types.EUI64, paths []string) (*ttnpb.EndDevice, error) {
	if r.GetByEUIFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.GetByEUIFunc(ctx, joinEUI, devEUI, paths)
}

func (r *MockDeviceRegistry) GetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string) (*ttnpb.EndDevice, error) {
	if r.SetByIDFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.GetByIDFunc(ctx, appID, devID, paths)
}

func (r *MockDeviceRegistry) SetByID(ctx context.Context, appID ttnpb.ApplicationIdentifiers, devID string, paths []string, f func(*ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error)) (*ttnpb.EndDevice, error) {
	if r.SetByIDFunc == nil {
		return nil, errors.New("Not implemented")
	}
	return r.SetByIDFunc(ctx, appID, devID, paths, f)
}

func (r *MockDeviceRegistry) RangeByAddr(devAddr types.DevAddr, paths []string, f func(*ttnpb.EndDevice) bool) error {
	if r.SetByIDFunc == nil {
		return errors.New("Not implemented")
	}
	return r.RangeByAddrFunc(devAddr, paths, f)
}

func TestDeviceRegistry(t *testing.T) {
	for _, tc := range []struct {
		Name string

		Context func(context.Context) context.Context

		GetByEUI func(context.Context, types.EUI64, types.EUI64, []string) (*ttnpb.EndDevice, error)

		JoinEUI types.EUI64
		DevEUI  types.EUI64

		DeviceResponse *ttnpb.EndDevice

		ErrorAssertion func(*testing.T, error) bool
	}{
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

			JoinEUI: types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},
			DevEUI:  types.EUI64{0x43, 0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff},

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

			reg := &MockDeviceRegistry{GetByEUIFunc: tc.GetByEUI}

			res, err := reg.GetByEUI(ctx, tc.JoinEUI, tc.DevEUI, nil)

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
