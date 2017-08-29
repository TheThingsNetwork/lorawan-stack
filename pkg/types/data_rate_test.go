// Copyright © 2017 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package types

import (
	"testing"

	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestDataRate(t *testing.T) {
	a := assertions.New(t)

	table := map[string]DataRate{
		`"SF7BW125"`: DataRate{LoRa: "SF7BW125"},
		`50000`:      DataRate{FSK: 50000},
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
