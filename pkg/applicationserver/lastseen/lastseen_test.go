// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package lastseen

import (
	"testing"
	"time"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestLastSeenWrite(t *testing.T) {
	ctx := test.Context()
	a := assertions.New(t)
	registeredApplicationID := ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}
	// This device gets registered in the device registry of the Application Server.
	testDevice1 := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &registeredApplicationID,
			DeviceId:       "foo-device-1",
		},
	}

	testDevice2 := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &registeredApplicationID,
			DeviceId:       "foo-device-2",
		},
	}

	testDevice3 := &ttnpb.EndDevice{
		Ids: &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: &registeredApplicationID,
			DeviceId:       "foo-device-3",
		},
	}
	lsProvider := NewBatchLastSeenProvider(ctx, 2, nil, nil)

	currentTime := time.Now()
	devs, err := lsProvider.push(unique.ID(ctx, testDevice1.Ids), ttnpb.ProtoTime(&currentTime))
	a.So(err, should.BeNil)
	a.So(devs, should.HaveLength, 0)

	devs, err = lsProvider.push(unique.ID(ctx, testDevice2.Ids), ttnpb.ProtoTime(&currentTime))
	a.So(err, should.BeNil)
	a.So(devs, should.HaveLength, 0)

	invalidTime := time.Now().Add(-1 * time.Minute)
	devs, err = lsProvider.push(unique.ID(ctx, testDevice1.Ids), ttnpb.ProtoTime(&invalidTime))
	a.So(err, should.BeNil)
	a.So(devs, should.HaveLength, 0)

	validTime := time.Now()
	devs, err = lsProvider.push(unique.ID(ctx, testDevice2.Ids), ttnpb.ProtoTime(&validTime))
	a.So(err, should.BeNil)
	a.So(devs, should.HaveLength, 0)

	devs, err = lsProvider.push(unique.ID(ctx, testDevice3.Ids), ttnpb.ProtoTime(&validTime))
	a.So(err, should.BeNil)
	a.So(devs, should.HaveLength, 2)

	secondBatch, err := lsProvider.clear()
	a.So(err, should.BeNil)
	a.So(secondBatch, should.HaveLength, 1)
	a.So(secondBatch[unique.ID(ctx, testDevice3.Ids)].Ids.DeviceId, should.Equal, testDevice3.Ids.DeviceId)
	a.So(secondBatch[unique.ID(ctx, testDevice3.Ids)].LastSeenAt, should.Resemble, ttnpb.ProtoTime(&validTime))

	for _, dev := range devs {
		if dev.Ids.DeviceId == testDevice1.Ids.DeviceId {
			a.So(dev.LastSeenAt, should.Resemble, ttnpb.ProtoTime(&currentTime))
		} else if dev.Ids.DeviceId == testDevice2.Ids.DeviceId {
			a.So(dev.LastSeenAt, should.Resemble, ttnpb.ProtoTime(&validTime))
		} else {
			t.Error("Unknown device in the map.")
		}
	}
}
