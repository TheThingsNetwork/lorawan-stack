// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package ttnpb_test

import (
	"regexp"
	testing "testing"

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
