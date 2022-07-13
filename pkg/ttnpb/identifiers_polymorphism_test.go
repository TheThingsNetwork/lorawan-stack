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

package ttnpb_test

import (
	"testing"

	"github.com/smartystreets/assertions"
	. "go.thethings.network/lorawan-stack/v3/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestEntityType(t *testing.T) {
	a := assertions.New(t)

	applicationID := ApplicationIdentifiers{ApplicationId: "foo"}
	clientID := ClientIdentifiers{ClientId: "foo"}
	endDeviceID := EndDeviceIdentifiers{DeviceId: "foo", ApplicationIds: &applicationID}
	gatewayID := GatewayIdentifiers{GatewayId: "foo"}
	organizationID := OrganizationIdentifiers{OrganizationId: "foo"}
	userID := UserIdentifiers{UserId: "foo"}

	a.So(applicationID.EntityType(), should.Equal, "application")
	a.So(applicationID.IDString(), should.Equal, "foo")
	a.So(applicationID.GetEntityIdentifiers().EntityType(), should.Equal, "application")
	a.So(applicationID.GetEntityIdentifiers().IDString(), should.Equal, "foo")

	a.So(clientID.EntityType(), should.Equal, "client")
	a.So(clientID.IDString(), should.Equal, "foo")
	a.So(clientID.GetEntityIdentifiers().EntityType(), should.Equal, "client")
	a.So(clientID.GetEntityIdentifiers().IDString(), should.Equal, "foo")

	a.So(endDeviceID.EntityType(), should.Equal, "end device")
	a.So(endDeviceID.IDString(), should.Equal, "foo.foo")
	a.So(endDeviceID.GetEntityIdentifiers().EntityType(), should.Equal, "end device")
	a.So(endDeviceID.GetEntityIdentifiers().IDString(), should.Equal, "foo.foo")

	a.So(gatewayID.EntityType(), should.Equal, "gateway")
	a.So(gatewayID.IDString(), should.Equal, "foo")
	a.So(gatewayID.GetEntityIdentifiers().EntityType(), should.Equal, "gateway")
	a.So(gatewayID.GetEntityIdentifiers().IDString(), should.Equal, "foo")

	a.So(organizationID.EntityType(), should.Equal, "organization")
	a.So(organizationID.IDString(), should.Equal, "foo")
	a.So(organizationID.GetEntityIdentifiers().EntityType(), should.Equal, "organization")
	a.So(organizationID.GetEntityIdentifiers().IDString(), should.Equal, "foo")
	a.So(organizationID.GetOrganizationOrUserIdentifiers().EntityType(), should.Equal, "organization")
	a.So(organizationID.GetOrganizationOrUserIdentifiers().IDString(), should.Equal, "foo")

	a.So(userID.EntityType(), should.Equal, "user")
	a.So(userID.IDString(), should.Equal, "foo")
	a.So(userID.GetEntityIdentifiers().EntityType(), should.Equal, "user")
	a.So(userID.GetEntityIdentifiers().IDString(), should.Equal, "foo")
	a.So(userID.GetOrganizationOrUserIdentifiers().EntityType(), should.Equal, "user")
	a.So(userID.GetOrganizationOrUserIdentifiers().IDString(), should.Equal, "foo")
}
