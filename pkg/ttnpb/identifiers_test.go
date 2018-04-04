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

package ttnpb_test

import (
	"regexp"
	"testing"

	. "github.com/TheThingsNetwork/ttn/pkg/ttnpb"
	"github.com/TheThingsNetwork/ttn/pkg/util/test"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

var idRegexp = regexp.MustCompile("^[0-9a-z](?:[_-]?[0-9a-z]){1,35}$")

func TestNewPopulatedEndDeviceIdentifiers(t *testing.T) {
	id := NewPopulatedEndDeviceIdentifiers(test.Randy, false)
	assertions.New(t).So(id.DeviceID == "" || idRegexp.MatchString(id.DeviceID), should.BeTrue)
	assertions.New(t).So(id.ApplicationID == "" || idRegexp.MatchString(id.ApplicationID), should.BeTrue)
}
