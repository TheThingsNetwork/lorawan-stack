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

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func organization() *ttnpb.Organization {
	now := time.Now()

	return &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
		Name:        "Foo Bar",
		Description: "foo",
		URL:         "http://foo.bar",
		Location:    "Baz",
		Email:       "foo@bar.baz",
		CreatedAt:   now,
		UpdatedAt:   now.Add(time.Hour),
	}
}

func TestShouldBeOrganization(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeOrganization(organization(), organization()), should.Equal, success)

	modified := organization()
	modified.CreatedAt = time.Now().Add(time.Minute)

	a.So(ShouldBeOrganization(modified, organization()), should.NotEqual, success)
}

func TestShouldBeOrganizationIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeOrganizationIgnoringAutoFields(organization(), organization()), should.Equal, success)

	modified := organization()
	modified.Description = "lol"

	a.So(ShouldBeOrganizationIgnoringAutoFields(modified, organization()), should.NotEqual, success)
}
