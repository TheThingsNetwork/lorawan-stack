// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/networkserver"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func eui64Ptr(eui types.EUI64) *types.EUI64 { return &eui }

func TestNetworkServerCleanup(t *testing.T) {
	a, ctx := test.New(t)

	appList := []ttnpb.ApplicationIdentifiers{
		{ApplicationId: "app-1"},
		{ApplicationId: "app-2"},
		{ApplicationId: "app-3"},
		{ApplicationId: "app-4"},
	}
	deviceList := []*ttnpb.EndDevice{
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[0],
				DeviceId:       "dev-1",
				JoinEui:        eui64Ptr(types.EUI64{0x41, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x41, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[0],
				DeviceId:       "dev-2",
				JoinEui:        eui64Ptr(types.EUI64{0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x42, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[1],
				DeviceId:       "dev-3",
				JoinEui:        eui64Ptr(types.EUI64{0x43, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x43, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[2],
				DeviceId:       "dev-4",
				JoinEui:        eui64Ptr(types.EUI64{0x44, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x44, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[3],
				DeviceId:       "dev-5",
				JoinEui:        eui64Ptr(types.EUI64{0x45, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x45, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
		{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: &appList[3],
				DeviceId:       "dev-6",
				JoinEui:        eui64Ptr(types.EUI64{0x46, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
				DevEui:         eui64Ptr(types.EUI64{0x46, 0x42, 0xff, 0xff, 0xff, 0xff, 0xff, 0xff}),
			},
		},
	}

	deviceRegistry, closeFn := networkserver.NewDeviceRegistry(ctx)
	defer closeFn()

	for _, dev := range deviceList {
		ret, _, err := deviceRegistry.SetByID(ctx, dev.Ids.ApplicationIds, dev.Ids.DeviceId, []string{
			"ids.application_ids",
			"ids.dev_eui",
			"ids.device_id",
			"ids.join_eui",
		}, func(ctx context.Context, stored *ttnpb.EndDevice) (*ttnpb.EndDevice, []string, error) {
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

	isDeviceSet := map[string]struct{}{
		unique.ID(ctx, deviceList[1].Ids): {},
		unique.ID(ctx, deviceList[4].Ids): {},
		unique.ID(ctx, deviceList[5].Ids): {},
	}

	deviceRegistryCleaner := &networkserver.RegistryCleaner{
		DevRegistry: deviceRegistry,
	}
	err := deviceRegistryCleaner.RangeToLocalSet(ctx)
	a.So(err, should.BeNil)
	a.So(deviceRegistryCleaner.LocalSet, should.HaveLength, 6)
	err = deviceRegistryCleaner.CleanData(ctx, isDeviceSet)
	a.So(err, should.BeNil)
	deviceRegistryCleaner.RangeToLocalSet(ctx)
	a.So(deviceRegistryCleaner.LocalSet, should.HaveLength, 3)
	a.So(deviceRegistryCleaner.LocalSet, should.ContainKey, unique.ID(ctx, deviceList[1].Ids))
	a.So(deviceRegistryCleaner.LocalSet, should.ContainKey, unique.ID(ctx, deviceList[4].Ids))
	a.So(deviceRegistryCleaner.LocalSet, should.ContainKey, unique.ID(ctx, deviceList[5].Ids))
}
