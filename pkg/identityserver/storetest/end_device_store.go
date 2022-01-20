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

package storetest

import (
	. "testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestEndDeviceStoreCRUD(t *T) {
	a, ctx := test.New(t)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ApplicationStore
		is.EndDeviceStore
	})
	defer st.DestroyDB(t, true, "applications", "attributes", "end_device_locations", "pictures") // TODO: Make sure (at least) attributes and end_device_locations are deleted when deleting end devices.
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement ApplicationStore and EndDeviceStore")
	}

	application, err := s.CreateApplication(ctx, &ttnpb.Application{
		Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	mask := fieldMask(
		"name", "description", "attributes", "version_ids",
		"network_server_address", "application_server_address", "join_server_address",
		"service_profile_id", "locations", "picture", "activated_at",
	)

	location := &ttnpb.Location{
		Latitude:  12.34,
		Longitude: 56.78,
		Altitude:  42,
		Accuracy:  2,
		Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
	}
	wifiLocation := &ttnpb.Location{
		Latitude:  12.34,
		Longitude: 56.78,
		Altitude:  42,
		Accuracy:  50,
		Source:    ttnpb.LocationSource_SOURCE_WIFI_RSSI_GEOLOCATION,
	}
	picture := &ttnpb.Picture{
		Embedded: &ttnpb.Picture_Embedded{
			MimeType: "image/png",
			Data:     []byte("foobarbaz"),
		},
	}
	var created *ttnpb.EndDevice

	t.Run("CreateEndDevice", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)
		stamp := start.Add(-1 * time.Minute)

		created, err = s.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "foo",
				JoinEui:        &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
				DevEui:         &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			},
			Name:        "Foo Name",
			Description: "Foo Description",
			Attributes:  attributes,
			VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
				BrandId:         "some_brand_id",
				ModelId:         "some_model_id",
				HardwareVersion: "hw_v3",
				FirmwareVersion: "fw_v3",
				BandId:          "some_band_id",
			},
			NetworkServerAddress:     "ns.example.com",
			ApplicationServerAddress: "as.example.com",
			JoinServerAddress:        "js.example.com",
			ServiceProfileId:         "some_profile_id",
			Locations: map[string]*ttnpb.Location{
				"":     location,
				"wifi": wifiLocation,
			},
			Picture:     picture,
			ActivatedAt: ttnpb.ProtoTimePtr(stamp),
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetApplicationIds(), should.Resemble, application.GetIds())
			a.So(created.GetIds().GetDeviceId(), should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Name")
			a.So(created.Description, should.Equal, "Foo Description")
			a.So(created.Attributes, should.Resemble, attributes)
			a.So(created.VersionIds, should.Resemble, &ttnpb.EndDeviceVersionIdentifiers{
				BrandId:         "some_brand_id",
				ModelId:         "some_model_id",
				HardwareVersion: "hw_v3",
				FirmwareVersion: "fw_v3",
				BandId:          "some_band_id",
			})
			a.So(created.NetworkServerAddress, should.Equal, "ns.example.com")
			a.So(created.ApplicationServerAddress, should.Equal, "as.example.com")
			a.So(created.JoinServerAddress, should.Equal, "js.example.com")
			a.So(created.ServiceProfileId, should.Equal, "some_profile_id")
			a.So(created.Locations, should.Resemble, map[string]*ttnpb.Location{
				"":     location,
				"wifi": wifiLocation,
			})
			a.So(created.Picture, should.Resemble, picture)
			a.So(*ttnpb.StdTime(created.ActivatedAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("CreateEndDevice_AfterCreate", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "foo",
			},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsAlreadyExists(err), should.BeTrue)
		}
	})

	t.Run("GetEndDevice", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "foo",
		}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetEndDevice_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "other",
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
		// 	ApplicationIds: application.GetIds(),
		// 	DeviceId:       "",
		// }, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("CountEndDevices", func(t *T) {
		a, ctx := test.New(t)
		count, err := s.CountEndDevices(ctx, nil)
		if a.So(err, should.BeNil) {
			a.So(count, should.Equal, 1)
		}
		count, err = s.CountEndDevices(ctx, application.GetIds())
		if a.So(err, should.BeNil) {
			a.So(count, should.Equal, 1)
		}
	})

	t.Run("CountEndDevices_Other", func(t *T) {
		a, ctx := test.New(t)
		count, err := s.CountEndDevices(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationId: "other",
		})
		if a.So(err, should.BeNil) {
			a.So(count, should.Equal, 0)
		}
	})

	t.Run("ListEndDevices", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListEndDevices(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
		got, err = s.ListEndDevices(ctx, application.GetIds(), mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	t.Run("ListEndDevices_Other", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListEndDevices(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationId: "other",
		}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("FindEndDevices", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindEndDevices(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	updatedLocation := &ttnpb.Location{
		Latitude:  12.34,
		Longitude: 56.78,
		Altitude:  42,
		Accuracy:  1,
		Source:    ttnpb.LocationSource_SOURCE_REGISTRY,
	}
	extraLocation := &ttnpb.Location{
		Latitude:  12.34,
		Longitude: 56.78,
		Altitude:  30,
		Accuracy:  5,
		Source:    ttnpb.LocationSource_SOURCE_COMBINED_GEOLOCATION,
	}
	updatedPicture := &ttnpb.Picture{
		Sizes: map[uint32]string{0: "https://example.com/device_picture.jpg"},
	}
	var updated *ttnpb.EndDevice

	t.Run("UpdateEndDevice", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)
		stamp := start.Add(time.Minute)

		updated, err = s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "foo",
			},
			Name:        "New Foo Name",
			Description: "New Foo Description",
			Attributes:  updatedAttributes,
			VersionIds: &ttnpb.EndDeviceVersionIdentifiers{
				BrandId:         "other_brand_id",
				ModelId:         "other_model_id",
				HardwareVersion: "hw_v3.1",
				FirmwareVersion: "fw_v3.1",
				BandId:          "other_band_id",
			},
			NetworkServerAddress:     "other-ns.example.com",
			ApplicationServerAddress: "other-as.example.com",
			JoinServerAddress:        "other-js.example.com",
			ServiceProfileId:         "other_profile_id",
			Locations: map[string]*ttnpb.Location{
				"":    updatedLocation,
				"geo": extraLocation,
			},
			Picture:     updatedPicture,
			ActivatedAt: ttnpb.ProtoTimePtr(stamp),
		}, mask)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.GetIds().GetDeviceId(), should.Equal, "foo")
			a.So(updated.Name, should.Equal, "New Foo Name")
			a.So(updated.Description, should.Equal, "New Foo Description")
			a.So(updated.Attributes, should.Resemble, updatedAttributes)
			a.So(updated.VersionIds, should.Resemble, &ttnpb.EndDeviceVersionIdentifiers{
				BrandId:         "other_brand_id",
				ModelId:         "other_model_id",
				HardwareVersion: "hw_v3.1",
				FirmwareVersion: "fw_v3.1",
				BandId:          "other_band_id",
			})
			a.So(updated.NetworkServerAddress, should.Equal, "other-ns.example.com")
			a.So(updated.ApplicationServerAddress, should.Equal, "other-as.example.com")
			a.So(updated.JoinServerAddress, should.Equal, "other-js.example.com")
			a.So(updated.ServiceProfileId, should.Equal, "other_profile_id")
			a.So(updated.Locations, should.Resemble, map[string]*ttnpb.Location{
				"":    updatedLocation,
				"geo": extraLocation,
			})
			a.So(updated.Picture, should.Resemble, updatedPicture)
			a.So(*ttnpb.StdTime(updated.ActivatedAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
		}
	})

	t.Run("UpdateEndDevice_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "other",
			},
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// _, err = s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
		// 	Ids: &ttnpb.EndDeviceIdentifiers{
		// 		ApplicationIds: application.GetIds(),
		// 		DeviceId:       "",
		// 	},
		// }, mask)
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetEndDevice_AfterUpdate", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "foo",
		}, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, updated)
		}
	})

	t.Run("DeleteEndDevice", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "foo",
		})
		a.So(err, should.BeNil)
	})

	t.Run("DeleteEndDevice_Other", func(t *T) {
		a, ctx := test.New(t)
		err := s.DeleteEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "other",
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
		// TODO: Enable test (https://github.com/TheThingsIndustries/lorawan-stack/issues/3034).
		// err = s.DeleteEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
		// 	ApplicationIds: application.GetIds(),
		// 	DeviceId:       "",
		// })
		// if a.So(err, should.NotBeNil) {
		// 	a.So(errors.IsNotFound(err), should.BeTrue)
		// }
	})

	t.Run("GetEndDevice_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "foo",
		}, mask)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})

	t.Run("CountEndDevices", func(t *T) {
		a, ctx := test.New(t)
		count, err := s.CountEndDevices(ctx, nil)
		if a.So(err, should.BeNil) {
			a.So(count, should.Equal, 0)
		}
		count, err = s.CountEndDevices(ctx, application.GetIds())
		if a.So(err, should.BeNil) {
			a.So(count, should.Equal, 0)
		}
	})

	t.Run("ListEndDevices", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListEndDevices(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
		got, err = s.ListEndDevices(ctx, application.GetIds(), mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("FindEndDevices_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindEndDevices(ctx, nil, mask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})
}
