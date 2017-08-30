// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry_test

import (
	"math/rand"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var randy = rand.New(rand.NewSource(42))

func newPopulatedEndDevice() *ttnpb.EndDevice {
	ed := &ttnpb.EndDevice{}
	devEUI := types.NewPopulatedEUI64(randy)
	ed.DevEUI = &devEUI
	joinEUI := types.NewPopulatedEUI64(randy)
	ed.JoinEUI = &joinEUI
	devAddr := types.NewPopulatedDevAddr(randy)
	ed.DevAddr = &devAddr
	ed.ApplicationID = "test"
	ed.DeviceID = "test"
	ed.TenantID = "test"
	return ed
}

func TestDeviceRegistry(t *testing.T) {
	a := assertions.New(t)
	r := New(mapstore.New())

	ed := newPopulatedEndDevice()

	device, err := r.Register(ed)
	a.So(err, should.BeNil)
	if a.So(device, should.NotBeNil) {
		a.So(device.EndDevice, should.Resemble, ed)
	}

	found, err := r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
		TenantID:      ed.TenantID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(found[0].EndDevice, should.Resemble, ed)
	}

	updated := newPopulatedEndDevice()
	for device.EndDeviceIdentifiers == updated.EndDeviceIdentifiers {
		updated = newPopulatedEndDevice()
	}
	device.EndDevice = updated
	a.So(device.Update(), should.BeNil)

	found, err = r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
		TenantID:      ed.TenantID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.HaveLength, 0)
	}

	found, err = r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        updated.DevEUI,
		JoinEUI:       updated.JoinEUI,
		DevAddr:       updated.DevAddr,
		DeviceID:      updated.DeviceID,
		ApplicationID: updated.ApplicationID,
		TenantID:      updated.TenantID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(found[0].EndDevice, should.Resemble, updated)
	}

	a.So(device.Deregister(), should.BeNil)

	found, err = r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        updated.DevEUI,
		JoinEUI:       updated.JoinEUI,
		DevAddr:       updated.DevAddr,
		DeviceID:      updated.DeviceID,
		ApplicationID: updated.ApplicationID,
		TenantID:      updated.TenantID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.HaveLength, 0)
	}
}
