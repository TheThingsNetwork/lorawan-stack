// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func TestBands(t *testing.T) {
	a := assertions.New(t)

	for _, band := range band.All {
		for _, version := range band.Versions() {
			_, err := band.Version(version)
			if !a.So(err, should.BeNil) {
				t.Logf("Band %s does not support intended LoRaWAN version", band.ID)
			}
		}
	}
}

func TestUnsupportedBand(t *testing.T) {
	a := assertions.New(t)

	b, err := band.GetByID(band.IN_865_867)
	a.So(err, should.BeNil)

	_, err = b.Version(band.RegionalParameters1_0)
	if !a.So(err, should.NotBeNil) {
		t.Log("LoRaWAN 1.0 is not supported for the Indian band")
	}
}
