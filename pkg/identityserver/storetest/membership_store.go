// Copyright © 2021 The Things Network Foundation, The Things Industries B.V.
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

package storetest

import (
	. "testing"

	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func (st *StoreTest) TestMembershipStoreCRUD(t *T) {
	a, ctx := test.New(t)

	app1 := st.population.NewApplication(nil)
	cli1 := st.population.NewClient(nil)
	gtw1 := st.population.NewGateway(nil)

	org1 := st.population.NewOrganization(nil)
	org2 := st.population.NewOrganization(nil)

	usr1 := st.population.NewUser()
	usr2 := st.population.NewUser()
	usr3 := st.population.NewUser()

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.MembershipStore
	})
	defer st.DestroyDB(t, true, "applications", "clients", "gateways", "organizations", "users", "accounts")
	defer s.Close()
	if !ok {
		t.Fatal("Store does not implement MembershipStore")
	}

	someRights := map[string]*ttnpb.Rights{
		app1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_ALL),
		cli1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_CLIENT_ALL),
		gtw1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_ALL),
		org1.EntityType(): ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_ALL),
	}
	allRights := ttnpb.RightsFrom(ttnpb.Right_RIGHT_ALL)

	err := s.SetMember(ctx, usr2.GetOrganizationOrUserIdentifiers(), org2.GetEntityIdentifiers(), allRights)
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}

	for _, ids := range []*ttnpb.EntityIdentifiers{
		app1.GetEntityIdentifiers(),
		cli1.GetEntityIdentifiers(),
		gtw1.GetEntityIdentifiers(),
		org1.GetEntityIdentifiers(),
	} {
		t.Run(ids.EntityType(), func(t *T) {
			t.Run("SetMember", func(t *T) {
				a, ctx := test.New(t)
				err := s.SetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids, someRights[ids.EntityType()])
				a.So(err, should.BeNil)
			})

			t.Run("GetMember", func(t *T) {
				a, ctx := test.New(t)
				rights, err := s.GetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids)
				if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
					a.So(rights, should.Resemble, someRights[ids.EntityType()])
				}
			})

			t.Run("FindMembers", func(t *T) {
				a, ctx := test.New(t)
				members, err := s.FindMembers(ctx, ids)
				if a.So(err, should.BeNil) && a.So(members, should.NotBeNil) && a.So(members, should.HaveLength, 1) {
					for memberIDs, rights := range members {
						a.So(memberIDs, should.Resemble, usr1.GetEntityIdentifiers())
						a.So(rights, should.Resemble, someRights[ids.EntityType()])
					}
				}
			})

			t.Run("FindMemberships", func(t *T) {
				a, ctx := test.New(t)
				res, err := s.FindMemberships(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids.EntityType(), false)
				if a.So(err, should.BeNil) && a.So(res, should.NotBeNil) && a.So(res, should.HaveLength, 1) {
					a.So(res[0], should.Resemble, ids)
				}
			})

			t.Run("SetMember_Update", func(t *T) {
				a, ctx := test.New(t)
				err := s.SetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids, allRights)
				a.So(err, should.BeNil)
			})

			t.Run("GetMember_AfterUpdate", func(t *T) {
				a, ctx := test.New(t)
				rights, err := s.GetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids)
				if a.So(err, should.BeNil) && a.So(rights, should.NotBeNil) {
					a.So(rights, should.Resemble, allRights)
				}
			})

			if ids.EntityType() != "organization" {
				t.Run("SetMember_Organization", func(t *T) {
					a, ctx := test.New(t)
					err := s.SetMember(ctx, org2.GetOrganizationOrUserIdentifiers(), ids, allRights)
					a.So(err, should.BeNil)
				})

				t.Run("FindMemberships_Indirect", func(t *T) {
					a, ctx := test.New(t)
					res, err := s.FindMemberships(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids.EntityType(), true)
					if a.So(err, should.BeNil) && a.So(res, should.NotBeNil) && a.So(res, should.HaveLength, 1) {
						a.So(res[0], should.Resemble, ids)
					}
				})

				t.Run("FindAccountMembershipChains_One", func(t *T) {
					a, ctx := test.New(t)
					chains, err := s.FindAccountMembershipChains(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids.EntityType())
					if a.So(err, should.BeNil) && a.So(chains, should.NotBeNil) && a.So(chains, should.HaveLength, 1) {
						for _, chain := range chains {
							a.So(chain.UserIdentifiers, should.Resemble, usr2.GetIds())
							a.So(chain.RightsOnOrganization, should.Resemble, allRights)
							a.So(chain.OrganizationIdentifiers, should.Resemble, org2.GetIds())
							a.So(chain.RightsOnEntity, should.Resemble, allRights)
							a.So(chain.EntityIdentifiers, should.Resemble, ids)
						}
					}
					selectedChains, err := s.FindAccountMembershipChains(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids.EntityType(), ids.IDString())
					if a.So(err, should.BeNil) && a.So(selectedChains, should.NotBeNil) && a.So(selectedChains, should.HaveLength, 1) {
						a.So(selectedChains[0], should.Resemble, chains[0])
					}
				})
			}

			t.Run("SetMember_Other", func(t *T) {
				a, ctx := test.New(t)
				err := s.SetMember(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids, allRights)
				a.So(err, should.BeNil)
			})

			if ids.EntityType() != "organization" {
				t.Run("FindAccountMembershipChains_Two", func(t *T) {
					a, ctx := test.New(t)
					chains, err := s.FindAccountMembershipChains(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids.EntityType())
					if a.So(err, should.BeNil) && a.So(chains, should.NotBeNil) && a.So(chains, should.HaveLength, 2) {
						for _, chain := range chains {
							a.So(chain.UserIdentifiers, should.Resemble, usr2.GetIds())
							if chain.OrganizationIdentifiers != nil {
								a.So(chain.RightsOnOrganization, should.Resemble, allRights)
								a.So(chain.OrganizationIdentifiers, should.Resemble, org2.GetIds())
								a.So(chain.RightsOnEntity, should.Resemble, allRights)
							} else {
								a.So(chain.RightsOnEntity, should.Resemble, allRights)
							}
							a.So(chain.EntityIdentifiers, should.Resemble, ids)
						}
					}
					selectedChains, err := s.FindAccountMembershipChains(ctx, usr2.GetOrganizationOrUserIdentifiers(), ids.EntityType(), ids.IDString())
					if a.So(err, should.BeNil) && a.So(selectedChains, should.NotBeNil) && a.So(selectedChains, should.HaveLength, 2) {
						a.So(selectedChains[0], should.Resemble, chains[0])
					}
				})
			}

			t.Run("SetMember_Delete", func(t *T) {
				a, ctx := test.New(t)
				err := s.SetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids, &ttnpb.Rights{})
				a.So(err, should.BeNil)
			})

			t.Run("GetMember_AfterDelete", func(t *T) {
				a, ctx := test.New(t)
				_, err := s.GetMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("DeleteAccountMembers", func(t *T) {
				a, ctx := test.New(t)

				err := s.SetMember(ctx, usr3.GetOrganizationOrUserIdentifiers(), ids, someRights[ids.EntityType()])
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				err = s.DeleteAccountMembers(ctx, usr3.GetOrganizationOrUserIdentifiers())
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				_, err = s.GetMember(ctx, usr3.GetOrganizationOrUserIdentifiers(), ids)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})

			t.Run("DeleteEntityMembers", func(t *T) {
				a, ctx := test.New(t)

				err := s.SetMember(ctx, usr3.GetOrganizationOrUserIdentifiers(), ids, someRights[ids.EntityType()])
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				err = s.DeleteEntityMembers(ctx, ids)
				if !a.So(err, should.BeNil) {
					t.FailNow()
				}

				_, err = s.GetMember(ctx, usr3.GetOrganizationOrUserIdentifiers(), ids)
				if a.So(err, should.NotBeNil) {
					a.So(errors.IsNotFound(err), should.BeTrue)
				}
			})
		})
	}

	err = s.SetMember(ctx, usr2.GetOrganizationOrUserIdentifiers(), org2.GetEntityIdentifiers(), &ttnpb.Rights{})
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
}

// TODO: Test Pagination (https://github.com/TheThingsNetwork/lorawan-stack/issues/5047).
