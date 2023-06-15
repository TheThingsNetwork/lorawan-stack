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
	"context"
	. "testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/protobuf/types/known/timestamppb"
)

var endDeviceMask = fieldMask(
	"name", "description", "attributes", "version_ids",
	"network_server_address", "application_server_address", "join_server_address",
	"service_profile_id", "locations", "picture", "activated_at", "last_seen_at",
	"claim_authentication_code", "serial_number",
	"lora_alliance_profile_ids",
)

func (st *StoreTest) TestEndDeviceStoreCRUD(t *T) {
	a, ctx := test.New(t)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ApplicationStore
		is.EndDeviceStore
	})
	defer st.DestroyDB(
		t, true,
		"applications", "attributes", "end_device_locations", "pictures",
	) // TODO: Make sure (at least) attributes and end_device_locations are deleted when deleting end devices.
	if !ok {
		t.Skip("Store does not implement ApplicationStore and EndDeviceStore")
	}
	defer s.Close()

	application, err := s.CreateApplication(ctx, &ttnpb.Application{
		Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

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
	createdLoRaIds := &ttnpb.LoRaAllianceProfileIdentifiers{
		VendorId:        1,
		VendorProfileId: 1,
	}
	start := time.Now().Truncate(time.Second)
	claim := &ttnpb.EndDeviceAuthenticationCode{
		ValidFrom: timestamppb.New(start),
		Value:     "secret",
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
				JoinEui:        types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
				DevEui:         types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
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
			SerialNumber:             "YYWWNNNNN1",
			ServiceProfileId:         "some_profile_id",
			LoraAllianceProfileIds:   createdLoRaIds,
			Locations: map[string]*ttnpb.Location{
				"":     location,
				"wifi": wifiLocation,
			},
			Picture:                 picture,
			ActivatedAt:             timestamppb.New(stamp),
			ClaimAuthenticationCode: claim,
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
			a.So(created.SerialNumber, should.Equal, "YYWWNNNNN1")
			a.So(created.ServiceProfileId, should.Equal, "some_profile_id")
			a.So(created.LoraAllianceProfileIds, should.Resemble, createdLoRaIds)
			a.So(created.Locations, should.Resemble, map[string]*ttnpb.Location{
				"":     location,
				"wifi": wifiLocation,
			})
			a.So(created.Picture, should.Resemble, picture)
			a.So(created.ClaimAuthenticationCode, should.Resemble, claim)
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
		}, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetEndDevice_ByEUIs", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			DevEui:  types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
			JoinEui: types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
		}, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.Resemble, created)
		}
	})

	t.Run("GetEndDevice_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.GetEndDevice(ctx, &ttnpb.EndDeviceIdentifiers{
			ApplicationIds: application.GetIds(),
			DeviceId:       "other",
		}, endDeviceMask)
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
		got, err := s.ListEndDevices(ctx, nil, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
		got, err = s.ListEndDevices(ctx, application.GetIds(), endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 1) {
			a.So(got[0], should.Resemble, created)
		}
	})

	t.Run("ListEndDevices_Other", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.ListEndDevices(ctx, &ttnpb.ApplicationIdentifiers{
			ApplicationId: "other",
		}, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("FindEndDevices", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindEndDevices(ctx, nil, endDeviceMask)
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
	updatedLoRaIds := &ttnpb.LoRaAllianceProfileIdentifiers{
		VendorId:        2,
		VendorProfileId: 2,
	}
	updatedCAC := &ttnpb.EndDeviceAuthenticationCode{
		ValidFrom: timestamppb.New(start),
		ValidTo:   timestamppb.New(start.Add(time.Hour)),
		Value:     "other secret",
	}
	var updated *ttnpb.EndDevice

	t.Run("UpdateEndDevice", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)
		stamp := start.Add(time.Minute)

		in := &ttnpb.EndDevice{
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
			SerialNumber:             "YYWWNNNNN2",
			ServiceProfileId:         "other_profile_id",
			LoraAllianceProfileIds:   updatedLoRaIds,
			Locations: map[string]*ttnpb.Location{
				"":    updatedLocation,
				"geo": extraLocation,
			},
			Picture:                 updatedPicture,
			ActivatedAt:             timestamppb.New(stamp),
			LastSeenAt:              timestamppb.New(stamp),
			ClaimAuthenticationCode: updatedCAC,
		}

		updated, err = s.UpdateEndDevice(ctx, in, endDeviceMask)
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
			a.So(updated.SerialNumber, should.Equal, "YYWWNNNNN2")
			a.So(updated.ServiceProfileId, should.Equal, "other_profile_id")
			a.So(updated.LoraAllianceProfileIds, should.Resemble, updatedLoRaIds)
			a.So(updated.Locations, should.Resemble, map[string]*ttnpb.Location{
				"":    updatedLocation,
				"geo": extraLocation,
			})
			a.So(updated.Picture, should.Resemble, updatedPicture)
			a.So(*ttnpb.StdTime(updated.ActivatedAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(updated.LastSeenAt), should.Equal, stamp)
			a.So(*ttnpb.StdTime(updated.CreatedAt), should.Equal, *ttnpb.StdTime(created.CreatedAt))
			a.So(*ttnpb.StdTime(updated.UpdatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(updated.ClaimAuthenticationCode, should.Resemble, updatedCAC)
		}
	})

	t.Run("UpdateEndDevice_Other", func(t *T) {
		a, ctx := test.New(t)
		_, err := s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "other",
			},
		}, endDeviceMask)
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
		}, endDeviceMask)
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
		}, endDeviceMask)
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
		got, err := s.ListEndDevices(ctx, nil, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
		got, err = s.ListEndDevices(ctx, application.GetIds(), endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})

	t.Run("FindEndDevices_AfterDelete", func(t *T) {
		a, ctx := test.New(t)
		got, err := s.FindEndDevices(ctx, nil, endDeviceMask)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.BeEmpty)
		}
	})
}

func (st *StoreTest) TestEndDeviceStorePagination(t *T) {
	usr1 := st.population.NewUser()
	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	var all []*ttnpb.EndDevice
	for i := 0; i < 7; i++ {
		all = append(all, st.population.NewEndDevice(app1.GetIds()))
	}

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.EndDeviceStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement EndDeviceStore")
	}
	defer s.Close()

	t.Run("ListEndDevices_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.ListEndDevices(paginateCtx, app1.GetIds(), endDeviceMask)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				for i, e := range got {
					a.So(e, should.Resemble, all[i+2*int(page-1)])
				}
			}

			a.So(total, should.Equal, 7)
		}
	})
}

func (st *StoreTest) TestEndDeviceBatchUpdate(t *T) {
	usr1 := st.population.NewUser()
	app1 := st.population.NewApplication(usr1.GetOrganizationOrUserIdentifiers())

	var all []*ttnpb.EndDevice
	for i := 0; i < 3; i++ {
		all = append(all, st.population.NewEndDevice(app1.GetIds()))
	}

	dev1 := all[0]
	dev2 := all[1]
	dev3 := all[2]

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.EndDeviceStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement EndDeviceStore")
	}
	defer s.Close()

	t.Run("BatchUpdateEndDeviceLastSeen", func(t *T) {
		a, ctx := test.New(t)

		validDevTime := time.Now().Truncate(time.Millisecond)
		dev1.LastSeenAt = timestamppb.New(validDevTime)
		dev2.LastSeenAt = timestamppb.New(validDevTime)
		dev3.LastSeenAt = timestamppb.New(validDevTime.Add(10 * time.Second))

		batch := []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate{
			{Ids: dev1.Ids, LastSeenAt: dev1.LastSeenAt},
			{Ids: dev2.Ids, LastSeenAt: dev2.LastSeenAt},
			{Ids: dev3.Ids, LastSeenAt: dev3.LastSeenAt},
		}

		err := s.BatchUpdateEndDeviceLastSeen(ctx, batch)
		a.So(err, should.BeNil)

		got, err := s.ListEndDevices(ctx, app1.GetIds(), []string{"last_seen_at"})
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 3) {
			for _, dev := range got {
				a.So(dev.LastSeenAt, should.NotBeNil)
				if dev.Ids.DeviceId == dev1.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime))
				} else if dev.Ids.DeviceId == dev2.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime))
				} else if dev.Ids.DeviceId == dev3.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime.Add(10*time.Second)))
				}
			}
		}

		invalidDev1Time := timestamppb.New(time.Now().Add(-10 * time.Minute).Truncate(time.Millisecond))
		invalidDev2Time := timestamppb.New(time.Now().Add(-5 * time.Minute).Truncate(time.Millisecond))
		invalidDev3Time := timestamppb.New(time.Now().Add(-1 * time.Minute).Truncate(time.Millisecond))
		dev1.LastSeenAt = invalidDev1Time
		dev2.LastSeenAt = invalidDev2Time
		dev3.LastSeenAt = invalidDev3Time

		batch = []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate{
			{Ids: dev1.Ids, LastSeenAt: dev1.LastSeenAt},
			{Ids: dev2.Ids, LastSeenAt: dev2.LastSeenAt},
			{Ids: dev3.Ids, LastSeenAt: dev3.LastSeenAt},
		}

		err = s.BatchUpdateEndDeviceLastSeen(ctx, batch)
		a.So(err, should.BeNil)
		got, err = s.ListEndDevices(ctx, app1.GetIds(), []string{"last_seen_at"})

		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) && a.So(got, should.HaveLength, 3) {
			for _, dev := range got {
				a.So(dev.LastSeenAt, should.NotBeNil)
				if dev.Ids.DeviceId == dev1.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime))
				} else if dev.Ids.DeviceId == dev2.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime))
				} else if dev.Ids.DeviceId == dev3.Ids.DeviceId {
					a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime.Add(10*time.Second)))
				}
			}
		}

		// Test duplicates in batch update call.
		batch = []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate{
			{Ids: dev1.Ids, LastSeenAt: dev1.LastSeenAt},
			{Ids: dev1.Ids, LastSeenAt: timestamppb.New(validDevTime.Add(10 * time.Second))},
		}

		err = s.BatchUpdateEndDeviceLastSeen(ctx, batch)
		a.So(err, should.BeNil)

		dev, err := s.GetEndDevice(ctx, dev1.Ids, []string{"last_seen_at"})
		if a.So(err, should.BeNil) && a.So(dev, should.NotBeNil) && a.So(dev.LastSeenAt, should.NotBeNil) {
			a.So(dev.LastSeenAt, should.Resemble, timestamppb.New(validDevTime.Add(10*time.Second)))
		}
	})
}

func (st *StoreTest) TestEndDeviceCAC(t *T) { //nolint:revive
	a, ctx := test.New(t)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.ApplicationStore
		is.EndDeviceStore
	})
	defer st.DestroyDB(
		t,
		true,
		"applications",
		"attributes",
		"end_device_locations",
		"pictures",
	)
	if !ok {
		t.Skip("Store does not implement ApplicationStore and EndDeviceStore")
	}
	defer s.Close()

	application, err := s.CreateApplication(ctx, &ttnpb.Application{
		Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
	})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	var created *ttnpb.EndDevice

	t.Run("CreateAndGetEndDeviceWithoutCAC", func(t *T) {
		a, ctx := test.New(t)
		var err error
		start := time.Now().Truncate(time.Second)

		created, err = s.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "foo",
				JoinEui:        types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
				DevEui:         types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}.Bytes(),
			},
		})

		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetApplicationIds(), should.Resemble, application.GetIds())
			a.So(created.GetIds().GetDeviceId(), should.Equal, "foo")
			a.So(*ttnpb.StdTime(created.CreatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(*ttnpb.StdTime(created.UpdatedAt), should.HappenWithin, 5*time.Second, start)
			a.So(created.ClaimAuthenticationCode, should.BeNil)
		}

		got, err := s.GetEndDevice(ctx, created.GetIds(), []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.ClaimAuthenticationCode, should.BeNil)
		}

		// Update the CAC value.
		updated, err := s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: created.GetIds(),
			ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
				Value: "bar",
			},
		}, []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.ClaimAuthenticationCode, should.Resemble, &ttnpb.EndDeviceAuthenticationCode{
				Value: "bar",
			})
		}

		// Update CAC validity fields individually.
		// Truncate to avoid nanosecond precision issues.
		now := time.Unix(time.Now().Unix(), 0)

		validFrom := now.Add(-1 * time.Hour)
		updated, err = s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: created.GetIds(),
			ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
				ValidFrom: ttnpb.ProtoTime(&validFrom),
			},
		}, []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.ClaimAuthenticationCode, should.Resemble, &ttnpb.EndDeviceAuthenticationCode{
				Value:     "bar",
				ValidFrom: ttnpb.ProtoTime(&validFrom),
			})
		}

		validTo := now.Add(-1 * time.Hour)
		updated, err = s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: created.GetIds(),
			ClaimAuthenticationCode: &ttnpb.EndDeviceAuthenticationCode{
				ValidTo: ttnpb.ProtoTime(&validTo),
			},
		}, []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.ClaimAuthenticationCode, should.Resemble, &ttnpb.EndDeviceAuthenticationCode{
				Value:     "bar",
				ValidFrom: ttnpb.ProtoTime(&validFrom),
				ValidTo:   ttnpb.ProtoTime(&validTo),
			})
		}

		// Clear the CAC value.
		updated, err = s.UpdateEndDevice(ctx, &ttnpb.EndDevice{
			Ids:                     created.GetIds(),
			ClaimAuthenticationCode: nil,
		}, []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.ClaimAuthenticationCode, should.BeNil)
		}

		got, err = s.GetEndDevice(ctx, created.GetIds(), []string{"claim_authentication_code"})
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.ClaimAuthenticationCode, should.BeNil)
		}
	})
}

// TestEndDeviceBatchOperations tests the EndDeviceBatchStore implementation.
func (st *StoreTest) TestEndDeviceBatchOperations(t *T) { // nolint:gocyclo
	a, ctx := test.New(t)

	s, ok := st.PrepareDB(t).(interface {
		Store
		is.EndDeviceStore
		is.ApplicationStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement TestEndDeviceBatchOperations")
	}
	defer s.Close()

	for _, ctx := range []context.Context{
		ctx,
	} {
		application, err := s.CreateApplication(ctx, &ttnpb.Application{
			Ids: &ttnpb.ApplicationIdentifiers{ApplicationId: "foo"},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		dev1, err := s.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "foo-1",
			},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		dev2, err := s.CreateEndDevice(ctx, &ttnpb.EndDevice{
			Ids: &ttnpb.EndDeviceIdentifiers{
				ApplicationIds: application.GetIds(),
				DeviceId:       "bar-1",
			},
		})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}

		dev1.LastSeenAt = timestamppb.New(time.Now().Truncate(time.Millisecond))
		dev2.LastSeenAt = timestamppb.New(time.Now().Add(-1 * time.Second).Truncate(time.Millisecond))

		batch := []*ttnpb.BatchUpdateEndDeviceLastSeenRequest_EndDeviceLastSeenUpdate{
			{Ids: dev1.Ids, LastSeenAt: dev1.LastSeenAt},
			{Ids: dev2.Ids, LastSeenAt: dev2.LastSeenAt},
		}
		err = s.BatchUpdateEndDeviceLastSeen(ctx, batch)
		a.So(err, should.BeNil)

		devs, err := s.ListEndDevices(ctx, application.Ids, []string{"last_seen_at"})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(devs, should.HaveLength, 2)

		for _, dev := range devs {
			if dev.Ids.DeviceId == dev1.Ids.DeviceId {
				a.So(dev.LastSeenAt, should.Resemble, dev1.LastSeenAt)
			} else if dev.Ids.DeviceId == dev2.Ids.DeviceId {
				a.So(dev.LastSeenAt, should.Resemble, dev2.LastSeenAt)
			} else {
				t.FailNow()
			}
		}

		// Batch Delete
		for _, tc := range []struct { // nolint:paralleltest
			Name            string
			Context         context.Context
			BatchDeleteFunc func(context.Context, []*ttnpb.EndDeviceIdentifiers) ([]*ttnpb.EndDeviceIdentifiers, error)
			ApplicationIDs  *ttnpb.ApplicationIdentifiers
			DeviceIDs       []string
			Response        []*ttnpb.EndDeviceIdentifiers
			ErrorAssertion  func(*T, error) bool
		}{
			{
				Name:           "Not Found",
				Context:        ctx,
				ApplicationIDs: application.Ids,
				DeviceIDs: []string{
					"unknown",
				},
				Response: []*ttnpb.EndDeviceIdentifiers{},
			},
			{
				Name:           "Valid Batch",
				Context:        ctx,
				ApplicationIDs: application.Ids,
				DeviceIDs: []string{
					dev1.Ids.DeviceId,
					dev2.Ids.DeviceId,
				},
				Response: []*ttnpb.EndDeviceIdentifiers{
					dev2.Ids,
					dev1.Ids,
				},
			},
		} {
			tc := tc
			t.Run(tc.Name, func(t *T) {
				a := assertions.New(t)
				deleted, err := s.BatchDeleteEndDevices(tc.Context, tc.ApplicationIDs, tc.DeviceIDs)
				if tc.ErrorAssertion != nil && a.So(tc.ErrorAssertion(t, err), should.BeTrue) {
					a.So(deleted, should.BeNil)
				} else if a.So(err, should.BeNil) {
					if tc.Response != nil {
						a.So(deleted, should.Resemble, tc.Response)
					} else {
						a.So(deleted, should.BeNil)
					}
				}
			})
		}

		// Check that the devices are deleted.
		devs, err = s.ListEndDevices(ctx, application.Ids, []string{"last_seen_at", "locations"})
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		a.So(devs, should.HaveLength, 0)
	}
}
