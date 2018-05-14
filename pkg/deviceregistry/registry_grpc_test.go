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

package deviceregistry_test

import (
	"context"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/errors/common"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	errshould "go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	_, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, errshould.Describe, common.ErrPermissionDenied)

	v, err := dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	devs, err := dr.ListDevices(context.Background(), &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, common.ErrPermissionDenied)

	devs, err = dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = pb.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], pb), should.BeEmpty)
	}

	_, err = dr.DeleteDevice(context.Background(), &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, common.ErrPermissionDenied)

	v, err = dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	_, err = dr.ListDevices(context.Background(), &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, common.ErrPermissionDenied)

	devs, err = dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(devs.EndDevices, should.BeEmpty)
}

func TestSetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	_, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, errshould.Describe, common.ErrPermissionDenied)

	v, err := dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.BeNil)
	a.So(v, should.Equal, ttnpb.Empty)

	old := deepcopy.Copy(pb).(*ttnpb.EndDevice)

	for pb.GetLocation().GetLatitude() == old.GetLocation().GetLatitude() {
		pb.Location = ttnpb.NewPopulatedLocation(test.Randy, false)
	}

	v, err = dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{
		Device: *pb,
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{"location.latitude"},
		},
	})
	a.So(err, should.BeNil)
	a.So(v, should.Equal, ttnpb.Empty)

	if old.Location == nil {
		old.Location = &ttnpb.Location{}
	}
	old.Location.Latitude = pb.GetLocation().GetLatitude()
	pb = old

	got, err := FindOneDeviceByIdentifiers(dr.Interface, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	pb.CreatedAt = got.GetCreatedAt()
	pb.UpdatedAt = got.GetUpdatedAt()
	if !a.So(got.EndDevice, should.Resemble, pb) {
		pretty.Ldiff(t, got.EndDevice, pb)
	}

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, errshould.Describe, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestListDevicesNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	devs, err := dr.ListDevices(ctx, &ttnpb.EndDeviceIdentifiers{})
	a.So(err, errshould.Describe, rights.ErrInvalidApplicationID)
	a.So(devs, should.BeNil)

	dev1, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	dev2, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	devs, err = dr.ListDevices(ctx, &dev1.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev1.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev1.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev1.EndDevice), should.BeEmpty)
	}

	devs, err = dr.ListDevices(ctx, &dev2.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev2.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev2.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev2.EndDevice), should.BeEmpty)
	}
}

func TestGetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestDeleteDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
	a.So(err, errshould.Describe, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestCheck(t *testing.T) {
	errTest := &errors.ErrDescriptor{
		MessageFormat: "Test",
		Type:          errors.Internal,
		Code:          1,
	}
	errTest.Register()

	var checkErr error

	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithListDevicesCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
		WithGetDeviceCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
		WithSetDeviceCheck(func(context.Context, *ttnpb.EndDevice, ...string) error { return checkErr }),
		WithDeleteDeviceCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
	)).(*RegistryRPC)

	ctx := rights.NewContext(context.Background(), []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE,
	})

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	t.Run("SetDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		v, err := dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
		a.So(err, errshould.Describe, common.ErrCheckFailed)
		a.So(v, should.BeNil)

		checkErr = errTest.New(nil)
		v, err = dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
		a.So(err, should.Equal, checkErr)
		a.So(v, should.BeNil)

		checkErr = nil
		v, err = dr.SetDevice(ctx, &ttnpb.SetDeviceRequest{Device: *pb})
		a.So(err, should.BeNil)
		a.So(v, should.NotBeNil)
	})

	if !t.Run("GetDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		ret, err := dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, errshould.Describe, common.ErrCheckFailed)
		a.So(ret, should.BeNil)

		checkErr = errTest.New(nil)
		ret, err = dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)
		a.So(ret, should.BeNil)

		checkErr = nil
		ret, err = dr.GetDevice(ctx, &pb.EndDeviceIdentifiers)
		if !a.So(err, should.BeNil) {
			t.FailNow()
		}
		ret.CreatedAt = pb.GetCreatedAt()
		ret.UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(ret, pb), should.BeEmpty)
	}) {
		t.FailNow()
	}

	t.Run("ListDevices", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		devs, err := dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, errshould.Describe, common.ErrCheckFailed)
		a.So(devs, should.BeNil)

		checkErr = errTest.New(nil)
		devs, err = dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)
		a.So(devs, should.BeNil)

		checkErr = nil
		devs, err = dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.BeNil)
		if a.So(devs, should.NotBeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
			devs.EndDevices[0].CreatedAt = pb.GetCreatedAt()
			devs.EndDevices[0].UpdatedAt = pb.GetUpdatedAt()
			a.So(pretty.Diff(devs.EndDevices[0], pb), should.BeEmpty)
		}
	})

	t.Run("DeleteDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		_, err := dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, errshould.Describe, common.ErrCheckFailed)

		checkErr = errTest.New(nil)
		_, err = dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)

		checkErr = nil
		_, err = dr.DeleteDevice(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.BeNil)

		devs, err := dr.ListDevices(ctx, &pb.EndDeviceIdentifiers)
		a.So(err, should.BeNil)
		if a.So(devs, should.NotBeNil) {
			a.So(devs.EndDevices, should.BeEmpty)
		}
	})
}
