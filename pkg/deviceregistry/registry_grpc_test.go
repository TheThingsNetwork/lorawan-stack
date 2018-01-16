// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry_test

import (
	"context"
	"testing"

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
	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	dev := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(v, should.NotBeNil)

	devs, err := dr.ListDevices(context.Background(), &dev.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(devs.EndDevices, should.HaveLength, 1) ||
		!a.So(pretty.Diff(devs.EndDevices[0], dev), should.BeEmpty) {
		return
	}

	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(v, should.NotBeNil)

	devs, err = dr.ListDevices(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(devs, should.BeNil)
}

func TestSetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	dev := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)
}

func TestListDevicesNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

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
	if !a.So(err, should.BeNil) ||
		!a.So(devs.EndDevices, should.HaveLength, 1) ||
		!a.So(devs.EndDevices[0], should.Resemble, dev1.EndDevice) {
		return
	}

	devs, err = dr.ListDevices(context.Background(), &dev2.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) ||
		!a.So(devs.EndDevices, should.HaveLength, 1) ||
		!a.So(devs.EndDevices[0], should.Resemble, dev2.EndDevice) {
		return
	}

	devs, err = dr.ListDevices(context.Background(), &ttnpb.EndDeviceIdentifiers{})
	if !a.So(err, should.BeNil) ||
		!a.So(devs.EndDevices, should.HaveLength, 2) {
		return
	}
}

func TestGetDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	dev := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)
}

func TestDeleteDeviceNoCheck(t *testing.T) {
	a := assertions.New(t)
	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())))

	dev := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	v, err := dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	_, err = dr.Interface.Create(dev)
	if !a.So(err, should.BeNil) {
		return
	}

	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(v, should.BeNil)
}

func TestCheck(t *testing.T) {
	a := assertions.New(t)

	errTest := &errors.ErrDescriptor{
		MessageFormat: "Test",
		Type:          errors.Internal,
		Code:          1,
	}
	errTest.Register()

	var checkErr error

	listCheck := func(context.Context, *ttnpb.EndDeviceIdentifiers) error {
		return checkErr
	}
	getCheck := func(context.Context, *ttnpb.EndDeviceIdentifiers) error {
		return checkErr
	}
	deleteCheck := func(context.Context, *ttnpb.EndDeviceIdentifiers) error {
		return checkErr
	}
	setCheck := func(context.Context, *ttnpb.EndDevice, ...string) error {
		return checkErr
	}

	dr := NewRPC(component.New(test.GetLogger(t), &component.Config{}), New(store.NewTypedStoreClient(mapstore.New())),
		WithListDevicesCheck(listCheck),
		WithGetDeviceCheck(getCheck),
		WithSetDeviceCheck(setCheck),
		WithDeleteDeviceCheck(deleteCheck),
	)

	dev := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	// Set

	checkErr = errors.New("err")
	v, err := dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
	a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
	a.So(v, should.BeNil)

	checkErr = errTest.New(nil)
	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(err, should.Equal, checkErr)
	a.So(v, should.BeNil)

	checkErr = nil
	v, err = dr.SetDevice(context.Background(), &ttnpb.SetDeviceRequest{Device: *dev})
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)

	// Get

	checkErr = errors.New("err")
	ret, err := dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
	a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
	a.So(ret, should.BeNil)

	checkErr = errTest.New(nil)
	ret, err = dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.Equal, checkErr)
	a.So(ret, should.BeNil)

	checkErr = nil
	ret, err = dr.GetDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(pretty.Diff(ret, dev), should.BeEmpty)

	// List

	checkErr = errors.New("err")
	devs, err := dr.ListDevices(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
	a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
	a.So(devs, should.BeNil)

	checkErr = errTest.New(nil)
	devs, err = dr.ListDevices(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.Equal, checkErr)
	a.So(devs, should.BeNil)

	checkErr = nil
	devs, err = dr.ListDevices(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	if a.So(devs, should.NotBeNil) && a.So(devs.EndDevices, should.HaveLength, 1) {
		a.So(pretty.Diff(devs.EndDevices[0], dev), should.BeEmpty)
	}

	// Delete

	checkErr = errors.New("err")
	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(errors.From(err).Code(), should.Equal, ErrCheckFailed.Code)
	a.So(errors.From(err).Type(), should.Equal, ErrCheckFailed.Type)
	a.So(v, should.BeNil)

	checkErr = errTest.New(nil)
	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.Equal, checkErr)
	a.So(v, should.BeNil)

	checkErr = nil
	v, err = dr.DeleteDevice(context.Background(), &dev.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	a.So(v, should.NotBeNil)
}
