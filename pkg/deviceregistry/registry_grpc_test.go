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

	"github.com/TheThingsNetwork/ttn/pkg/component"
	. "github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRegistryRPC(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(v, should.NotBeNil)

	devs, err := dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = ed.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = ed.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], ed), should.BeEmpty)
	}

	v, err = dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(v, should.NotBeNil)

	devs, err = dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(devs.EndDevices, should.BeEmpty)
}

func TestSetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)
}

func TestListDevicesNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	devs, err := dr.ListDevices(context.Background(), nil)
	a.So(err, should.NotBeNil)
	a.So(devs, should.BeNil)

	dev1, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		return
	}

	dev2, err := dr.Interface.Create(ttnpb.NewPopulatedEndDevice(test.Randy, false))
	if !a.So(err, should.BeNil) {
		return
	}

	devs, err = dr.ListDevices(context.Background(), &dev1.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev1.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev1.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev1.EndDevice), should.BeEmpty)
	}

	devs, err = dr.ListDevices(context.Background(), &dev2.EndDeviceIdentifiers)
	if a.So(err, should.BeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		devs.EndDevices[0].CreatedAt = dev2.EndDevice.GetCreatedAt()
		devs.EndDevices[0].UpdatedAt = dev2.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(devs.EndDevices[0], dev2.EndDevice), should.BeEmpty)
	}
}

func TestGetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	ed.CreatedAt = time.Time{}
	ed.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)
}

func TestDeleteDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	ed.CreatedAt = time.Time{}
	ed.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	ed.CreatedAt = time.Time{}
	ed.UpdatedAt = time.Time{}
	_, err = dr.Interface.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
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

	dr := NewRPC(component.MustNew(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())),
		WithListDevicesCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
		WithGetDeviceCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
		WithSetDeviceCheck(func(context.Context, *ttnpb.EndDevice, ...string) error { return checkErr }),
		WithDeleteDeviceCheck(func(context.Context, *ttnpb.EndDeviceIdentifiers) error { return checkErr }),
	)

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	t.Run("SetDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
		a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
		a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
		a.So(v, should.BeNil)

		checkErr = errTest.New(nil)
		v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
		a.So(err, should.Equal, checkErr)
		a.So(v, should.BeNil)

		checkErr = nil
		v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *ed})
		a.So(err, should.BeNil)
		a.So(v, should.NotBeNil)
	})

	if !t.Run("GetDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		ret, err := dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
		a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
		a.So(ret, should.BeNil)

		checkErr = errTest.New(nil)
		ret, err = dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)
		a.So(ret, should.BeNil)

		checkErr = nil
		ret, err = dr.GetDevice(context.Background(), &ed.EndDeviceIdentifiers)
		if !a.So(err, should.BeNil) {
			return
		}
		ret.CreatedAt = ed.GetCreatedAt()
		ret.UpdatedAt = ed.GetUpdatedAt()
		a.So(pretty.Diff(ret, ed), should.BeEmpty)
	}) {
		return
	}

	t.Run("ListDevices", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		devs, err := dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
		a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
		a.So(devs, should.BeNil)

		checkErr = errTest.New(nil)
		devs, err = dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)
		a.So(devs, should.BeNil)

		checkErr = nil
		devs, err = dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.BeNil)
		if a.So(devs, should.NotBeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
			devs.EndDevices[0].CreatedAt = ed.GetCreatedAt()
			devs.EndDevices[0].UpdatedAt = ed.GetUpdatedAt()
			a.So(pretty.Diff(devs.EndDevices[0], ed), should.BeEmpty)
		}
	})

	t.Run("DeleteDevice", func(t *testing.T) {
		a := assertions.New(t)

		checkErr = errors.New("err")
		_, err := dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
		a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)

		checkErr = errTest.New(nil)
		_, err = dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.Equal, checkErr)

		checkErr = nil
		_, err = dr.DeleteDevice(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.BeNil)

		devs, err := dr.ListDevices(context.Background(), &ed.EndDeviceIdentifiers)
		a.So(err, should.BeNil)
		if a.So(devs, should.NotBeNil) {
			a.So(devs.EndDevices, should.BeEmpty)
		}
	})
}
