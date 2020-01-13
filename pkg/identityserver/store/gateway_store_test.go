// Copyright © 2019 The Things Network Foundation, The Things Industries B.V.
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

package store

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/errors"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/types"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestGatewayStore(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Gateway{}, &GatewayAntenna{}, &Attribute{})
		store := GetGatewayStore(db)

		eui := &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}
		created, err := store.CreateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "foo",
				EUI:       eui,
			},
			Name:        "Foo Gateway",
			Description: "The Amazing Foo Gateway",
			Attributes: map[string]string{
				"foo": "bar",
				"bar": "baz",
				"baz": "qux",
			},
			Antennas: []ttnpb.GatewayAntenna{
				{Gain: 3, Location: ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1}},
			},
		})

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.GatewayID, should.Equal, "foo")
			a.So(created.Name, should.Equal, "Foo Gateway")
			a.So(created.Description, should.Equal, "The Amazing Foo Gateway")
			a.So(created.Attributes, should.HaveLength, 3)
			if a.So(created.Antennas, should.HaveLength, 1) {
				a.So(created.Antennas[0].Gain, should.Equal, 3)
			}
			a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
			a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		}

		got, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"}, &pbtypes.FieldMask{Paths: []string{"name", "attributes"}})

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GatewayID, should.Equal, "foo")
			a.So(got.Name, should.Equal, "Foo Gateway")
			a.So(got.Description, should.BeEmpty)
			a.So(got.Attributes, should.HaveLength, 3)
			a.So(got.CreatedAt, should.Equal, created.CreatedAt)
			a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)
		}

		byEUI, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{EUI: &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}}, &pbtypes.FieldMask{Paths: []string{"name"}})

		a.So(err, should.BeNil)
		if a.So(byEUI, should.NotBeNil) {
			a.So(byEUI.GatewayID, should.Equal, got.GatewayID)
		}

		_, err = store.UpdateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "bar"},
		}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		updated, err := store.UpdateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			Name:               "Foobar Gateway",
			Description:        "The Amazing Foobar Gateway",
			Attributes: map[string]string{
				"foo": "bar",
				"baz": "baz",
				"qux": "foo",
			},
			Antennas: []ttnpb.GatewayAntenna{
				{Gain: 6, Location: ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1}, Attributes: map[string]string{"direction": "west"}},
				{Gain: 6, Location: ttnpb.Location{Latitude: 12.345, Longitude: 23.456, Altitude: 1090, Accuracy: 1}, Attributes: map[string]string{"direction": "east"}},
			},
		}, &pbtypes.FieldMask{Paths: []string{"description", "attributes", "antennas"}})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Description, should.Equal, "The Amazing Foobar Gateway")
			a.So(updated.Attributes, should.HaveLength, 3)
			if a.So(updated.Antennas, should.HaveLength, 2) {
				a.So(updated.Antennas[0].Gain, should.Equal, 6)
				a.So(updated.Antennas[0].Attributes, should.HaveLength, 1)
				a.So(updated.Antennas[1].Gain, should.Equal, 6)
				a.So(updated.Antennas[1].Attributes, should.HaveLength, 1)
			}
			a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
			a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)
		}

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"}, nil)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GatewayID, should.Equal, created.GatewayID)
			a.So(got.Name, should.Equal, created.Name)
			a.So(got.Description, should.Equal, updated.Description)
			a.So(got.Attributes, should.Resemble, updated.Attributes)
			a.So(got.Antennas, should.HaveLength, len(updated.Antennas))
			a.So(got.CreatedAt, should.Equal, created.CreatedAt)
			a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)
		}

		list, err := store.FindGateways(ctx, nil, &pbtypes.FieldMask{Paths: []string{"name"}})

		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
		}

		updated, err = store.UpdateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			Antennas:           []ttnpb.GatewayAntenna{},
		}, &pbtypes.FieldMask{Paths: []string{"antennas"}})

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Antennas, should.HaveLength, 0)
		}

		err = store.DeleteGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"})

		a.So(err, should.BeNil)

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"}, nil)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsNotFound(err), should.BeTrue)
		}

		list, err = store.FindGateways(ctx, nil, nil)

		a.So(err, should.BeNil)
		a.So(list, should.BeEmpty)

		got, err = store.CreateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "reuse-foo-eui",
				EUI:       eui,
			},
		})

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.GatewayID, should.Equal, "reuse-foo-eui")
			a.So(got.EUI, should.Resemble, eui)
		}
	})
}
