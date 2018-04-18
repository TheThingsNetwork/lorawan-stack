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

	. "github.com/TheThingsNetwork/ttn/pkg/deviceregistry"
	"github.com/TheThingsNetwork/ttn/pkg/store"
	"github.com/TheThingsNetwork/ttn/pkg/store/mapstore"
	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/types"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	pbtypes "github.com/gogo/protobuf/types"
	"github.com/kr/pretty"
	"github.com/mohae/deepcopy"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestRegistry(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedStoreClient(mapstore.New()))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	dev, err := r.Create(deepcopy.Copy(ed).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		ed.CreatedAt = dev.EndDevice.GetCreatedAt()
		ed.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, ed)
	}

	found, err := r.FindBy(ed, "EndDeviceIdentifiers")
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		found[0].CreatedAt = dev.EndDevice.GetCreatedAt()
		found[0].UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(found[0].EndDevice, ed), should.BeEmpty)
	}

	updated := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for dev.EndDevice.EndDeviceIdentifiers.Equal(updated.EndDeviceIdentifiers) {
		updated = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	updated.CreatedAt = dev.EndDevice.GetCreatedAt()
	updated.UpdatedAt = dev.EndDevice.GetUpdatedAt()
	dev.EndDevice = updated

	if !a.So(dev.Store(), should.BeNil) {
		return
	}

	found, err = r.FindBy(ed, "EndDeviceIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}

	ed = updated

	found, err = r.FindBy(ed, "EndDeviceIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		found[0].CreatedAt = ed.GetCreatedAt()
		found[0].UpdatedAt = ed.GetUpdatedAt()
		a.So(pretty.Diff(found[0].EndDevice, ed), should.BeEmpty)
	}

	a.So(dev.Delete(), should.BeNil)

	found, err = r.FindBy(ed, "EndDeviceIdentifiers")
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}
}

func TestFindDeviceByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedStoreClient(mapstore.New()))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	ed.Attributes = &pbtypes.Struct{
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

	dev, err := r.Create(deepcopy.Copy(ed).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		ed.CreatedAt = dev.EndDevice.GetCreatedAt()
		ed.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, ed)
	}

	found, err := FindDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		a.So(pretty.Diff(found[0].EndDevice, ed), should.BeEmpty)
	}

	updated := ttnpb.NewPopulatedEndDevice(test.Randy, false)
	for dev.EndDevice.EndDeviceIdentifiers.Equal(updated.EndDeviceIdentifiers) {
		updated = ttnpb.NewPopulatedEndDevice(test.Randy, false)
	}
	updated.CreatedAt = ed.GetCreatedAt()
	updated.UpdatedAt = ed.GetUpdatedAt()
	dev.EndDevice = updated

	if !a.So(dev.Store(), should.BeNil) {
		return
	}

	found, err = FindDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}

	ed = updated

	found, err = FindDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) && a.So(found, should.HaveLength, 1) {
		ed.CreatedAt = found[0].EndDevice.GetCreatedAt()
		ed.UpdatedAt = found[0].EndDevice.GetUpdatedAt()
		a.So(pretty.Diff(found[0].EndDevice, ed), should.BeEmpty)
	}

	a.So(dev.Delete(), should.BeNil)

	found, err = FindDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	a.So(err, should.BeNil)
	if a.So(found, should.NotBeNil) {
		a.So(found, should.BeEmpty)
	}
}

func TestFindOneDeviceByIdentifiers(t *testing.T) {
	a := assertions.New(t)
	r := New(store.NewTypedStoreClient(mapstore.New()))

	ed := ttnpb.NewPopulatedEndDevice(test.Randy, false)

	found, err := FindOneDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	a.So(err, should.NotBeNil)
	a.So(found, should.BeNil)

	dev, err := r.Create(deepcopy.Copy(ed).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		ed.CreatedAt = dev.EndDevice.GetCreatedAt()
		ed.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, ed)
	}

	found, err = FindOneDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
	if !a.So(err, should.BeNil) {
		return
	}
	a.So(pretty.Diff(found.EndDevice, ed), should.BeEmpty)

	dev, err = r.Create(deepcopy.Copy(ed).(*ttnpb.EndDevice))
	if !a.So(err, should.BeNil) {
		return
	}
	if a.So(dev, should.NotBeNil) {
		ed.CreatedAt = dev.EndDevice.GetCreatedAt()
		ed.UpdatedAt = dev.EndDevice.GetUpdatedAt()
		a.So(dev.EndDevice, should.Resemble, ed)
	}

	found, err = FindOneDeviceByIdentifiers(r, &ed.EndDeviceIdentifiers)
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
	err = dev.Store()
	if err != nil {
		panic(fmt.Errorf("Failed to update device %s", err))
	}

	dev, err = FindOneDeviceByIdentifiers(r, &ttnpb.EndDeviceIdentifiers{
		ApplicationID: "test",
	})
	if err != nil {
		panic(fmt.Errorf("Failed to find device by identifiers %s", err))
	}

	err = dev.Delete()
	if err != nil {
		panic(fmt.Errorf("Failed to delete device %s", err))
	}
}
