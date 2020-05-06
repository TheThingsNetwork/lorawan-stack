// Copyright © 2020 The Things Network Foundation, The Things Industries B.V.
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

package udp

import (
	"testing"
	"time"

	"github.com/smartystreets/assertions"
	"go.thethings.network/lorawan-stack/v3/pkg/util/test/assertions/should"
)

func TestTimestamps(t *testing.T) {
	a := assertions.New(t)

	timestamps := newTimestamps(4)

	ret := timestamps.Append(time.Now())
	a.So(ret, should.BeZeroValue)
	for i := 0; i < 3; i++ {
		ret := timestamps.Append(time.Now().Add(time.Hour))
		a.So(ret, should.BeZeroValue)
	}

	val := timestamps.Append(time.Now())
	a.So(val.Before(time.Now()), should.BeTrue)

	val = timestamps.Append(time.Now())
	a.So(val.After(time.Now()), should.BeTrue)
}
