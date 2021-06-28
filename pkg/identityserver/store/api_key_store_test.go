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
	"strings"
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/jinzhu/gorm"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
)

func TestAPIKeyStore(t *testing.T) {
	ctx := test.Context()

	WithDB(t, func(t *testing.T, db *gorm.DB) {
		prepareTest(db,
			&APIKey{},
			&Account{}, &User{}, &Organization{},
			&Application{}, &Client{}, &Gateway{},
		)

		s := newStore(db)
		store := GetAPIKeyStore(db)

		s.createEntity(ctx, &User{Account: Account{UID: "test-user"}})
		userIDs := &ttnpb.UserIdentifiers{UserId: "test-user"}

		s.createEntity(ctx, &Organization{Account: Account{UID: "test-org"}})
		orgIDs := &ttnpb.OrganizationIdentifiers{OrganizationId: "test-org"}

		s.createEntity(ctx, &Application{ApplicationID: "test-app"})
		appIDs := &ttnpb.ApplicationIdentifiers{ApplicationId: "test-app"}

		s.createEntity(ctx, &Gateway{GatewayID: "test-gtw"})
		gtwIDs := &ttnpb.GatewayIdentifiers{GatewayId: "test-gtw"}

		expiryTime := time.Now().Add(24 * time.Hour)

		for _, tt := range []struct {
			Name        string
			Identifiers *ttnpb.EntityIdentifiers
			Rights      []ttnpb.Right
		}{
			{
				Name:        "Application",
				Identifiers: appIDs.GetEntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
			},
			{
				Name:        "Gateway",
				Identifiers: gtwIDs.GetEntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
			},
			{
				Name:        "Organization",
				Identifiers: orgIDs.GetEntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL, ttnpb.RIGHT_GATEWAY_ALL},
			},
			{
				Name:        "User",
				Identifiers: userIDs.GetEntityIdentifiers(),
				Rights:      []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL, ttnpb.RIGHT_GATEWAY_ALL},
			},
		} {
			t.Run(tt.Name, func(t *testing.T) {
				a := assertions.New(t)

				key := &ttnpb.APIKey{
					ID:        strings.ToUpper(fmt.Sprintf("%sKEYID", tt.Name)),
					Key:       strings.ToUpper(fmt.Sprintf("%sKEY", tt.Name)),
					Name:      fmt.Sprintf("%s API key", tt.Name),
					Rights:    tt.Rights,
					ExpiresAt: &expiryTime,
				}
				created, err := store.CreateAPIKey(ctx, tt.Identifiers, key)

				a.So(err, should.BeNil)
				if a.So(created, should.NotBeNil) {
					a.So(created.ID, should.Equal, key.ID)
					a.So(created.Key, should.Equal, key.Key)
					a.So(created.Name, should.Equal, key.Name)
					a.So(created.Rights, should.Resemble, key.Rights)
					a.So(created.CreatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
					a.So(created.UpdatedAt, should.HappenAfter, time.Now().Add(-1*time.Hour))
				}

				keys, err := store.FindAPIKeys(ctx, tt.Identifiers)

				a.So(err, should.BeNil)
				if a.So(keys, should.HaveLength, 1) {
					a.So(keys[0], should.Resemble, created)
				}

				ids, got, err := store.GetAPIKey(ctx, key.ID)

				a.So(err, should.BeNil)
				a.So(ids, should.Resemble, tt.Identifiers)
				a.So(got, should.Resemble, created)

				updated, err := store.UpdateAPIKey(ctx, tt.Identifiers, &ttnpb.APIKey{
					ID:     strings.ToUpper(fmt.Sprintf("%sKEYID", tt.Name)),
					Name:   fmt.Sprintf("Updated %s API key", tt.Name),
					Rights: tt.Rights,
				},
					&pbtypes.FieldMask{Paths: []string{"rights", "name"}})

				a.So(err, should.BeNil)

				ids, got, err = store.GetAPIKey(ctx, key.ID)

				a.So(err, should.BeNil)
				a.So(ids, should.Resemble, tt.Identifiers)
				if a.So(got, should.NotBeNil) {
					a.So(got, should.Resemble, updated)
					a.So(got.Name, should.NotEqual, key.Name)
					a.So(got.Rights, should.Resemble, key.Rights)
				}

				updated, err = store.UpdateAPIKey(ctx, tt.Identifiers, &ttnpb.APIKey{
					ID: strings.ToUpper(fmt.Sprintf("%sKEYID", tt.Name)),
					// Empty rights,
				},
					&pbtypes.FieldMask{Paths: []string{"rights"}})

				a.So(err, should.BeNil)
				a.So(updated, should.BeNil)

				_, _, err = store.GetAPIKey(ctx, key.ID)

				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}

				keys, err = store.FindAPIKeys(ctx, tt.Identifiers)

				a.So(err, should.BeNil)
				a.So(keys, should.HaveLength, 0)
			})
		}

		t.Run("Delete entity API keys", func(t *testing.T) {
			a := assertions.New(t)

			// Create three API keys for entity
			for i := 0; i < 3; i++ {
				key := &ttnpb.APIKey{
					ID:     strings.ToUpper(fmt.Sprintf("KEYID%d", i)),
					Key:    strings.ToUpper(fmt.Sprintf("KEY%d", i)),
					Name:   fmt.Sprintf("API key %d", i),
					Rights: []ttnpb.Right{ttnpb.RIGHT_APPLICATION_ALL},
				}

				store.CreateAPIKey(ctx, userIDs.GetEntityIdentifiers(), key)
			}
			err := store.DeleteEntityAPIKeys(ctx, userIDs.GetEntityIdentifiers())

			a.So(err, should.BeNil)

			keys, err := store.FindAPIKeys(ctx, userIDs.GetEntityIdentifiers())

			// Check if all API keys are deleted
			a.So(err, should.BeNil)
			a.So(keys, should.HaveLength, 0)
		})
	})
}
