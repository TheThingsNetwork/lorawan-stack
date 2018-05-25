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

package ttnpb

import (
	"testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestIsIDAllowed(t *testing.T) {
	a := assertions.New(t)

	settings := new(IdentityServerSettings)

	// all ids are allowed
	settings.BlacklistedIDs = nil
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeTrue)
	settings.BlacklistedIDs = []string{}
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeTrue)

	// `admin` is blacklisted
	settings.BlacklistedIDs = []string{"admin"}
	a.So(settings.IsIDAllowed("foobar"), should.BeTrue)
	a.So(settings.IsIDAllowed("admin"), should.BeFalse)
}
