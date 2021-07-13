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

package identityserver

import (
	"testing"
	"time"

	pbtypes "github.com/gogo/protobuf/types"
	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/types"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
	"google.golang.org/grpc"
)

func init() {
	// remove gateways assigned to the user by the populator
	userID := paginationUser.UserIdentifiers
	for _, gw := range population.Gateways {
		for id, collaborators := range population.Memberships {
			if gw.IDString() == id.IDString() {
				for i, collaborator := range collaborators {
					if collaborator.IDString() == userID.GetUserId() {
						collaborators = collaborators[:i+copy(collaborators[i:], collaborators[i+1:])]
					}
				}
			}
		}
	}

	// add deterministic number of gateways
	for i := 0; i < 3; i++ {
		gatewayID := population.Gateways[i].GetEntityIdentifiers()
		population.Memberships[gatewayID] = append(population.Memberships[gatewayID], &ttnpb.Collaborator{
			OrganizationOrUserIdentifiers: *paginationUser.OrganizationOrUserIdentifiers(),
			Rights:                        []ttnpb.Right{ttnpb.RIGHT_GATEWAY_ALL},
		})
	}
}

func TestGatewaysPermissionDenied(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"},
			},
			Collaborator: *ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"},
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsUnauthenticated(err), should.BeTrue)
		}

		listRes, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		a.So(err, should.BeNil)
		if a.So(listRes, should.NotBeNil) {
			a.So(listRes.Gateways, should.BeEmpty)
		}

		_, err = reg.List(ctx, &ttnpb.ListGatewaysRequest{
			Collaborator: ttnpb.UserIdentifiers{UserId: "foo-usr"}.OrganizationOrUserIdentifiers(),
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"},
				Name:               "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Delete(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"})

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}
	})
}

func TestGatewaysCRUD(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		userID, creds := population.Users[defaultUserIdx].UserIdentifiers, userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		eui := types.EUI64{1, 2, 3, 4, 5, 6, 7, 8}

		is.config.UserRights.CreateGateways = false

		_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: "foo",
					Eui:       &eui,
				},
				Name: "Foo Gateway",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		is.config.UserRights.CreateGateways = true

		created, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: "foo",
					Eui:       &eui,
				},
				Name: "Foo Gateway",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, "Foo Gateway")
		}

		got, err := reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
			if a.So(got.Eui, should.NotBeNil) {
				a.So(*got.Eui, should.Equal, eui)
			}
		}

		ids, err := reg.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			Eui: eui,
		}, credsWithoutRights)

		a.So(err, should.BeNil)
		if a.So(ids, should.NotBeNil) {
			a.So(ids.GatewayId, should.Equal, created.GatewayId)
		}

		_, err = reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: "bar",
					Eui:       &eui,
				},
				Name: "Bar Gateway",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(err, should.HaveSameErrorDefinitionAs, errGatewayEUITaken)
		}

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"attributes"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updated, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: created.GatewayIdentifiers,
				Name:               "Updated Name",
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"name"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(updated, should.NotBeNil) {
			a.So(updated.Name, should.Equal, "Updated Name")
		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{nil, userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
				Collaborator: collaborator,
			}, creds)

			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list.Gateways, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Gateways {
					if item.GatewayId == created.GatewayId {
						found = true
						a.So(item.Name, should.Equal, updated.Name)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, &created.GatewayIdentifiers, creds)

		a.So(err, should.BeNil)
	})
}

func TestGatewaysPagination(t *testing.T) {
	a := assertions.New(t)

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		userID := paginationUser.UserIdentifiers
		creds := userCreds(paginationUserIdx)

		reg := ttnpb.NewGatewayRegistryClient(cc)

		list, err := reg.List(test.Context(), &ttnpb.ListGatewaysRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         1,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 2)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListGatewaysRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         2,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.HaveLength, 1)
		}

		list, err = reg.List(test.Context(), &ttnpb.ListGatewaysRequest{
			FieldMask:    &pbtypes.FieldMask{Paths: []string{"name"}},
			Collaborator: userID.OrganizationOrUserIdentifiers(),
			Limit:        2,
			Page:         3,
		}, creds)

		a.So(err, should.BeNil)
		if a.So(list, should.NotBeNil) {
			a.So(list.Gateways, should.BeEmpty)
		}
	})
}

func TestGatewaysSecrets(t *testing.T) {
	a := assertions.New(t)
	ctx := test.Context()

	testWithIdentityServer(t, func(is *IdentityServer, cc *grpc.ClientConn) {
		reg := ttnpb.NewGatewayRegistryClient(cc)

		userID, creds := population.Users[defaultUserIdx].UserIdentifiers, userCreds(defaultUserIdx)
		credsWithoutRights := userCreds(defaultUserIdx, "key without rights")

		eui := types.EUI64{0x11, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}

		targetCUPSURI := "https://thethings.example.com"
		otherTargetCUPSURI := "https://thenotthings.example.com:1234"

		gatewayID := "foo-with-secrets"
		gatewayName := "Foo Gateway with Secrets"

		from := time.Now()
		to := from.Add(5 * time.Minute)

		gtwClaimAuthCode := ttnpb.GatewayClaimAuthenticationCode{
			ValidFrom: &from,
			ValidTo:   &to,
			Secret: &ttnpb.Secret{
				KeyId: "is-test",
				Value: []byte("my claim auth code"),
			},
		}

		otherGtwClaimAuthCode := ttnpb.GatewayClaimAuthenticationCode{
			ValidFrom: &from,
			ValidTo:   &to,
			Secret: &ttnpb.Secret{
				KeyId: "is-test",
				Value: []byte("my other claim auth code"),
			},
		}

		secret := &ttnpb.Secret{
			KeyId: "is-test",
			Value: []byte("my very secret value"),
		}

		is.config.UserRights.CreateGateways = false

		_, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: gatewayID,
					Eui:       &eui,
				},
				Name: gatewayName,
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)
		a.So(errors.IsPermissionDenied(err), should.BeTrue)

		is.config.UserRights.CreateGateways = true

		// Plaintext
		euiWithoutEncKey := types.EUI64{0x22, 0x22, 0x33, 0x44, 0x55, 0x66, 0x77, 0x88}
		gatewayIDWithoutEncKey := "foo-without-encryption-key"
		gatewayNameWithoutEncKey := "Foo Gateway without encryption key"

		createdWithoutEncKey, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: gatewayIDWithoutEncKey,
					Eui:       &euiWithoutEncKey,
				},
				Name:                    gatewayNameWithoutEncKey,
				LBSLNSSecret:            secret,
				ClaimAuthenticationCode: &gtwClaimAuthCode,
				TargetCUPSURI:           targetCUPSURI,
				TargetCUPSKey:           secret,
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(createdWithoutEncKey, should.NotBeNil) {
			a.So(createdWithoutEncKey.Name, should.Equal, gatewayNameWithoutEncKey)
			a.So(createdWithoutEncKey.LBSLNSSecret, should.NotBeNil)
			a.So(createdWithoutEncKey.ClaimAuthenticationCode, should.NotBeNil)
			a.So(createdWithoutEncKey.TargetCUPSKey, should.NotBeNil)
		}

		got, err := reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: createdWithoutEncKey.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"name", "lbs_lns_secret", "claim_authentication_code", "target_cups_uri", "target_cups_key"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, createdWithoutEncKey.Name)
			if a.So(got.Eui, should.NotBeNil) {
				a.So(*got.Eui, should.Equal, euiWithoutEncKey)
			}
			a.So(got.LBSLNSSecret.Value, should.Resemble, secret.Value)
			a.So(got.ClaimAuthenticationCode, should.NotBeNil)
			a.So(got.ClaimAuthenticationCode.Secret, should.NotBeNil)
			a.So(got.ClaimAuthenticationCode.Secret.Value, should.Resemble, gtwClaimAuthCode.Secret.Value)
			a.So(got.TargetCUPSKey.Value, should.Resemble, secret.Value)
			a.So(got.TargetCUPSURI, should.Equal, targetCUPSURI)
		}

		// With Encryption Key
		is.config.Gateways.EncryptionKeyID = "is-test"

		created, err := reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: gatewayID,
					Eui:       &eui,
				},
				Name:                    gatewayName,
				LBSLNSSecret:            secret,
				ClaimAuthenticationCode: &gtwClaimAuthCode,
				TargetCUPSURI:           targetCUPSURI,
				TargetCUPSKey:           secret,
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		a.So(err, should.BeNil)
		if a.So(created, should.NotBeNil) {
			a.So(created.Name, should.Equal, gatewayName)
			a.So(created.LBSLNSSecret, should.NotBeNil)
			a.So(created.ClaimAuthenticationCode, should.NotBeNil)
			a.So(createdWithoutEncKey.TargetCUPSKey, should.NotBeNil)
		}

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"name", "lbs_lns_secret", "claim_authentication_code", "target_cups_uri", "target_cups_key"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
			if a.So(got.Eui, should.NotBeNil) {
				a.So(*got.Eui, should.Equal, eui)
			}
			a.So(got.LBSLNSSecret, should.Resemble, secret)
			a.So(got.ClaimAuthenticationCode, should.NotBeNil)
			a.So(got.ClaimAuthenticationCode.Secret, should.NotBeNil)
			a.So(got.ClaimAuthenticationCode.Secret.Value, should.Resemble, gtwClaimAuthCode.Secret.Value)
			a.So(got.TargetCUPSKey.Value, should.Resemble, secret.Value)
			a.So(got.TargetCUPSURI, should.Equal, targetCUPSURI)
		}

		// Check that `claim_authentication_code` can only be updated/retrieved as a whole.
		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"claim_authentication_code.valid_from"}},
		}, creds)
		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.ClaimAuthenticationCode, should.BeNil)
		}
		cacUpdated, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers:      created.GatewayIdentifiers,
				ClaimAuthenticationCode: &otherGtwClaimAuthCode,
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"claim_authentication_code.secret"}},
		}, creds)
		a.So(err, should.BeNil)
		a.So(cacUpdated, should.NotBeNil)
		a.So(cacUpdated.ClaimAuthenticationCode, should.BeNil)

		// Validity check on `claim_authentication_code`.
		validFrom := time.Now().UTC()
		validTo := from.Add(10 * time.Minute)
		cacWithoutSecret, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: created.GatewayIdentifiers,
				ClaimAuthenticationCode: &ttnpb.GatewayClaimAuthenticationCode{
					ValidFrom: &validFrom,
					ValidTo:   &validTo,
				},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"claim_authentication_code"}},
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(cacWithoutSecret, should.BeNil)

		validFrom = time.Now().UTC()
		validTo = from.Add(-20 * time.Minute)
		cacWithoutInvalidTime, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: created.GatewayIdentifiers,
				ClaimAuthenticationCode: &ttnpb.GatewayClaimAuthenticationCode{
					Secret: &ttnpb.Secret{
						Value: []byte("test"),
					},
					ValidFrom: &validFrom,
					ValidTo:   &validTo,
				},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"claim_authentication_code"}},
		}, creds)
		a.So(err, should.NotBeNil)
		a.So(cacWithoutInvalidTime, should.BeNil)

		// Get By EUI
		ids, err := reg.GetIdentifiersForEUI(ctx, &ttnpb.GetGatewayIdentifiersForEUIRequest{
			Eui: eui,
		}, credsWithoutRights)

		a.So(err, should.BeNil)
		if a.So(ids, should.NotBeNil) {
			a.So(ids.GatewayId, should.Equal, created.GatewayId)
		}

		_, err = reg.Create(ctx, &ttnpb.CreateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: ttnpb.GatewayIdentifiers{
					GatewayId: "bar",
					Eui:       &eui,
				},
				Name: "Bar Gateway",
			},
			Collaborator: *userID.OrganizationOrUserIdentifiers(),
		}, creds)

		if a.So(err, should.NotBeNil) {
			a.So(err, should.HaveSameErrorDefinitionAs, errGatewayEUITaken)
		}

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"ids"}},
		}, credsWithoutRights)

		a.So(err, should.BeNil)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"lbs_lns_secret"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"claim_authentication_code"}},
		}, credsWithoutRights)

		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		updatedSecretValue := []byte("my new secret value")

		updated, err := reg.Update(ctx, &ttnpb.UpdateGatewayRequest{
			Gateway: ttnpb.Gateway{
				GatewayIdentifiers: created.GatewayIdentifiers,
				LBSLNSSecret: &ttnpb.Secret{
					Value: updatedSecretValue,
				},
				ClaimAuthenticationCode: &otherGtwClaimAuthCode,
				TargetCUPSURI:           otherTargetCUPSURI,
				TargetCUPSKey: &ttnpb.Secret{
					Value: updatedSecretValue,
				},
			},
			FieldMask: &pbtypes.FieldMask{Paths: []string{"lbs_lns_secret", "claim_authentication_code", "target_cups_key", "target_cups_uri"}},
		}, creds)

		a.So(err, should.BeNil)
		a.So(updated, should.NotBeNil)
		a.So(updated.LBSLNSSecret, should.NotBeNil)
		a.So(updated.LBSLNSSecret.Value, should.Resemble, updatedSecretValue)
		a.So(updated.ClaimAuthenticationCode, should.NotBeNil)
		a.So(updated.ClaimAuthenticationCode.Secret.Value, should.Resemble, otherGtwClaimAuthCode.Secret.Value)
		a.So(updated.TargetCUPSKey, should.NotBeNil)
		a.So(updated.TargetCUPSKey.Value, should.Resemble, updatedSecretValue)

		got, err = reg.Get(ctx, &ttnpb.GetGatewayRequest{
			GatewayIdentifiers: created.GatewayIdentifiers,
			FieldMask:          &pbtypes.FieldMask{Paths: []string{"name", "lbs_lns_secret", "claim_authentication_code", "target_cups_key", "target_cups_uri"}},
		}, creds)

		a.So(err, should.BeNil)
		if a.So(got, should.NotBeNil) {
			a.So(got.Name, should.Equal, created.Name)
			if a.So(got.Eui, should.NotBeNil) {
				a.So(*got.Eui, should.Equal, eui)
			}
			if a.So(got.LBSLNSSecret, should.NotBeNil) {
				a.So(got.LBSLNSSecret.Value, should.Resemble, []byte("my new secret value"))
			}
			if a.So(got.ClaimAuthenticationCode, should.NotBeNil) && a.So(got.ClaimAuthenticationCode.Secret, should.NotBeNil) {
				a.So(got.ClaimAuthenticationCode.Secret.Value, should.Resemble, otherGtwClaimAuthCode.Secret.Value)
			}
			if a.So(got.TargetCUPSKey, should.NotBeNil) {
				a.So(got.TargetCUPSKey.Value, should.Resemble, []byte("my new secret value"))
			}
			a.So(got.TargetCUPSURI, should.Equal, otherTargetCUPSURI)

		}

		for _, collaborator := range []*ttnpb.OrganizationOrUserIdentifiers{userID.OrganizationOrUserIdentifiers()} {
			list, err := reg.List(ctx, &ttnpb.ListGatewaysRequest{
				FieldMask:    &pbtypes.FieldMask{Paths: []string{"lbs_lns_secret", "claim_authentication_code", "target_cups_uri", "target_cups_key"}},
				Collaborator: collaborator,
			}, creds)
			a.So(err, should.BeNil)
			if a.So(list, should.NotBeNil) && a.So(list.Gateways, should.NotBeEmpty) {
				var found bool
				for _, item := range list.Gateways {
					if item.GatewayId == created.GatewayId {
						found = true
						a.So(item.LBSLNSSecret, should.Resemble, got.LBSLNSSecret)
						a.So(item.ClaimAuthenticationCode, should.Resemble, got.ClaimAuthenticationCode)
						a.So(item.TargetCUPSKey, should.Resemble, got.TargetCUPSKey)
						a.So(item.TargetCUPSURI, should.Equal, got.TargetCUPSURI)
					}
				}
				a.So(found, should.BeTrue)
			}
		}

		_, err = reg.Delete(ctx, &createdWithoutEncKey.GatewayIdentifiers, creds)
		a.So(err, should.BeNil)

		_, err = reg.Delete(ctx, &created.GatewayIdentifiers, creds)
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, &created.GatewayIdentifiers, creds)
		if a.So(err, should.NotBeNil) {
			a.So(errors.IsPermissionDenied(err), should.BeTrue)
		}

		_, err = reg.Purge(ctx, &created.GatewayIdentifiers, userCreds(adminUserIdx))
		a.So(err, should.BeNil)

		_, err = reg.Purge(ctx, &createdWithoutEncKey.GatewayIdentifiers, userCreds(adminUserIdx))
		a.So(err, should.BeNil)
	})
}
