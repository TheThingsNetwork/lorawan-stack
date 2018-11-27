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

package store

import (
	"testing"
	"time"

	"github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestEndDeviceStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		db.AutoMigrate(&EndDevice{}, &Attribute{}, &EndDeviceLocation{})
		store := GetEndDeviceStore(db)

		created, err := store.CreateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "test",
				},
				DeviceID: "foo",
			},
			Name:        "Foo EndDevice",
			Description: "The Amazing Foo EndDevice",
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
		a.So(created.DeviceID, should.Equal, "foo")
		a.So(created.Name, should.Equal, "Foo EndDevice")
		a.So(created.Description, should.Equal, "The Amazing Foo EndDevice")
		a.So(created.Attributes, should.HaveLength, 3)
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		got, err := store.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "foo",
		}, &types.FieldMask{Paths: []string{"name", "attributes", "locations"}})
		a.So(err, should.BeNil)
		a.So(got.DeviceID, should.Equal, "foo")
		a.So(got.Name, should.Equal, "Foo EndDevice")
		a.So(got.Description, should.BeEmpty)
		a.So(got.Attributes, should.HaveLength, 3)
		if a.So(got.Locations, should.HaveLength, 1) {
			a.So(got.Locations[""], should.Resemble, &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY})
		}
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)

		_, err = store.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "test",
				},
				DeviceID: "bar",
			},
		}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "test",
				},
				DeviceID: "foo",
			},
			Name:        "Foobar EndDevice",
			Description: "The Amazing Foobar EndDevice",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
			Locations: map[string]*ttnpb.Location{
				"":    {Latitude: 12.3456, Longitude: 23.4567, Altitude: 1091, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY},
				"geo": {Latitude: 12.345, Longitude: 23.456, Accuracy: 500, Source: ttnpb.SOURCE_LORA_RSSI_GEOLOCATION},
			},
		}, &types.FieldMask{Paths: []string{"description", "attributes", "locations"}})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar EndDevice")
		a.So(updated.Attributes, should.HaveLength, 3)
		if a.So(updated.Locations, should.HaveLength, 2) {
			a.So(updated.Locations[""], should.Resemble, &ttnpb.Location{Latitude: 12.3456, Longitude: 23.4567, Altitude: 1091, Accuracy: 1, Source: ttnpb.SOURCE_REGISTRY})
			a.So(updated.Locations["geo"], should.Resemble, &ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Accuracy: 500, Source: ttnpb.SOURCE_LORA_RSSI_GEOLOCATION})
		}
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)

		_, err = store.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "test",
				},
				DeviceID: "foo",
			},
			Description: "The Foobar EndDevice",
			UpdatedAt:   created.UpdatedAt,
		}, &types.FieldMask{Paths: []string{"description"}})
		a.So(err, should.NotBeNil)

		got, err = store.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "foo",
		}, nil)
		a.So(err, should.BeNil)
		a.So(got.DeviceID, should.Equal, created.DeviceID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.Attributes, should.Resemble, updated.Attributes)
		a.So(got.Locations, should.Resemble, updated.Locations)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)

		list, err := store.ListEndDevices(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationID: "test",
		}, &types.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		err = store.DeleteEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "foo",
		})
		a.So(err, should.BeNil)

		got, err = store.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "foo",
		}, nil)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.ListEndDevices(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationID: "test",
		}, nil)
		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)

	})
}
