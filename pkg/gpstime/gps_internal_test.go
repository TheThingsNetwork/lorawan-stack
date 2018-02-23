// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package gpstime

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestIsGPSLeap(t *testing.T) {
	a := assertions.New(t)
	for _, v := range leaps {
		a.So(IsLeap(v), should.BeTrue)
		a.So(IsLeap(v+1), should.BeFalse)
	}
}
