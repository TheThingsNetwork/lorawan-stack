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
	"fmt"
	"testing"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	. "go.thethings.network/lorawan-stack/pkg/deviceregistry"
	"go.thethings.network/lorawan-stack/pkg/store"
	"go.thethings.network/lorawan-stack/pkg/store/mapstore"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestRegistry(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	dev, err := r.Create(deepcopy.Copy(pb).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		pb.CreatedAt = dev.EndDevice.GetCreatedAt()
		pb.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, pb)
	}

	i := 0
	err = r.Range(pb, 1, func(found *Device) bool {
		i++
		a.So(pretty.Diff(found.EndDevice, pb), should.BeEmpty)
		return true
	}, "EndDeviceIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	updated := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for dev.EndDevice.EndDeviceIdentifiers.Equal(updated.EndDeviceIdentifiers) {
		updated = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	dev.EndDevice = updated

	if !a.So(dev.Store(), should.BeNil) {
		return
	}

	i = 0
	err = r.Range(pb, 1, func(*Device) bool { i++; return true }, "EndDeviceIdentifiers")
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)

	pb = updated

	i = 0
	err = r.Range(pb, 1, func(found *Device) bool {
		i++
		pb.UpdatedAt = found.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(found.EndDevice, pb), should.BeEmpty)
		return true
	}, "EndDeviceIdentifiers")
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
	a.So(i, should.Equal, 1)

	a.So(dev.Delete(), should.BeNil)

	i = 0
	err = r.Range(pb, 1, func(*Device) bool { i++; return true }, "EndDeviceIdentifiers")
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)
}

func TestFindDeviceByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	pb.Attributes = &pbtypes.Struct{
		Fields: map[string]*pbtypes.Value{
			"null":   {Kind: &pbtypes.Value_NullValue{}},
			"bool":   {Kind: &pbtypes.Value_BoolValue{BoolValue: true}},
			"str":    {Kind: &pbtypes.Value_StringValue{StringValue: "bar"}},
			"number": {Kind: &pbtypes.Value_NumberValue{NumberValue: 42}},
			"list": {Kind: &pbtypes.Value_ListValue{ListValue: &pbtypes.ListValue{Values: []*pbtypes.Value{
				{&pbtypes.Value_BoolValue{BoolValue: true}},
				{&pbtypes.Value_StringValue{StringValue: "bar"}},
			}}}},
			"struct": {Kind: &pbtypes.Value_StructValue{StructValue: &pbtypes.Struct{Fields: map[string]*pbtypes.Value{
				"bool": {Kind: &pbtypes.Value_BoolValue{BoolValue: true}},
				"str":  {Kind: &pbtypes.Value_StringValue{StringValue: "bar"}},
			}}}},
		},
	}

	dev, err := r.Create(deepcopy.Copy(pb).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		pb.CreatedAt = dev.EndDevice.GetCreatedAt()
		pb.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, pb)
	}

	i := 0
	err = FindDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers, 1, func(found *Device) bool {
		i++
		a.So(pretty.Diff(found.EndDevice, pb), should.BeEmpty)
		return true
	})
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 1)

	updated := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for dev.EndDevice.EndDeviceIdentifiers.Equal(updated.EndDeviceIdentifiers) {
		updated = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	dev.EndDevice = updated

	if !a.So(dev.Store(), should.BeNil) {
		return
	}

	i = 0
	err = FindDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers, 1, func(*Device) bool { i++; return true })
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)

	pb = updated

	i = 0
	err = FindDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers, 1, func(found *Device) bool {
		i++
		pb.UpdatedAt = found.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(found.EndDevice, pb), should.BeEmpty)
		return true
	})
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 1)

	a.So(dev.Delete(), should.BeNil)

	i = 0
	err = FindDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers, 1, func(*Device) bool { i++; return true })
	a.So(err, should.BeNil)
	a.So(i, should.Equal, 0)
}

func TestFindOneDeviceByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	pb := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	found, err := FindOneDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	dev, err := r.Create(deepcopy.Copy(pb).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		pb.CreatedAt = dev.EndDevice.GetCreatedAt()
		pb.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, pb)
	}

	found, err = FindOneDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(pretty.Diff(found.EndDevice, pb), should.BeEmpty)

	dev, err = r.Create(deepcopy.Copy(pb).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		pb.CreatedAt = dev.EndDevice.GetCreatedAt()
		pb.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, pb)
	}

	found, err = FindOneDeviceByIdentifiers(r, &pb.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)
}

func ExampleRegistry() {
	r := New(store.NewTypedMapStoreClient(mapstore.New()))

	devEUI := types.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	joinEUI := types.EUI64([8]byte{0, 1, 2, 3, 4, 5, 6, 7})
	devAddr := types.DevAddr([4]byte{0, 1, 2, 3})
	pb := &ttnpb.EndDevice{
		EndDeviceIdentifiers: ttnpb.EndDeviceIdentifiers{
			ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{
				ApplicationID: "test",
			},
			DeviceID: "test",
			DevEUI:   &devEUI,
			JoinEUI:  &joinEUI,
			DevAddr:  &devAddr,
		},
	}

	dev, err := r.Create(pb)
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
	err = dev.Store("NextDevNonce", "NextJoinNonce", "NextRJCount0", "NextRJCount1", "DeviceIdentifiers.DeviceID", "DevAddr")
	if err != nil {
		panic(fmt.Errorf("Failed to update device %s", err))
	}

	dev, err = FindOneDeviceByIdentifiers(r, &ttnpb.EndDeviceIdentifiers{
		ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationID: "test"},
	})
	if err != nil {
		panic(fmt.Errorf("Failed to find device by identifiers %s", err))
	}

	err = dev.Delete()
	if err != nil {
		panic(fmt.Errorf("Failed to delete device %s", err))
	}
}
