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

package sql_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/identityserver/test"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
)

func TestSettings(t *testing.T) {
	a := assertions.New(t)
	s := testStore(t, database)

	settings := ttnpb.IdentityServerSettings{
		BlacklistedIDs: []string{"a"},
		AllowedEmails:  []string{},
	}
	a.So(s.Settings.Set(settings), should.BeNil)

	found, err := s.Settings.Get()
	a.So(err, should.BeNil)
	a.So(found, test.ShouldBeSettingsIgnoringAutoFields, settings)
}
