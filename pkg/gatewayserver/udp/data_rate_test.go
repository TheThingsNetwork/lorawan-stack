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

package udp

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
	"go.thethings.network/lorawan-stack/pkg/types"
)

func TestDataRate(t *testing.T) {
	a := assertions.New(t)

	table := map[string]DataRate{
		`"SF7BW125"`: {types.DataRate{LoRa: "SF7BW125"}},
		`50000`:      {types.DataRate{FSK: 50000}},
	}

	for s, dr := range table {
		enc, err := dr.MarshalJSON()
		a.So(err, should.BeNil)
		a.So(string(enc), should.Equal, s)

		var dec DataRate
		err = dec.UnmarshalJSON(enc)
		a.So(err, should.BeNil)
		a.So(dec, should.Resemble, dr)
	}

	var dr DataRate
	err := dr.UnmarshalJSON([]byte{})
	a.So(err, should.NotBeNil)
}
