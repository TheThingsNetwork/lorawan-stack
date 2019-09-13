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

package identityserver

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"google.golang.org/grpc"
)

func TestEndDevicesPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewEndDeviceRegistryClient(cc)

		joinEUI := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := types.EUI64{8, 7, 6, 5, 4, 3, 2, 1}

		_, err := reg.Create(ctx, &ttnpb.CreateEndDeviceRequest{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "test-device-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "test-app-id",
					},
				},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetEndDeviceRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID: "test-device-id",
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
					ApplicationID: "test-app-id",
				},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.GetIdentifiersForEUIs(ctx, &ttnpb.GetEndDeviceIdentifiersForEUIsRequest{
			JoinEUI: joinEUI,
			DevEUI:  devEUI,
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		_, err = reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app-id",
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID: "test-device-id",
					ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
						ApplicationID: "test-app-id",
					},
				},
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.EndDeviceIdentifiers{
			DeviceID: "test-device-id",
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test-app-id",
			},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestEndDevicesCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewEndDeviceRegistryClient(cc)

		userID := defaultUser.UserIdentifiers
		creds := userCreds(defaultUserIdx)
		app := userApplications(&userID).Applications[0]

		joinEUI := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		devEUI := types.EUI64{8, 7, 6, 5, 4, 3, 2, 1}

		start := time.Now()

		created, err := reg.Create(ctx, &ttnpb.CreateEndDeviceRequest{
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-device-id",
					ApplicationIdentifiers: app.ApplicationIdentifiers,
					JoinEUI:                &joinEUI,
					DevEUI:                 &devEUI,
				},
				Name: "test-device-name",
			},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.CreatedAt, should.HappenAfter, start)
			a.So(created.UpdatedAt, should.HappenAfter, start)
			a.So(created.Name, should.Equal, "test-device-name")
		}

		got, err := reg.Get(ctx, &ttnpb.GetEndDeviceRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-device-id",
				ApplicationIdentifiers: app.ApplicationIdentifiers,
			},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, "test-device-name")
		}

		ids, err := reg.GetIdentifiersForEUIs(ctx, &ttnpb.GetEndDeviceIdentifiersForEUIsRequest{
			JoinEUI: joinEUI,
			DevEUI:  devEUI,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(ids, should.NotBeNil) {
			a.So(*ids, should.Resemble, created.EndDeviceIdentifiers)
		}

		list, err := reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			FieldMask:              pbtypes.FieldMask{Paths: []string{"name"}},
			ApplicationIdentifiers: app.ApplicationIdentifiers,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) && a.So(list.EndDevices, should.HaveLength, 1) {
			if a.So(list.EndDevices[0], should.NotBeNil) {
				a.So(list.EndDevices[0].Name, should.Equal, "test-device-name")
			}
		}

		start = time.Now()

		updated, err := reg.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			EndDevice: ttnpb.EndDevice{
				EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
					DeviceID:               "test-device-id",
					ApplicationIdentifiers: app.ApplicationIdentifiers,
				},
				Name: "test-device-name-new",
			},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "test-device-name-new")
			a.So(updated.UpdatedAt, should.HappenAfter, start)
		}

		_, err = reg.Delete(ctx, &ttnpb.EndDeviceIdentifiers{
			DeviceID:               "test-device-id",
			ApplicationIdentifiers: app.ApplicationIdentifiers,
		}, creds)

		a.So(err, should.BeNil)

		_, err = reg.Get(ctx, &ttnpb.GetEndDeviceRequest{
			FieldMask: pbtypes.FieldMask{Paths: []string{"name"}},
			EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
				DeviceID:               "test-device-id",
				ApplicationIdentifiers: app.ApplicationIdentifiers,
			},
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}
	})
}
