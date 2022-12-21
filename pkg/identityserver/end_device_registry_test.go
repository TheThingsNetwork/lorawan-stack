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

package identityserver

import (
	"testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

func TestEndDevicesPermissionDenied(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	dev1 := p.NewEndDevice(app1.GetIds())

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewEndDeviceRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateEndDeviceRequest{
			EndDevice: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: app1.GetIds(),
					DeviceId:       "foo-dev",
				},
			},
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIds: dev1.GetIds(),
			FieldMask:    ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: app1.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			EndDevice: &ttnpb.EndDevice{
				Ids:  dev1.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		})
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, dev1.GetIds())
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	}, withPrivateTestDatabase(p))
}

func TestEndDevicesCRUD(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	for i := 0; i < 5; i++ {
		p.NewEndDevice(app1.GetIds())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewEndDeviceRegistryClient(cc)

		// Test batch fetch with cluster authorization
		list, err := reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			FieldMask: ttnpb.FieldMask("ids"),
		}, is.WithClusterAuth())
		if a.So(err, should.BeNil) {
			a.So(list.EndDevices, should.HaveLength, 5)
		}

		created, err := reg.Create(ctx, &ttnpb.CreateEndDeviceRequest{
			EndDevice: &ttnpb.EndDevice{
				Ids: &ttnpb.EndDeviceIdentifiers{
					ApplicationIds: app1.GetIds(),
					DeviceId:       "foo",
				},
				Name: "Foo Device",
			},
		}, creds)
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Device")
		}

		got, err := reg.Get(ctx, &ttnpb.GetEndDeviceRequest{
			EndDeviceIds: created.GetIds(),
			FieldMask:    ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateEndDeviceRequest{
			EndDevice: &ttnpb.EndDevice{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		list, err = reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: app1.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.EndDevices, should.HaveLength, 6) {
			var found bool
			for _, item := range list.EndDevices {
				if item.GetIds().GetDeviceId() == created.GetIds().GetDeviceId() {
					found = true
					a.So(item.Name, should.Equal, updated.Name)
				}
			}
			a.So(found, should.BeTrue)
		}

		_, err = reg.Delete(ctx, created.GetIds(), creds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}

func TestEndDevicesPagination(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	app1 := p.NewApplication(usr1.GetOrganizationOrUserIdentifiers())
	for i := 0; i < 3; i++ {
		p.NewEndDevice(app1.GetIds())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewEndDeviceRegistryClient(cc)

		var md metadata.MD

		list, err := reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: app1.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
			Limit:          2,
			Page:           1,
		}, creds, grpc.Header(&md))
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.EndDevices, should.HaveLength, 2)
			a.So(md.Get("x-total-count"), should.Resemble, []string{"3"})
		}

		list, err = reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: app1.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
			Limit:          2,
			Page:           2,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.EndDevices, should.HaveLength, 1)
		}

		list, err = reg.List(ctx, &ttnpb.ListEndDevicesRequest{
			ApplicationIds: app1.GetIds(),
			FieldMask:      ttnpb.FieldMask("name"),
			Limit:          2,
			Page:           3,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.EndDevices, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}
