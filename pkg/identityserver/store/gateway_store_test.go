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
		db.AutoMigrate(&Gateway{})
		store := GetGatewayStore(db)

		created, err := store.CreateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{
				GatewayID: "foo",
				EUI:       &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8},
			},
			Name:        "Foo Gateway",
			Description: "The Amazing Foo Gateway",
		})
		a.So(err, should.BeNil)
		a.So(created.GatewayID, should.Equal, "foo")
		a.So(created.Name, should.Equal, "Foo Gateway")
		a.So(created.Description, should.Equal, "The Amazing Foo Gateway")
		a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
		a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))

		got, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"}, &pbtypes.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		a.So(got.GatewayID, should.Equal, "foo")
		a.So(got.Name, should.Equal, "Foo Gateway")
		a.So(got.Description, should.BeEmpty)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, created.UpdatedAt)

		byEUI, err := store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{EUI: &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}}, &pbtypes.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		a.So(byEUI.GatewayID, should.Equal, got.GatewayID)

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
		}, &pbtypes.FieldMask{Paths: []string{"description"}})
		a.So(err, should.BeNil)
		a.So(updated.Description, should.Equal, "The Amazing Foobar Gateway")
		a.So(updated.CreatedAt, should.Equal, created.CreatedAt)
		a.So(updated.UpdatedAt, should.HappenAfter, created.CreatedAt)

		_, err = store.UpdateGateway(ctx, &ttnpb.Gateway{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayID: "foo"},
			Description:        "The Foobar Gateway",
			UpdatedAt:          created.UpdatedAt,
		}, &pbtypes.FieldMask{Paths: []string{"description"}})
		a.So(err, should.NotBeNil)

		got, err = store.GetGateway(ctx, &ttnpb.GatewayIdentifiers{GatewayID: "foo"}, nil)
		a.So(err, should.BeNil)
		a.So(got.GatewayID, should.Equal, created.GatewayID)
		a.So(got.Name, should.Equal, created.Name)
		a.So(got.Description, should.Equal, updated.Description)
		a.So(got.CreatedAt, should.Equal, created.CreatedAt)
		a.So(got.UpdatedAt, should.Equal, updated.UpdatedAt)

		list, err := store.FindGateways(ctx, nil, &pbtypes.FieldMask{Paths: []string{"name"}})
		a.So(err, should.BeNil)
		if a.So(list, should.HaveLength, 1) {
			a.So(list[0].Name, should.EndWith, got.Name)
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

	})
}
