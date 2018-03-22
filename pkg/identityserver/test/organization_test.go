// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package test

import (
	"testing"
	"time"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func organization() *ttnpb.Organization {
	return &ttnpb.Organization{
		OrganizationIdentifiers: ttnpb.OrganizationIdentifiers{OrganizationID: "foo"},
		Name:        "Foo Bar",
		Description: "foo",
		URL:         "http://foo.bar",
		Location:    "Baz",
		Email:       "foo@bar.baz",
		CreatedAt:   time.Now().UTC(),
		UpdatedAt:   time.Now().UTC().Add(time.Hour),
	}
}

func TestShouldBeOrganization(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeOrganization(organization(), organization()), should.Equal, success)

	modified := organization()
	modified.CreatedAt = time.Now().UTC().Add(time.Minute)

	a.So(ShouldBeOrganization(modified, organization()), should.NotEqual, success)
}

func TestShouldBeOrganizationIgnoringAutoFields(t *testing.T) {
	a := assertions.New(t)

	a.So(ShouldBeOrganizationIgnoringAutoFields(organization(), organization()), should.Equal, success)

	modified := organization()
	modified.Description = "lol"

	a.So(ShouldBeOrganizationIgnoringAutoFields(modified, organization()), should.NotEqual, success)
}
