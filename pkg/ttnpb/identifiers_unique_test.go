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
	"context"
	testing "testing"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestUserIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := UserIdentifiers{
		UserID: "foo",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo")
}

func TestApplicationIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := ApplicationIdentifiers{
		ApplicationID: "foo",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo")
}

func TestGatewayIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := GatewayIdentifiers{
		GatewayID: "foo",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo")
}

func TestEndDeviceIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := EndDeviceIdentifiers{
		ApplicationIdentifiers: ApplicationIdentifiers{
			ApplicationID: "foo",
		},
		DeviceID: "bar",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo:bar")
}

func TestClientIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := ClientIdentifiers{
		ClientID: "foo",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo")
}

func TestOrganizationIdentifiersUniqueID(t *testing.T) {
	a := assertions.New(t)

	ctx := context.Background()
	ids := OrganizationIdentifiers{
		OrganizationID: "foo",
	}
	a.So(ids.UniqueID(ctx), should.Equal, "foo")
}
