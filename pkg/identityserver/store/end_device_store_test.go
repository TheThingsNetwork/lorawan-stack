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

package store

import (
	"testing"
	"time"

	ptypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestEndDeviceStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &EndDevice{}, &Attribute{}, &EndDeviceLocation{})
		store := GetEndDeviceStore(db)

		deviceID := ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
			JoinEUI:                &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			DevEUI:                 &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			DeviceID:               "foo",
		}

		deviceNewID := ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
			DeviceID:               "bar",
		}

		created, err := store.CreateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: deviceID,
			Name:                 "Foo EndDevice",
			Description:          "The Amazing Foo EndDevice",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
			Locations: map[string]*ttnpb.Location{
				"": {Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY},
			},
		})
		a.So(err, should.BeNil)
		a.So(created.DeviceID, should.Equal, deviceID.DeviceID)
		a.So(created.Name, should.Equal, "Foo EndDevice")
		a.So(created.Description, should.Equal, "The Amazing Foo EndDevice")
		a.So(created.Attributes, should.HaveLength, 3)
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		got, err := store.GetEndDevice(ctx,
			&deviceID,
			&ptypes.FieldMask{Paths: []string{"name", "attributes", "locations"}},
		)
		a.So(err, should.BeNil)
		a.So(got.DeviceID, should.Equal, deviceID.DeviceID)
		a.So(got.Name, should.Equal, "Foo EndDevice")
		a.So(got.Description, should.BeEmpty)
		a.So(got.Attributes, should.HaveLength, 3)
		if a.So(got.Locations, should.HaveLength, 1) {
			a.So(got.Locations[""], should.Resemble, &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY})
		}
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)

		_, err = store.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: deviceNewID,
		}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: deviceID,
			Name:                 "Foobar EndDevice",
			Description:          "The Amazing Foobar EndDevice",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
			Locations: map[string]*ttnpb.Location{
				"":    {Latitude: 12.3456, Longitude: 23.4567, Altitude: 1091, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY},
				"geo": {Latitude: 12.345, Longitude: 23.456, Accuracy: 500, Source: ttnpb.SOURCE_LORA_RSSI_GEOLOCATION},
			},
		}, &ptypes.FieldMask{Paths: []string{"description", "attributes", "locations"}})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar EndDevice")
		a.So(updated.Attributes, should.HaveLength, 3)
		if a.So(updated.Locations, should.HaveLength, 2) {
			a.So(updated.Locations[""], should.Resemble, &ttnpb.Location{Latitude: 12.3456, Longitude: 23.4567, Altitude: 1091, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY})
			a.So(updated.Locations["geo"], should.Resemble, &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Accuracy: 500, Source: ttnpb.SOURCE_LORA_RSSI_GEOLOCATION})
		}
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)

		got, err = store.GetEndDevice(ctx, &deviceID, nil)
		a.So(err, should.BeNil)
		a.So(got.DeviceID, should.Equal, created.DeviceID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.Attributes, should.Resemble, updated.Attributes)
		a.So(got.Locations, should.Resemble, updated.Locations)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)

		count, err := store.CountEndDevices(ctx, &deviceID.ApplicationIdentifiers)
		a.So(err, should.BeNil)
		a.So(count, should.Equal, 1)

		list, err := store.ListEndDevices(ctx,
			&deviceID.ApplicationIdentifiers,
			&ptypes.FieldMask{Paths: []string{"name"}},
		)
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		devices, err := store.FindEndDevices(ctx,
			[]*ttnpb.EndDeviceIdentifiers{&deviceID},
			&ptypes.FieldMask{Paths: []string{"name"}},
		)
		a.So(err, should.BeNil)
		if a.So(devices, should.HaveLength, 1) {
			a.So(devices[0].Name, should.EndWith, got.Name)
		}

		createdNew, err := store.CreateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: deviceNewID,
			Name:                 "Bar EndDevice",
			Description:          "The Amazing Bar EndDevice",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
			Locations: map[string]*ttnpb.Location{
				"": {Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY},
			},
		})
		a.So(err, should.BeNil)
		a.So(createdNew.DeviceID, should.Equal, deviceNewID.DeviceID)
		a.So(createdNew.Name, should.Equal, "Bar EndDevice")
		a.So(createdNew.Description, should.Equal, "The Amazing Bar EndDevice")
		a.So(createdNew.Attributes, should.HaveLength, 3)
		a.So(createdNew.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(createdNew.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		count, err = store.CountEndDevices(ctx, &deviceID.ApplicationIdentifiers)
		a.So(err, should.BeNil)
		a.So(count, should.Equal, 2)

		list, err = store.ListEndDevices(ctx,
			&deviceID.ApplicationIdentifiers,
			nil,
		)
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 2) {
			a.So(list, should.Contain, got)
			a.So(list, should.Contain, createdNew)
		}

		devices, err = store.FindEndDevices(ctx,
			[]*ttnpb.EndDeviceIdentifiers{&deviceID, &deviceNewID},
			nil,
		)
		a.So(err, should.BeNil)
		if a.So(devices, should.HaveLength, 2) {
			a.So(list, should.Contain, got)
			a.So(devices, should.Contain, createdNew)
		}

		err = store.DeleteEndDevice(ctx, &deviceID)
		a.So(err, should.BeNil)

		got, err = store.GetEndDevice(ctx, &deviceID, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		err = store.DeleteEndDevice(ctx, &deviceNewID)
		a.So(err, should.BeNil)

		got, err = store.GetEndDevice(ctx, &deviceNewID, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		count, err = store.CountEndDevices(ctx, &deviceID.ApplicationIdentifiers)
		a.So(err, should.BeNil)
		a.So(count, should.Equal, 0)

		list, err = store.ListEndDevices(ctx, &deviceID.ApplicationIdentifiers, nil)
		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)

		devices, err = store.FindEndDevices(ctx,
			[]*ttnpb.EndDeviceIdentifiers{&deviceID},
			nil)
		a.So(err, should.BeNil)
		a.So(devices, should.BeEmpty)

		devices, err = store.FindEndDevices(ctx,
			[]*ttnpb.EndDeviceIdentifiers{
				&deviceID,
				{
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test-another"},
					DeviceID:               "baz",
				},
			},
			nil)
		a.So(devices, should.BeNil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsInvalidArgument(err), should.BeTrue)
		}
	})
}
