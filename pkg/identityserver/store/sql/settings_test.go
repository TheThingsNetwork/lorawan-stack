// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package sql

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestSettings(t *testing.T) {
	a := assertions.New(t)
	s := cleanStore(t, database)

	settings := &ttnpb.IdentityServerSettings{
		BlacklistedIDs: []string{"a"},
		AllowedEmails:  []string{},
	}

	a.So(s.Settings.Set(settings), should.BeNil)

	found, err := s.Settings.Get()
	a.So(err, should.BeNil)
	a.So(found, should.Resemble, settings)
}
