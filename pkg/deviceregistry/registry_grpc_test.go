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
	"go.thethings.network/lorawan-stack/pkg/auth/rights"
	"go.thethings.network/lorawan-stack/pkg/component"
	. "go.thethings.network/lorawan-stack/pkg/deviceregistry"
	errors "go.thethings.network/lorawan-stack/pkg/errorsv3"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

var (
	ctxWithoutRights = rights.NewContextWithFetcher(
		context.Background(),
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) ([]ttnpb.Right, error) {
			return []ttnpb.Right{}, nil
		}),
	)
	ctxWithRights = rights.NewContextWithFetcher(
		context.Background(),
		rights.FetcherFunc(func(ctx context.Context, ids ttnpb.Identifiers) ([]ttnpb.Right, error) {
			return []ttnpb.Right{ttnpb.RIGHT_APPLICATION_DEVICES_READ, ttnpb.RIGHT_APPLICATION_DEVICES_WRITE}, nil
		}),
	)
)

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	dev, err := dr.SetDevice(ctxWithoutRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(dev, should.BeNil)

	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	devs, err := dr.ListDevices(ctxWithoutRights, &pb.EndDeviceIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	devs, err = dr.ListDevices(ctxWithRights, &pb.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = pb.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = pb.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], pb), should.BeEmpty)
	}

	_, err = dr.DeleteDevice(ctxWithoutRights, &pb.EndDeviceIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	v, err := dr.DeleteDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(v, should.Equal, ttnpb.Empty)

	_, err = dr.ListDevices(ctxWithoutRights, &pb.EndDeviceIdentifiers)
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	devs, err = dr.ListDevices(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(devs.EndDevices, should.BeEmpty)
}

func TestSetDeviceNoProcessor(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	_, err := dr.SetDevice(ctxWithoutRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)

	dev, err := dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	old := deepcopy.Copy(pb).(*ttnpb.EndDevice)

	for pb.GetLocation().GetLatitude() == old.GetLocation().GetLatitude() {
		pb.Location = ttnpb.NewPopulatedLocation(test.Randy, false)
	}

	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{
		Device: *pb,
		FieldMask: &pbtypes.FieldMask{
			Paths: []string{"location.latitude"},
		},
	})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	if !a.So(dev, should.Resemble, pb) {
		pretty.Ldiff(t, dev, pb)
	}

	if old.Location == nil {
		old.Location = &ttnpb.Location{}
	}
	old.Location.Latitude = pb.GetLocation().GetLatitude()
	pb = old

	got, err := FindByIdentifiers(dr.Interface, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	pb.CreatedAt = got.GetCreatedAt()
	pb.UpdatedAt = got.GetUpdatedAt()
	a.So(pretty.Diff(got.EndDevice, pb), should.BeEmpty)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(dev, should.BeNil)
}

func TestListDevices(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	dev1, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	dev2, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	devs, err := dr.ListDevices(ctxWithRights, &dev1.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev1.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev1.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev1.EndDevice), should.BeEmpty)
	}

	devs, err = dr.ListDevices(ctxWithRights, &dev2.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev2.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev2.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev2.EndDevice), should.BeEmpty)
	}
}

func TestGetDevice(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.GetDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	pb.CreatedAt = time.Time{}
	pb.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.GetDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestDeleteDevice(t *testing.T) {
	a := assertions.New(t)
	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())))).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.DeleteDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrDeviceNotFound)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(pb)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	v, err = dr.DeleteDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
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

	v, err = dr.DeleteDevice(ctxWithRights, &pb.EndDeviceIdentifiers)
	a.So(err, should.EqualErrorOrDefinition, ErrTooManyDevices)
	a.So(v, should.BeNil)
}

func TestSetDeviceProcessor(t *testing.T) {
	errTest := errors.DefineInternal("test", "test")

	var procErr error

	dr := test.Must(NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedMapStoreClient(mapstore.New())),
		WithSetDeviceProcessor(func(_ context.Context, _ bool, dev *ttnpb.EndDevice, fields ...string) (*ttnpb.EndDevice, []string, error) {
			if procErr != nil {
				return nil, nil, procErr
			}
			return dev, fields, nil
		}),
	)).(*RegistryRPC)

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	a := assertions.New(t)

	dev, err := dr.SetDevice(ctxWithoutRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(errors.IsPermissionDenied(err), should.BeTrue)
	a.So(dev, should.BeNil)

	procErr = errors.New("err")
	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.HaveSameErrorDefinitionAs, ErrProcessorFailed)
	a.So(dev, should.BeNil)

	procErr = errTest.WithAttributes("foo", "bar")
	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.Resemble, procErr)
	a.So(dev, should.BeNil)

	procErr = nil
	dev, err = dr.SetDevice(ctxWithRights, &ttnpb.SetDeviceRequest{Device: *pb})
	a.So(err, should.BeNil)
	pb.CreatedAt = dev.GetCreatedAt()
	pb.UpdatedAt = dev.GetUpdatedAt()
	a.So(dev, should.Resemble, pb)
}
