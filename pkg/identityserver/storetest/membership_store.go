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
	"sort"
	. "testing"

	"github.com/smarty/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/errors"
	"go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	is "go.thethings.network/lorawan-stack/v3/pkg/identityserver/store"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

type organizationOrUserIdentifiersByID []*ttnpb.OrganizationOrUserIdentifiers

func (a organizationOrUserIdentifiersByID) Len() int      { return len(a) }
func (a organizationOrUserIdentifiersByID) Swap(i, j int) { a[i], a[j] = a[j], a[i] }
func (a organizationOrUserIdentifiersByID) Less(i, j int) bool {
	return a[i].IDString() < a[j].IDString()
}

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
	if !ok {
		t.Skip("Store does not implement MembershipStore")
	}
	defer s.Close()

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
					for _, v := range members {
						memberIDs, rights := v.Ids, v.Rights
						a.So(memberIDs.GetEntityIdentifiers(), should.Resemble, usr1.GetEntityIdentifiers())
						a.So(rights, should.Resemble, someRights[ids.EntityType()])
					}
				}
			})

			t.Run("CountMemberships", func(t *T) {
				a, ctx := test.New(t)
				got, err := s.CountMemberships(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids.EntityType())
				if a.So(err, should.BeNil) {
					a.So(got, should.Equal, 1)
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
				err := s.DeleteMember(ctx, usr1.GetOrganizationOrUserIdentifiers(), ids)
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

	err = s.DeleteMember(ctx, usr2.GetOrganizationOrUserIdentifiers(), org2.GetEntityIdentifiers())
	if !a.So(err, should.BeNil) {
		t.FailNow()
	}
}

func (st *StoreTest) TestMembershipStorePagination(t *T) {
	var apps []*ttnpb.Application
	for i := 0; i < 7; i++ {
		apps = append(apps, st.population.NewApplication(nil))
	}

	var memberIDs []*ttnpb.OrganizationOrUserIdentifiers
	for i := 0; i < 7; i++ {
		ids := st.population.NewUser().GetOrganizationOrUserIdentifiers()
		memberIDs = append(memberIDs, ids)
		st.population.NewMembership(ids, apps[0].GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
	}

	for i := 1; i < 7; i++ {
		st.population.NewMembership(memberIDs[0], apps[i].GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
	}

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.MembershipStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement MembershipStore")
	}
	defer s.Close()

	t.Run("FindMembers_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.FindMembers(paginateCtx, apps[0].GetEntityIdentifiers())
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}
				gotIDs := make([]*ttnpb.OrganizationOrUserIdentifiers, 0, len(got))
				for _, v := range got {
					gotIDs = append(gotIDs, v.Ids)
				}
				sort.Sort(organizationOrUserIdentifiersByID(gotIDs))
				for i, ids := range gotIDs {
					a.So(ids, should.Resemble, memberIDs[i+2*int(page-1)])
				}
			}

			a.So(total, should.Equal, 7)
		}
	})

	t.Run("FindMembers_Ordered", func(t *T) {
		a, ctx := test.New(t)

		for _, tc := range []struct {
			Order    string
			Expected func(*assertions.Assertion, []*is.MemberByID)
		}{
			{
				Order: "",
				Expected: func(a *assertions.Assertion, got []*is.MemberByID) {
					for i, elem := range got {
						a.So(elem.Ids, should.Resemble, memberIDs[i])
						a.So(elem.Rights, should.Resemble, &ttnpb.Rights{
							Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
						})
					}
				},
			},
			{
				Order: "id",
				Expected: func(a *assertions.Assertion, got []*is.MemberByID) {
					for i, elem := range got {
						a.So(elem.Ids, should.Resemble, memberIDs[i])
						a.So(elem.Rights, should.Resemble, &ttnpb.Rights{
							Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
						})
					}
				},
			},
			{
				Order: "-id",
				Expected: func(a *assertions.Assertion, got []*is.MemberByID) {
					for i, elem := range got {
						a.So(len(memberIDs), should.Equal, len(got))
						a.So(elem.Ids, should.Resemble, memberIDs[len(memberIDs)-1-i])
						a.So(elem.Rights, should.Resemble, &ttnpb.Rights{
							Rights: []ttnpb.Right{ttnpb.Right_RIGHT_APPLICATION_ALL},
						})
					}
				},
			},
		} {
			orderCtx := store.WithOrder(ctx, tc.Order)
			got, err := s.FindMembers(orderCtx, apps[0].GetEntityIdentifiers())
			a.So(err, should.BeNil)
			tc.Expected(a, got)
		}
	})

	t.Run("FindMemberships_Paginated", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		for _, page := range []uint32{1, 2, 3, 4} {
			paginateCtx := store.WithPagination(ctx, 2, page, &total)

			got, err := s.FindMemberships(paginateCtx, memberIDs[0], "application", false)
			if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
				if page == 4 {
					a.So(got, should.HaveLength, 1)
				} else {
					a.So(got, should.HaveLength, 2)
				}

				for i, ids := range got {
					a.So(ids.GetApplicationIds(), should.Resemble, apps[i+2*int(page-1)].GetIds())
				}
			}

			a.So(total, should.Equal, 7)
		}
	})
}

// TestMembershipStorePaginationDefaults tests the default pagination values.
func (st *StoreTest) TestMembershipStorePaginationDefaults(t *T) {
	store.SetPaginationDefaults(store.PaginationDefaults{
		DefaultLimit: 7,
	})

	var apps []*ttnpb.Application
	for i := 0; i < 15; i++ {
		apps = append(apps, st.population.NewApplication(nil))
	}

	var memberIDs []*ttnpb.OrganizationOrUserIdentifiers
	for i := 0; i < 15; i++ {
		ids := st.population.NewUser().GetOrganizationOrUserIdentifiers()
		memberIDs = append(memberIDs, ids)
		st.population.NewMembership(ids, apps[0].GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
	}

	for i := 1; i < 15; i++ {
		st.population.NewMembership(memberIDs[0], apps[i].GetEntityIdentifiers(), ttnpb.Right_RIGHT_APPLICATION_ALL)
	}

	s, ok := st.PrepareDB(t).(interface {
		Store

		is.MembershipStore
	})
	defer st.DestroyDB(t, false)
	if !ok {
		t.Skip("Store does not implement MembershipStore")
	}
	defer s.Close()

	t.Run("FindMembers_PageLimit", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		paginateCtx := store.WithPagination(ctx, 0, 0, &total)
		got, err := s.FindMembers(paginateCtx, apps[0].GetEntityIdentifiers())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.HaveLength, 7)
		}

		paginateCtx = store.WithPagination(ctx, 0, 2, &total)
		got, err = s.FindMembers(paginateCtx, apps[0].GetEntityIdentifiers())
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.HaveLength, 7)
		}
	})

	t.Run("FindMemberships_PageLimit", func(t *T) {
		a, ctx := test.New(t)

		var total uint64
		paginateCtx := store.WithPagination(ctx, 0, 0, &total)
		got, err := s.FindMemberships(paginateCtx, memberIDs[0], "application", false)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.HaveLength, 7)
		}

		paginateCtx = store.WithPagination(ctx, 0, 2, &total)
		got, err = s.FindMemberships(paginateCtx, memberIDs[0], "application", false)
		if a.So(err, should.BeNil) && a.So(got, should.NotBeNil) {
			a.So(got, should.HaveLength, 7)
		}
	})
}
