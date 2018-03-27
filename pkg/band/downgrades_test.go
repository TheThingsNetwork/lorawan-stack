// Copyright © 2018 The Things Network Foundation, distributed under the MIT license (see LICENSE file)

package band_test

import (
	"testing"

	"github.com/TheThingsNetwork/ttn/pkg/band"
	"github.com/smartystreets/assertions"
	"github.com/smartystreets/assertions/should"
)

func compatibleVerifier(t *testing.T) func(version band.RegionalParametersVersion, versionName string, bandIDs ...string) {
	return func(version band.RegionalParametersVersion, versionName string, bandIDs ...string) {
		for _, id := range bandIDs {
			b, err := band.GetByID(id)
			if err != nil {
				t.Fatalf("Could not retrieve band %s: %s\n", id, err)
			}

			if _, err = b.Version(version); err != nil {
				t.Fatalf("Band %s does not support intended LoRaWAN Regional Parameters version %s\n", b.ID, versionName)
			}
		}
	}
}

func TestBands(t *testing.T) {
	verifyCompatibility := compatibleVerifier(t)

	bands := []band.ID{band.EU_863_870, band.US_902_928, band.CN_779_787, band.EU_443}
	verifyCompatibility(band.RegionalParameters1_0, "1.0", bands...)

	bands = append(bands, band.AU_915_928, band.CN_470_510)
	verifyCompatibility(band.RegionalParameters1_0_1, "1.0.1", bands...)

	bands = append(bands, band.AS_923, band.KR_920_923, band.IN_865_867)
	verifyCompatibility(band.RegionalParameters1_0_2, "1.0.2", bands...)
	verifyCompatibility(band.RegionalParameters1_1A, "1.1A", bands...)

	bands = append(bands, band.RU_864_870)
	verifyCompatibility(band.CurrentVersion, "1.1B", bands...)
}

func TestUnsupportedBand(t *testing.T) {
	a := assertions.New(t)

	b, err := band.GetByID(band.IN_865_867)
	a.So(err, should.BeNil)

	_, err = b.Version(band.RegionalParameters1_0)
	if !a.So(err, should.NotBeNil) {
		t.Log("LoRaWAN Regional Parameters 1.0 is not supported for the Indian band")
	}
}
