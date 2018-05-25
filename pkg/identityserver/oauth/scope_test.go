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

package oauth_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/pkg/identityserver/oauth"
	"go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestParseScope(t *testing.T) {
	a := assertions.New(t)

	// Valid.
	{
		rights, err := ParseScope("RIGHT_APPLICATION_INFO RIGHT_APPLICATION_TRAFFIC_READ")
		a.So(err, should.BeNil)
		a.So(rights, should.Resemble, []ttnpb.Right{
			ttnpb.RIGHT_APPLICATION_INFO,
			ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
		})
	}

	// Invalid.
	{
		rights, err := ParseScope("RIGHT_APPLICATION_TRAFFIC_READ RIGHT_WEIRD")
		a.So(err, should.NotBeNil)
		a.So(rights, should.BeNil)
	}
}

func TestStringScope(t *testing.T) {
	a := assertions.New(t)

	rights := []ttnpb.Right{
		ttnpb.RIGHT_APPLICATION_INFO,
		ttnpb.RIGHT_APPLICATION_TRAFFIC_READ,
	}

	a.So(Scope(rights), should.Equal, "RIGHT_APPLICATION_INFO RIGHT_APPLICATION_TRAFFIC_READ")
}
