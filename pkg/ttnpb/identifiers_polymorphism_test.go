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
	. "go.thethings.network/lorawan-stack/pkg/ttnpb"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestEntityType(t *testing.T) {
	a := assertions.New(t)

	applicationID := ApplicationIdentifiers{ApplicationID: "foo"}
	clientID := ClientIdentifiers{ClientID: "foo"}
	endDeviceID := EndDeviceIdentifiers{DeviceID: "foo", ApplicationIdentifiers: applicationID}
	gatewayID := GatewayIdentifiers{GatewayID: "foo"}
	organizationID := OrganizationIdentifiers{OrganizationID: "foo"}
	userID := UserIdentifiers{UserID: "foo"}

	a.So(applicationID.EntityType(), should.Equal, "application")
	a.So(applicationID.IDString(), should.Equal, "foo")

	a.So(clientID.EntityType(), should.Equal, "client")
	a.So(clientID.IDString(), should.Equal, "foo")

	a.So(endDeviceID.EntityType(), should.Equal, "end device")
	a.So(endDeviceID.IDString(), should.Equal, "foo.foo")

	a.So(gatewayID.EntityType(), should.Equal, "gateway")
	a.So(gatewayID.IDString(), should.Equal, "foo")

	a.So(organizationID.EntityType(), should.Equal, "organization")
	a.So(organizationID.IDString(), should.Equal, "foo")

	a.So(userID.EntityType(), should.Equal, "user")
	a.So(userID.IDString(), should.Equal, "foo")

	for _, id := range []Identifiers{
		&applicationID,
		&clientID,
		&endDeviceID,
		&gatewayID,
		&organizationID,
		&userID,
	} {
		a.So(id.Identifiers(), should.Resemble, id)
		a.So(id.EntityIdentifiers().Identifiers(), should.Resemble, id)

		a.So(id.Identifiers().EntityType(), should.Equal, id.EntityType())
		a.So(id.EntityIdentifiers().EntityType(), should.Equal, id.EntityType())
		a.So(id.EntityIdentifiers().Identifiers().EntityType(), should.Equal, id.EntityType())

		a.So(id.Identifiers().IDString(), should.Equal, id.IDString())
		a.So(id.EntityIdentifiers().IDString(), should.Equal, id.IDString())
		a.So(id.EntityIdentifiers().Identifiers().IDString(), should.Equal, id.IDString())

		if orgOrUsr, ok := id.(interface {
			OrganizationOrUserIdentifiers() *OrganizationOrUserIdentifiers
		}); ok {
			ouid := orgOrUsr.OrganizationOrUserIdentifiers()
			a.So(ouid.EntityType(), should.Equal, id.EntityType())
			a.So(ouid.IDString(), should.Equal, id.IDString())

			a.So(ouid.Identifiers(), should.Resemble, id)
			a.So(ouid.EntityIdentifiers().Identifiers(), should.Resemble, id)

			a.So(ouid.Identifiers().EntityType(), should.Equal, id.EntityType())
			a.So(ouid.EntityIdentifiers().EntityType(), should.Equal, id.EntityType())
			a.So(ouid.EntityIdentifiers().Identifiers().EntityType(), should.Equal, id.EntityType())

			a.So(ouid.Identifiers().IDString(), should.Equal, id.IDString())
			a.So(ouid.EntityIdentifiers().IDString(), should.Equal, id.IDString())
			a.So(ouid.EntityIdentifiers().Identifiers().IDString(), should.Equal, id.IDString())
		}
	}
}
