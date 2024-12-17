// Copyright © 2022 The Things Network Foundation, The Things Industries B.V.
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

package identityserver

import (
	"testing"
	"time"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/storetest"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/types/known/timestamppb"
)

const noOfGateways = 3

func TestGatewaysPermissions(t *testing.T) {
	p := &storetest.Population{}
	usr1 := p.NewUser()
	gtw1 := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		t.Run("Invalid credentials", func(t *testing.T) { // nolint:paralleltest
			_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"},
				},
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: gtw1.GetIds(),
				FieldMask:  ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsUnauthenticated(err), should.BeTrue)
			}

			listRes, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask: ttnpb.FieldMask("name"),
			})
			a.So(err, should.BeNil)
			if a.So(listRes, should.NotBeNil) {
				a.So(listRes.Gateways, should.BeEmpty)
			}

			_, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
				FieldMask:    ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids:  gtw1.GetIds(),
					Name: "Updated Name",
				},
				FieldMask: ttnpb.FieldMask("name"),
			})
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Delete(ctx, gtw1.GetIds())
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})

		t.Run("Admin read-only", func(t *testing.T) { // nolint:paralleltest
			_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids: &ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"},
				},
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			}, readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: gtw1.GetIds(),
				FieldMask:  ttnpb.FieldMask("name"),
			}, readOnlyAdminKeyCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask: ttnpb.FieldMask("name"),
			}, readOnlyAdminKeyCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
				Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
				FieldMask:    ttnpb.FieldMask("name"),
			}, readOnlyAdminKeyCreds)
			a.So(errors.IsPermissionDenied(err), should.BeFalse)

			_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids:  gtw1.GetIds(),
					Name: "Updated Name",
				},
				FieldMask: ttnpb.FieldMask("name"),
			}, readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}

			_, err = reg.Delete(ctx, gtw1.GetIds(), readOnlyAdminKeyCreds)
			if a.So(err, should.NotBeNil) {
				a.So(errors.IsPermissionDenied(err), should.BeTrue)
			}
		})
	}, withPrivateTestDatabase(p))
}

func TestGatewaysCRUD(t *testing.T) {
	p := &storetest.Population{}

	adminUsr := p.NewUser()
	adminUsr.Admin = true
	adminKey, _ := p.NewAPIKey(adminUsr.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	adminCreds := rpcCreds(adminKey)

	usr1 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	}

	usr2 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewGateway(usr2.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)
	keyWithoutRights, _ := p.NewAPIKey(usr1.GetEntityIdentifiers())
	credsWithoutRights := rpcCreds(keyWithoutRights)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		eui := &types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

		is.config.UserRights.CreateGateways = false

		_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: &ttnpb.Gateway{
				Ids: &ttnpb.GatewayIdentifiers{
					GatewayId: "foo",
					Eui:       eui.Bytes(),
				},
				Name: "Foo Gateway",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		is.config.UserRights.CreateGateways = true

		created, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: &ttnpb.Gateway{
				Ids: &ttnpb.GatewayIdentifiers{
					GatewayId: "foo",
					Eui:       eui.Bytes(),
				},
				Name: "Foo Gateway",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.BeNil) && a.So(created, should.NotBeNil) {
			a.So(created.GetIds().GetEui(), should.Resemble, eui.Bytes())
			a.So(created.Name, should.Equal, "Foo Gateway")
		}

		got, err := reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIds: created.GetIds(),
			FieldMask:  ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got.GetIds().GetEui(), should.Resemble, created.Ids.Eui)
			a.So(got.Name, should.Equal, created.Name)
		}

		ids, err := reg.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			Eui: eui.Bytes(),
		}, credsWithoutRights)
		if a.So(err, should.BeNil) && a.So(ids, should.NotBeNil) {
			a.So(ids.GetGatewayId(), should.Equal, created.GetIds().GetGatewayId())
		}

		_, err = reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: &ttnpb.Gateway{
				Ids: &ttnpb.GatewayIdentifiers{
					GatewayId: "bar",
					Eui:       eui.Bytes(),
				},
				Name: "Bar Gateway",
			},
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, creds)
		if a.So(err, should.NotBeNil) {
			a.So(err, should.HaveSameErrorDefinitionAs, errGatewayEUITaken)
		}

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIds: created.GetIds(),
			FieldMask:  ttnpb.FieldMask("ids"),
		}, credsWithoutRights)
		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIds: created.GetIds(),
			FieldMask:  ttnpb.FieldMask("attributes"),
		}, credsWithoutRights)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		// TODO: Remove (https://github.com/TheThingsNetwork/lorawan-stack/issues/6804)
		t.Run("Contact_info fieldmask", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)
			got, err := reg.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: created.GetIds(),
				FieldMask:  ttnpb.FieldMask("contact_info"),
			}, creds)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.ContactInfo, should.HaveLength, 2)
				a.So(got.ContactInfo[0].Value, should.Equal, usr1.PrimaryEmailAddress)
				a.So(got.ContactInfo[0].ContactType, should.Equal, ttnpb.ContactType_CONTACT_TYPE_OTHER)
				a.So(got.ContactInfo[1].Value, should.Equal, usr1.PrimaryEmailAddress)
				a.So(got.ContactInfo[1].ContactType, should.Equal, ttnpb.ContactType_CONTACT_TYPE_TECHNICAL)
			}

			// Testing the `PublicSafe` method, which should not return the contact_info's email address when the caller
			// does not have the appropriate rights.
			got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
				GatewayIds: created.GetIds(),
				FieldMask:  ttnpb.FieldMask("contact_info"),
			}, credsWithoutRights)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				a.So(got.ContactInfo, should.HaveLength, 0)
			}
		})

		updated, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: &ttnpb.Gateway{
				Ids:  created.GetIds(),
				Name: "Updated Name",
			},
			FieldMask: ttnpb.FieldMask("name"),
		}, creds)
		if a.So(err, should.BeNil) && a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		t.Run("Contact Info Restrictions", func(t *testing.T) { // nolint:paralleltest
			a, ctx := test.New(t)

			oldSetOtherAsContacts := is.config.CollaboratorRights.SetOthersAsContacts
			t.Cleanup(func() { is.config.CollaboratorRights.SetOthersAsContacts = oldSetOtherAsContacts })
			is.config.CollaboratorRights.SetOthersAsContacts = false

			// Set usr-2 as collaborator to client.
			gac := ttnpb.NewGatewayAccessClient(cc)
			_, err := gac.SetCollaborator(ctx, &ttnpb.SetGatewayCollaboratorRequest{
				GatewayIds: created.GetIds(),
				Collaborator: &ttnpb.Collaborator{
					Ids:    usr2.GetOrganizationOrUserIdentifiers(),
					Rights: []ttnpb.Right{ttnpb.Right_RIGHT_ALL},
				},
			}, creds)
			a.So(err, should.BeNil)

			// Attempt to set another collaborator as administrative contact.
			_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids:                   created.GetIds(),
					AdministrativeContact: usr2.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("administrative_contact"),
			}, creds)
			a.So(errors.IsPermissionDenied(err), should.BeTrue)

			// Admin can bypass contact info restrictions.
			_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids:                   created.GetIds(),
					AdministrativeContact: usr1.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("administrative_contact"),
			}, adminCreds)
			a.So(err, should.BeNil)

			is.config.CollaboratorRights.SetOthersAsContacts = true

			// Now usr-1 can set usr-2 as technical contact.
			_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
				Gateway: &ttnpb.Gateway{
					Ids:              created.GetIds(),
					TechnicalContact: usr2.GetOrganizationOrUserIdentifiers(),
				},
				FieldMask: ttnpb.FieldMask("technical_contact"),
			}, creds)
			a.So(err, should.BeNil)
		})

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{
			nil, usr1.GetOrganizationOrUserIdentifiers(),
		} {
			list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask:    ttnpb.FieldMask("name"),
				Collaborator: collaborator,
			}, creds)
			if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) && a.So(list.Gateways, should.HaveLength, 6) {
				var found bool
				for _, item := range list.Gateways {
					if item.GetIds().GetGatewayId() == created.GetIds().GetGatewayId() {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, created.GetIds(), creds)
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, created.GetIds(), creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, created.GetIds(), adminCreds)
		a.So(err, should.BeNil)
	}, withPrivateTestDatabase(p))
}

func TestGatewaysPagination(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	for i := 0; i < 3; i++ {
		p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		var md metadata.MD

		list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds, grpc.Header(&md))
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 2)
			a.So(md.Get("x-total-count"), should.Resemble, []string{"3"})
		}

		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 1)
		}

		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.BeEmpty)
		}
	}, withPrivateTestDatabase(p))
}

func TestGatewayBatchOperations(t *testing.T) {
	t.Parallel()
	a, ctx := test.New(t)
	p := &storetest.Population{}
	usr1 := p.NewUser()
	usr2 := p.NewUser()
	gtwIDs := make([]*ttnpb.GatewayIdentifiers, 0, noOfGateways)
	for i := 0; i < noOfGateways; i++ {
		gtw := p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
		gtw.Attributes = map[string]string{
			"foo": "bar",
		}
		gtw.UpdateLocationFromStatus = true
		gtwIDs = append(gtwIDs, gtw.GetIds())
	}
	limitedKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_USER_GATEWAYS_LIST)
	limitedCreds := rpcCreds(limitedKey)

	fullKey, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_GATEWAY_ALL)
	fullCreds := rpcCreds(fullKey)

	usr2Key, _ := p.NewAPIKey(usr2.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	usr2Creds := rpcCreds(usr2Key)

	readOnlyAdmin := p.NewUser()
	readOnlyAdmin.Admin = true
	readOnlyAdminKey, _ := p.NewAPIKey(readOnlyAdmin.GetEntityIdentifiers(), ttnpb.AllReadAdminRights.GetRights()...)
	readOnlyAdminKeyCreds := rpcCreds(readOnlyAdminKey)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayBatchRegistryClient(cc)
		readReg := ttnpb.NewGatewayRegistryClient(cc)

		// Read after create.
		gtws, err := readReg.List(ctx, &ttnpb.ListGatewaysRequest{
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, limitedCreds)
		a.So(err, should.BeNil)
		a.So(len(gtws.Gateways), should.Equal, noOfGateways)

		// ClusterAuth.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: gtwIDs,
		}, is.WithClusterAuth())
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// Insufficient rights.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: gtwIDs,
		}, limitedCreds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// Read-only admin rights.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: gtwIDs,
		}, readOnlyAdminKeyCreds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// User without rights on gateways.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: gtwIDs,
		}, usr2Creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// Unknown gateway ID.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: []*ttnpb.GatewayIdentifiers{
				{
					GatewayId: "unknown",
				},
			},
		}, fullCreds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// One unknown in batch.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: []*ttnpb.GatewayIdentifiers{
				{
					GatewayId: "unknown",
				},
				gtwIDs[0],
				gtwIDs[1],
			},
		}, fullCreds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		// Valid Batch.
		_, err = reg.Delete(ctx, &ttnpb.BatchDeleteGatewaysRequest{
			GatewayIds: gtwIDs,
		}, fullCreds)
		a.So(err, should.BeNil)

		// Read after delete.
		gtws, err = readReg.List(ctx, &ttnpb.ListGatewaysRequest{
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
		}, limitedCreds)
		a.So(err, should.BeNil)
		a.So(len(gtws.Gateways), should.Equal, 0)
	}, withPrivateTestDatabase(p))
}

func TestGatewaysFilter(t *testing.T) {
	p := &storetest.Population{}

	usr1 := p.NewUser()
	for i := 0; i < 5; i++ {
		p.NewGateway(usr1.GetOrganizationOrUserIdentifiers())
	}

	key, _ := p.NewAPIKey(usr1.GetEntityIdentifiers(), ttnpb.Right_RIGHT_ALL)
	creds := rpcCreds(key)

	t.Parallel()
	a, ctx := test.New(t)

	testWithIdentityServer(t, func(_ *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		// Filter by 1 hour ago.
		list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Filters: []*ttnpb.ListGatewaysRequest_Filter{
				{
					Field: &ttnpb.ListGatewaysRequest_Filter_UpdatedSince{
						UpdatedSince: timestamppb.New(time.Now().Add(-time.Hour)),
					},
				},
			},
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 5)
		}

		// Filter by now. The timestamp is newer then the last `update_at`, so no devices should be returned.
		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Filters: []*ttnpb.ListGatewaysRequest_Filter{
				{
					Field: &ttnpb.ListGatewaysRequest_Filter_UpdatedSince{
						UpdatedSince: timestamppb.New(time.Now()),
					},
				},
			},
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 0)
		}

		// Filter by 1 hour ago with pagination.
		list, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask:    ttnpb.FieldMask("name"),
			Collaborator: usr1.GetOrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
			Filters: []*ttnpb.ListGatewaysRequest_Filter{
				{
					Field: &ttnpb.ListGatewaysRequest_Filter_UpdatedSince{
						UpdatedSince: timestamppb.New(time.Now().Add(-time.Hour)),
					},
				},
			},
		}, creds)
		if a.So(err, should.BeNil) && a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 2)
		}
	}, withPrivateTestDatabase(p))
}
