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
	"fmt"
	"testing"

	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test"
)

func TestEntitySearch(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db, &Attribute{}, &Application{}, &Client{}, &Gateway{}, &Account{}, &User{}, &Organization{})

		store := newStore(db)
		s := GetEntitySearch(db)

		for _, name := range []string{"foo", "bar"} {
			store.createEntity(ctx, &Application{
				ApplicationID: fmt.Sprintf("the-%s-app", name),
				Name:          fmt.Sprintf("The %s application", name),
				Description:   fmt.Sprintf("This application does %s stuff", name),
				Attributes: []Attribute{
					{Key: "test", Value: name},
				},
			})

			store.createEntity(ctx, &Client{
				ClientID:    fmt.Sprintf("the-%s-cli", name),
				Name:        fmt.Sprintf("The %s client", name),
				Description: fmt.Sprintf("This client does %s stuff", name),
				Attributes: []Attribute{
					{Key: "test", Value: name},
				},
			})

			store.createEntity(ctx, &Gateway{
				GatewayID:   fmt.Sprintf("the-%s-gtw", name),
				Name:        fmt.Sprintf("The %s gateway", name),
				Description: fmt.Sprintf("This gateway does %s stuff", name),
				Attributes: []Attribute{
					{Key: "test", Value: name},
				},
			})

			store.createEntity(ctx, &User{
				Account: Account{
					UID: fmt.Sprintf("the-%s-usr", name),
				},
				Name:        fmt.Sprintf("The %s user", name),
				Description: fmt.Sprintf("This user does %s stuff", name),
				Attributes: []Attribute{
					{Key: "test", Value: name},
				},
				PrimaryEmailAddress: fmt.Sprintf("%s@example.com", name),
			})

			store.createEntity(ctx, &Organization{
				Account: Account{
					UID: fmt.Sprintf("the-%s-org", name),
				},
				Name:        fmt.Sprintf("The %s organization", name),
				Description: fmt.Sprintf("This organization does %s stuff", name),
				Attributes: []Attribute{
					{Key: "test", Value: name},
				},
			})
		}

		for _, entityType := range []string{"application", "client", "gateway", "user", "organization"} {
			t.Run(entityType, func(t *testing.T) {
				ids, err := s.FindEntities(ctx, &ttnpb.SearchEntitiesRequest{
					IDContains:          "foo",
					NameContains:        "foo",
					DescriptionContains: "foo",
					AttributesContain: map[string]string{
						"test": "foo",
					},
				}, entityType)

				a.So(err, should.BeNil)
				a.So(ids, should.HaveLength, 1)

				ids, err = s.FindEntities(ctx, &ttnpb.SearchEntitiesRequest{
					IDContains: "foo",
				}, entityType)

				a.So(err, should.BeNil)
				a.So(ids, should.HaveLength, 1)

				ids, err = s.FindEntities(ctx, &ttnpb.SearchEntitiesRequest{
					NameContains: "foo",
				}, entityType)

				a.So(err, should.BeNil)
				a.So(ids, should.HaveLength, 1)

				ids, err = s.FindEntities(ctx, &ttnpb.SearchEntitiesRequest{
					DescriptionContains: "foo",
				}, entityType)

				a.So(err, should.BeNil)
				a.So(ids, should.HaveLength, 1)

				ids, err = s.FindEntities(ctx, &ttnpb.SearchEntitiesRequest{
					AttributesContain: map[string]string{
						"test": "foo",
					},
				}, entityType)

				a.So(err, should.BeNil)
				a.So(ids, should.HaveLength, 1)
			})
		}
	})
}
