// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package deviceregistry_test

import (
	"fmt"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/kr/pretty"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDeviceRegistry(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedStoreClient(mapstore.New()))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	device, err := r.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(device, should.NotBeNil) {
		a.So(device.EndDevice, should.Resemble, ed)
	}

	found, err := r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(pretty.Diff(found[0].EndDevice, ed), should.BeEmpty)
	}

	updated := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for device.EndDeviceIdentifiers == updated.EndDeviceIdentifiers {
		updated = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	device.EndDevice = updated
	if !a.So(device.Update(), should.BeNil) {
		return
	}

	found, err = r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
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
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(pretty.Diff(found[0].EndDevice, updated), should.BeEmpty)
	}

	a.So(device.Delete(), should.BeNil)

	found, err = r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		DevEUI:        updated.DevEUI,
		JoinEUI:       updated.JoinEUI,
		DevAddr:       updated.DevAddr,
		DeviceID:      updated.DeviceID,
		ApplicationID: updated.ApplicationID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.HaveLength, 0)
	}
}

func TestFindOneDeviceByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedStoreClient(mapstore.New()))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	found, err := FindOneDeviceByIdentifiers(r, &ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
	})
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	device, err := r.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(device, should.NotBeNil) {
		a.So(device.EndDevice, should.Resemble, ed)
	}

	found, err = FindOneDeviceByIdentifiers(r, &ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
	})
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && !a.So(found.EndDevice, should.Resemble, device.EndDevice) {
		pretty.Ldiff(t, found.EndDevice, device.EndDevice)
	}
	device, err = r.Create(ed)
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(device, should.NotBeNil) {
		a.So(device.EndDevice, should.Resemble, ed)
	}

	found, err = FindOneDeviceByIdentifiers(r, &ttnpb.EndDeviceIdentifiers{
		DevEUI:        ed.DevEUI,
		JoinEUI:       ed.JoinEUI,
		DevAddr:       ed.DevAddr,
		DeviceID:      ed.DeviceID,
		ApplicationID: ed.ApplicationID,
	})
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)
}

func ExampleRegistry() {
	r := New(store.NewTypedStoreClient(mapstore.New()))

	devEUI := types.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	joinEUI := types.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	devAddr := types.DevAddr([4]byte{0, 1, 2, 3})
	ed := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationID: "test",
			DeviceID:      "test",
			DevEUI:        &devEUI,
			JoinEUI:       &joinEUI,
			DevAddr:       &devAddr,
		},
	}

	dev, err := r.Create(ed)
	if err != nil {
		panic(fmt.Errorf("Failed to create device %s", err))
	}

	dev.NextDevNonce++
	dev.NextJoinNonce++
	dev.NextRJCount0++
	dev.NextRJCount1++
	dev.DeviceID = "differentID"
	devAddr = types.DevAddr([4]byte{4, 3, 2, 1})
	dev.DevAddr = &devAddr
	err = dev.Update()
	if err != nil {
		panic(fmt.Errorf("Failed to update device %s", err))
	}

	devs, err := r.FindDeviceByIdentifiers(&ttnpb.EndDeviceIdentifiers{
		ApplicationID: "test",
	})
	if err != nil {
		panic(fmt.Errorf("Failed to find device by identifiers %s", err))
	}
	if len(devs) != 1 {
		panic(fmt.Errorf("Expected to find 1 device, got %d", len(devs)))
	}
	dev = devs[0]

	err = dev.Delete()
	if err != nil {
		panic(fmt.Errorf("Failed to delete device %s", err))
	}
}
