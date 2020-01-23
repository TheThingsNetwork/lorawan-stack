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

package gpstime

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/pkg/util/test/assertions/should"
)

func TestIsLeapSecond(t *testing.T) {
	a := assertions.New(t)
	for _, v := range leaps {
		a.So(IsLeapSecond(time.Duration(v)*time.Second-time.Millisecond), should.BeFalse)
		a.So(IsLeapSecond(time.Duration(v)*time.Second-time.Microsecond), should.BeFalse)
		a.So(IsLeapSecond(time.Duration(v)*time.Second-time.Nanosecond), should.BeFalse)
		a.So(IsLeapSecond(time.Duration(v)*time.Second), should.BeTrue)
		a.So(IsLeapSecond(time.Duration(v)*time.Second+time.Nanosecond), should.BeTrue)
		a.So(IsLeapSecond(time.Duration(v)*time.Second+time.Microsecond), should.BeTrue)
		a.So(IsLeapSecond(time.Duration(v)*time.Second+999*time.Millisecond), should.BeTrue)
		a.So(IsLeapSecond(time.Duration(v)*time.Second+time.Second), should.BeFalse)
	}
}
