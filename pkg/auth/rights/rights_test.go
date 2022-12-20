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

package rights

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/unique"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestContext(t *testing.T) {
	a := assertions.New(t)

	ctx := test.Context()
	rights, ok := fromContext(ctx)
	a.So(ok, should.BeFalse)
	a.So(rights, should.Resemble, &Rights{})

	fooRights := &Rights{
		ApplicationRights: *NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, &ttnpb.ApplicationIdentifiers{
				ApplicationId: "foo-app",
			}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
		}),
		ClientRights: *NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, &ttnpb.ClientIdentifiers{
				ClientId: "foo-cli",
			}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_CLIENT_INFO),
		}),
		GatewayRights: *NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, &ttnpb.GatewayIdentifiers{
				GatewayId: "foo-gtw",
			}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_GATEWAY_INFO),
		}),
		OrganizationRights: *NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, &ttnpb.OrganizationIdentifiers{
				OrganizationId: "foo-org",
			}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_ORGANIZATION_INFO),
		}),
		UserRights: *NewMap(map[string]*ttnpb.Rights{
			unique.ID(ctx, &ttnpb.UserIdentifiers{
				UserId: "foo-usr",
			}): ttnpb.RightsFrom(ttnpb.Right_RIGHT_USER_INFO),
		}),
	}

	ctx = NewContext(ctx, fooRights)

	rights, ok = fromContext(ctx)
	a.So(ok, should.BeTrue)
	a.So(rights, should.Resemble, fooRights)
	a.So(rights.IncludesApplicationRights(
		unique.ID(ctx, &ttnpb.ApplicationIdentifiers{ApplicationId: "foo-app"}),
		ttnpb.Right_RIGHT_APPLICATION_INFO,
	), should.BeTrue)
	a.So(rights.IncludesClientRights(
		unique.ID(ctx, &ttnpb.ClientIdentifiers{ClientId: "foo-cli"}),
		ttnpb.Right_RIGHT_CLIENT_INFO,
	), should.BeTrue)
	a.So(rights.IncludesGatewayRights(
		unique.ID(ctx, &ttnpb.GatewayIdentifiers{GatewayId: "foo-gtw"}),
		ttnpb.Right_RIGHT_GATEWAY_INFO,
	), should.BeTrue)
	a.So(rights.IncludesOrganizationRights(
		unique.ID(ctx, &ttnpb.OrganizationIdentifiers{OrganizationId: "foo-org"}),
		ttnpb.Right_RIGHT_ORGANIZATION_INFO,
	), should.BeTrue)
	a.So(rights.IncludesUserRights(
		unique.ID(ctx, &ttnpb.UserIdentifiers{UserId: "foo-usr"}),
		ttnpb.Right_RIGHT_USER_INFO,
	), should.BeTrue)
}

func TestMap(t *testing.T) {
	t.Parallel()
	a := assertions.New(t)

	expectedMap := func(rights map[string]*ttnpb.Rights) *Map {
		m := &Map{}
		for k, v := range rights {
			m.syncMap.Store(k, v)
		}
		return m
	}

	var (
		fooID   = "foo"
		barID   = "bar"
		wrongID = "baz"
	)
	for _, tc := range []struct {
		Name      string
		Rights    map[string]*ttnpb.Rights
		NewMap    func(*testing.T, map[string]*ttnpb.Rights) *Map
		Assertion func(*testing.T, map[string]*ttnpb.Rights, *Map) bool
	}{
		{
			Name:   "`NewMap` with nil Rights",
			Rights: nil,
			NewMap: func(t *testing.T, r map[string]*ttnpb.Rights) *Map {
				t.Helper()
				return NewMap(r)
			},
			Assertion: func(t *testing.T, _ map[string]*ttnpb.Rights, m *Map) bool {
				t.Helper()
				return a.So(m, should.Resemble, &Map{})
			},
		},

		{
			Name: "NewMap with argument",
			Rights: map[string]*ttnpb.Rights{
				fooID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
				barID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
			},
			NewMap: func(t *testing.T, r map[string]*ttnpb.Rights) *Map {
				t.Helper()
				return NewMap(r)
			},
			Assertion: func(t *testing.T, r map[string]*ttnpb.Rights, m *Map) bool {
				t.Helper()
				for k, v := range r {
					got, ok := m.GetRights(k)
					if !a.So(ok, should.BeTrue) {
						return false
					}
					if !a.So(got, should.Resemble, v) {
						return false
					}
				}
				return true
			},
		},

		{
			Name: "GetRights correct ID",
			Rights: map[string]*ttnpb.Rights{
				fooID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
				barID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
			},
			NewMap: func(t *testing.T, r map[string]*ttnpb.Rights) *Map {
				t.Helper()
				return expectedMap(r)
			},
			Assertion: func(t *testing.T, r map[string]*ttnpb.Rights, m *Map) bool {
				t.Helper()
				for k, v := range r {
					got, ok := m.GetRights(k)
					if !a.So(ok, should.BeTrue) {
						return false
					}
					if !a.So(got, should.Resemble, v) {
						return false
					}
				}
				return true
			},
		},

		{
			Name: "SetRights with key and value",
			Rights: map[string]*ttnpb.Rights{
				fooID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
				barID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_DEVICES_READ),
			},
			NewMap: func(t *testing.T, r map[string]*ttnpb.Rights) *Map {
				t.Helper()
				m := &Map{}
				for k, v := range r {
					m.SetRights(k, v)
				}
				return m
			},
			Assertion: func(t *testing.T, r map[string]*ttnpb.Rights, m *Map) bool {
				t.Helper()
				for k, v := range r {
					got, ok := m.GetRights(k)
					if !a.So(ok, should.BeTrue) {
						return false
					}
					if !a.So(got, should.Resemble, v) {
						return false
					}
				}
				return true
			},
		},

		{
			Name: "GetRights wrong ID",
			Rights: map[string]*ttnpb.Rights{
				fooID: ttnpb.RightsFrom(ttnpb.Right_RIGHT_APPLICATION_INFO),
			},
			NewMap: func(t *testing.T, r map[string]*ttnpb.Rights) *Map {
				t.Helper()
				return expectedMap(r)
			},
			Assertion: func(t *testing.T, _ map[string]*ttnpb.Rights, m *Map) bool {
				t.Helper()
				got, ok := m.GetRights(wrongID)
				return a.So(got, should.BeNil) && a.So(ok, should.BeFalse)
			},
		},
	} {
		tc := tc
		t.Run(tc.Name, func(t *testing.T) {
			t.Parallel()
			a.So(tc.Assertion(t, tc.Rights, tc.NewMap(t, tc.Rights)), should.BeTrue)
		})
	}
}
