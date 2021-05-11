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
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
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

			for _, devName := range []string{"baz", "qux"} {
				store.createEntity(ctx, &EndDevice{
					ApplicationID: fmt.Sprintf("the-%s-app", name),
					DeviceID:      fmt.Sprintf("the-%s-device", devName),
					Name:          fmt.Sprintf("The %s device in %s", devName, name),
					Description:   fmt.Sprintf("This device does %s stuff for %s", devName, name),
					Attributes: []Attribute{
						{Key: "test", Value: devName},
					},
				})
			}

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

		t.Run("Application", func(t *testing.T) {
			ids, err := s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				IDContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				NameContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				DescriptionContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("Client", func(t *testing.T) {
			ids, err := s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				IDContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				NameContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				DescriptionContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("Gateway", func(t *testing.T) {
			ids, err := s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				IDContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				NameContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				DescriptionContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("Organization", func(t *testing.T) {
			ids, err := s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				IDContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				NameContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				DescriptionContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("User", func(t *testing.T) {
			ids, err := s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				IDContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				NameContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				DescriptionContains: "foo",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		store.deleteEntity(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: fmt.Sprintf("the-foo-app")})
		store.deleteEntity(ctx, &ttnpb.ClientIdentifiers{ClientId: fmt.Sprintf("the-foo-cli")})
		store.deleteEntity(ctx, &ttnpb.GatewayIdentifiers{GatewayId: fmt.Sprintf("the-foo-gtw")})
		store.deleteEntity(ctx, &ttnpb.UserIdentifiers{UserId: fmt.Sprintf("the-foo-usr")})
		store.deleteEntity(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: fmt.Sprintf("the-foo-org")})

		t.Run("deleted application", func(t *testing.T) {
			ids, err := s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 0)

			ids, err = s.FindApplications(ctx, nil, &ttnpb.SearchApplicationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
				Deleted: true,
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("deleted client", func(t *testing.T) {
			ids, err := s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 0)

			ids, err = s.FindClients(ctx, nil, &ttnpb.SearchClientsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
				Deleted: true,
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("deleted gateway", func(t *testing.T) {
			ids, err := s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 0)

			ids, err = s.FindGateways(ctx, nil, &ttnpb.SearchGatewaysRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
				Deleted: true,
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("deleted organization", func(t *testing.T) {
			ids, err := s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 0)

			ids, err = s.FindOrganizations(ctx, nil, &ttnpb.SearchOrganizationsRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
				Deleted: true,
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("deleted user", func(t *testing.T) {
			ids, err := s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 0)

			ids, err = s.FindUsers(ctx, nil, &ttnpb.SearchUsersRequest{
				IDContains:          "foo",
				NameContains:        "foo",
				DescriptionContains: "foo",
				AttributesContain: map[string]string{
					"test": "foo",
				},
				Deleted: true,
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})

		t.Run("end_device", func(t *testing.T) {
			ids, err := s.FindEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "the-foo-app"},
				IDContains:             "baz",
				NameContains:           "baz",
				DescriptionContains:    "baz",
				AttributesContain: map[string]string{
					"test": "baz",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "the-foo-app"},
				IDContains:             "baz",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "the-foo-app"},
				NameContains:           "baz",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "the-foo-app"},
				DescriptionContains:    "baz",
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)

			ids, err = s.FindEndDevices(ctx, &ttnpb.SearchEndDevicesRequest{
				ApplicationIdentifiers: ttnpb.ApplicationIdentifiers{ApplicationId: "the-foo-app"},
				AttributesContain: map[string]string{
					"test": "baz",
				},
			})

			a.So(err, should.BeNil)
			a.So(ids, should.HaveLength, 1)
		})
	})
}
