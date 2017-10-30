// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/cmd/shared"
	"github.com/TheThingsNetwork/ttn/pkg/component"
	. "github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/errors"
	"github.com/TheThingsNetwork/ttn/pkg/log"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var logger log.Stack

func init() {
	var err error
	logger, err = log.NewLogger(log.WithLevel(log.DebugLevel))
	if err != nil {
		panic(err)
	}
}

type findDeviceByIdentifiersOut struct {
	devs []*Device
	err  error
}

type mockRegistry struct {
	Interface
	findDeviceByIdentifiersOut findDeviceByIdentifiersOut
}

func (r *mockRegistry) FindDeviceByIdentifiers(ids ...*ttnpb.EndDeviceIdentifiers) ([]*Device, error) {
	out := r.findDeviceByIdentifiersOut
	return out.devs, out.err
}

func TestListDevices(t *testing.T) {
	a := assertions.New(t)

	reg := &mockRegistry{}
	dr := NewRPC(component.New(logger, &component.Config{shared.DefaultServiceBase}), reg)
	for _, tc := range []struct {
		findDeviceByIdentifiersOut findDeviceByIdentifiersOut
		devs                       *ttnpb.EndDevices
		err                        error
	}{
		{
			findDeviceByIdentifiersOut{[]*Device{{EndDevice: &ttnpb.EndDevice{}}}, nil},
			&ttnpb.EndDevices{[]*ttnpb.EndDevice{&ttnpb.EndDevice{}}},
			nil,
		},
		{
			findDeviceByIdentifiersOut{},
			nil,
			ErrDeviceNotFound.New(nil),
		},
		{
			findDeviceByIdentifiersOut{nil, errors.New("test")},
			nil,
			errors.New("test"),
		},
	} {
		reg.findDeviceByIdentifiersOut = tc.findDeviceByIdentifiersOut
		devs, err := dr.ListDevices(nil, nil)
		a.So(err, should.Resemble, tc.err)
		a.So(devs, should.Resemble, tc.devs)
	}
}

func TestGetDevice(t *testing.T) {
	a := assertions.New(t)

	reg := &mockRegistry{}
	dr := NewRPC(nil, reg)
	for _, tc := range []struct {
		findDeviceByIdentifiersOut findDeviceByIdentifiersOut
		dev                        *ttnpb.EndDevice
		err                        error
	}{
		{
			findDeviceByIdentifiersOut{[]*Device{{EndDevice: &ttnpb.EndDevice{}}}, nil},
			&ttnpb.EndDevice{},
			nil,
		},
		{
			findDeviceByIdentifiersOut{},
			nil,
			ErrDeviceNotFound.New(nil),
		},
		{
			findDeviceByIdentifiersOut{nil, errors.New("test")},
			nil,
			errors.New("test"),
		},
		{
			findDeviceByIdentifiersOut{[]*Device{{EndDevice: &ttnpb.EndDevice{}}, {EndDevice: &ttnpb.EndDevice{}}}, nil},
			nil,
			ErrTooManyDevices.New(nil),
		},
	} {
		reg.findDeviceByIdentifiersOut = tc.findDeviceByIdentifiersOut
		dev, err := dr.GetDevice(nil, nil)
		a.So(err, should.Resemble, tc.err)
		a.So(dev, should.Resemble, tc.dev)
	}
}
