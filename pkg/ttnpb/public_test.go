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
	"go.thethings.network/lorawan-stack/v3/pkg/util/test"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestApplicationPublicSafe(t *testing.T) {
	a := assertions.New(t)

	a.So(((*Application)(nil)).PublicSafe(), should.BeNil)

	src := NewPopulatedApplication(test.Randy, true)
	safe := src.PublicSafe()

	a.So(safe, should.NotResemble, src)
	a.So(safe.Name, should.BeEmpty)
}

func TestClientPublicSafe(t *testing.T) {
	a := assertions.New(t)

	a.So(((*Client)(nil)).PublicSafe(), should.BeNil)

	src := NewPopulatedClient(test.Randy, true)
	safe := src.PublicSafe()

	a.So(safe, should.NotResemble, src)
	a.So(safe.Name, should.Equal, src.Name)
}

func TestGatewayPublicSafe(t *testing.T) {
	a := assertions.New(t)

	a.So(((*Gateway)(nil)).PublicSafe(), should.BeNil)

	src := NewPopulatedGateway(test.Randy, true)
	safe := src.PublicSafe()

	a.So(safe, should.NotResemble, src)
	a.So(safe.Name, should.Equal, src.Name)
}

func TestOrganizationPublicSafe(t *testing.T) {
	a := assertions.New(t)

	a.So(((*Organization)(nil)).PublicSafe(), should.BeNil)

	src := NewPopulatedOrganization(test.Randy, true)
	safe := src.PublicSafe()

	a.So(safe, should.NotResemble, src)
	a.So(safe.Name, should.Equal, src.Name)
}

func TestUserPublicSafe(t *testing.T) {
	a := assertions.New(t)

	a.So(((*User)(nil)).PublicSafe(), should.BeNil)

	src := NewPopulatedUser(test.Randy, true)
	safe := src.PublicSafe()

	a.So(safe, should.NotResemble, src)
	a.So(safe.Name, should.Equal, src.Name)
}
